package token

import (
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"os"
	"regexp"
	"strings"
	"time"
)

type Auth struct {
	URL   string `json:"url"`
	Token string `json:"token"`
	Code  int    `json:"code"`
}

type Router struct {
	IP       string
	Password string
	Headers  map[string]string
	Session  *http.Client
	Data     map[string]string
	Token    string
	Path     string
	Stok     string
}

type InitInfo struct {
	Hardware       string `json:"hardware"`
	RomVersion     string `json:"romversion"`
	SerialNumber   string `json:"id"`
	RouterName     string `json:"routername"`
	NewEncryptMode int    `json:"newEncryptMode"`
}

func hashSHA1(data string) string {
	h := sha1.New()
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

func hashSHA256(data string) string {
	h := sha256.New()
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

func (r *Router) login() error {
	jar, _ := cookiejar.New(nil)
	r.Session = &http.Client{Jar: jar}

	url := fmt.Sprintf("http://%s/cgi-bin/luci/web", r.IP)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	for key, value := range r.Headers {
		req.Header.Set(key, value)
	}

	resp, err := r.Session.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	rData := strings.ReplaceAll(string(body), "\r", "")
	rData = strings.ReplaceAll(rData, "\n", "")
	rData = strings.ReplaceAll(rData, "\t", "")

	keyRegex := regexp.MustCompile(`key:.*?'(.*?)',`)
	key := keyRegex.FindStringSubmatch(rData)[1]

	deviceIDRegex := regexp.MustCompile(`deviceId = '(.*?)';`)
	deviceID := deviceIDRegex.FindStringSubmatch(rData)[1]

	initInfoURL := fmt.Sprintf("http://%s/cgi-bin/luci/api/xqsystem/init_info", r.IP)
	req, err = http.NewRequest("GET", initInfoURL, nil)
	if err != nil {
		return err
	}

	for key, value := range r.Headers {
		req.Header.Set(key, value)
	}

	initInfoResp, err := r.Session.Do(req)
	if err != nil {
		return err
	}
	defer initInfoResp.Body.Close()

	var initInfo InitInfo
	err = json.NewDecoder(initInfoResp.Body).Decode(&initInfo)
	if err != nil {
		return err
	}

	r.Data = map[string]string{
		"hardware":      initInfo.Hardware,
		"rom_version":   initInfo.RomVersion,
		"serial_number": initInfo.SerialNumber,
		"router_name":   initInfo.RouterName,
	}

	pwd := r.Password
	nonce := fmt.Sprintf("0_%s_%d_962", deviceID, time.Now().Unix())

	var passWord string
	if initInfo.NewEncryptMode == 1 {
		a := hashSHA256(pwd + key)
		passWord = hashSHA256(nonce + a)
	} else {
		a := hashSHA1(pwd + key)
		passWord = hashSHA1(nonce + a)
	}

	logURL := fmt.Sprintf("http://%s/cgi-bin/luci/api/xqsystem/login", r.IP)
	data := fmt.Sprintf("username=admin&password=%s&logtype=2&nonce=%s", passWord, nonce)

	req, err = http.NewRequest("POST", logURL, strings.NewReader(data))
	if err != nil {
		return err
	}

	for key, value := range r.Headers {
		req.Header.Set(key, value)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	loginResp, err := r.Session.Do(req)
	if err != nil {
		return err
	}
	defer loginResp.Body.Close()

	cookies := loginResp.Cookies()
	if len(cookies) > 0 {
		r.Headers["Cookies"] = cookies[0].Value
	}

	var loginData map[string]interface{}
	err = json.NewDecoder(loginResp.Body).Decode(&loginData)
	if err != nil {
		return err
	}

	r.Token = loginData["token"].(string)
	r.Path = loginData["url"].(string)

	stokRegex := regexp.MustCompile(`;stok=(.*?)/`)
	r.Stok = stokRegex.FindStringSubmatch(r.Path)[1]

	return nil
}

func GetToken(ip, passwd string) Auth {
	router := Router{
		IP:       ip,
		Password: passwd,
		Headers: map[string]string{
			"Connection": "keep-alive",
			"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.72 Safari/537.36",
		},
	}

	auth := Auth{}

	err := router.login()
	if err != nil {
		log.Fatalf("Error: %v", err)
		os.Exit(1)
	} else {
		log.Println("Login successful!")
		auth.URL = router.Path
		auth.Token = router.Stok
		auth.Code = 200
	}

	return auth
}
