package token

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Auth struct {
	URL   string `json:"url"`
	Token string `json:"token"`
	Code  int    `json:"code"`
}

func GetToken(ip, passwd string) Auth {
	log.Println("开始获取Token")
	client := http.Client{}
	res, err := client.Get(fmt.Sprintf("http://%s/cgi-bin/luci/web/home", ip))
	if err != nil {
		log.Println("获取路由器登录页错误，可能原因：1.配置的路由器IP错误", err)
		os.Exit(1)
	}
	body, _ := ioutil.ReadAll(res.Body)
	src := string(body)
	re, err1 := regexp.Compile("key: '(.*)'")
	key := strings.Split(re.FindAllString(src, -1)[0], "'")[1]
	re, err2 := regexp.Compile("deviceId = '(.*)'")
	mac := strings.Split(re.FindAllString(src, -1)[0], "'")[1]
	count := 0
	if err1 != nil || err2 != nil {
		GetToken(ip, passwd)
		count++
		if count >= 5 {
			log.Println("获取key或mac失败，可能原因：路由器固件升级改版", err1, err2)
			os.Exit(1)
		}
	}
	nonce := "0_" + mac + "_" + strconv.Itoa(int(time.Now().Unix())) + "_" + strconv.Itoa(int(rand.Float64()*10000))
	pwd := sha1.New()
	pwd.Write([]byte(passwd + key))
	hexPwd1 := fmt.Sprintf("%x", pwd.Sum(nil))
	pwd2 := sha1.New()
	pwd2.Write([]byte(nonce + hexPwd1))
	hexPwd2 := fmt.Sprintf("%x", pwd2.Sum(nil))
	data := make(url.Values)
	data["logtype"] = []string{"2"}
	data["nonce"] = []string{nonce}
	data["password"] = []string{hexPwd2}
	data["username"] = []string{"admin"}
	res, _ = client.PostForm("http://"+ip+"/cgi-bin/luci/api/xqsystem/login", data)
	body, _ = ioutil.ReadAll(res.Body)
	auth := Auth{}

	if err := json.Unmarshal(body, &auth); err != nil || auth.Code == 401 {
		log.Println("获取认证错误，可能原因：1.账号或者密码错误，2.账号权限不足", err)
		os.Exit(1)
	}
	log.Println("获取Token成功")

	return auth
}
