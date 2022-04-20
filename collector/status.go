package collector

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/helloworlde/miwifi-exporter/config"
)

type Status struct {
	Dev []struct {
		Mac              string      `json:"mac"`
		Maxdownloadspeed string      `json:"maxdownloadspeed"`
		Upload           interface{} `json:"upload"`
		Upspeed          interface{} `json:"upspeed"`
		Downspeed        interface{} `json:"downspeed"`
		Online           string      `json:"online"`
		Devname          string      `json:"devname"`
		Maxuploadspeed   string      `json:"maxuploadspeed"`
		Download         interface{} `json:"download"`
	} `json:"dev"`
	Code int `json:"code"`
	Mem  struct {
		Usage float64 `json:"usage"`
		Total string  `json:"total"`
		Hz    string  `json:"hz"`
		Type  string  `json:"type"`
	} `json:"mem"`
	Temperature int `json:"temperature"`
	Count       struct {
		All    int `json:"all"`
		Online int `json:"online"`
	} `json:"count"`
	Hardware struct {
		Mac      string `json:"mac"`
		Platform string `json:"platform"`
		Version  string `json:"version"`
		Channel  string `json:"channel"`
		Sn       string `json:"sn"`
	} `json:"hardware"`
	UpTime string `json:"upTime"`
	CPU    struct {
		Core int     `json:"core"`
		Hz   string  `json:"hz"`
		Load float64 `json:"load"`
	} `json:"cpu"`
	Wan struct {
		Downspeed        string `json:"downspeed"`
		Maxdownloadspeed string `json:"maxdownloadspeed"`
		History          string `json:"history"`
		Devname          string `json:"devname"`
		Upload           string `json:"upload"`
		Upspeed          string `json:"upspeed"`
		Maxuploadspeed   string `json:"maxuploadspeed"`
		Download         string `json:"download"`
	} `json:"wan"`
}

type MACtoIP struct {
	Mac  string `json:"mac"`
	List []struct {
		Mac       string `json:"mac"`
		Oname     string `json:"oname"`
		Isap      int    `json:"isap"`
		Parent    string `json:"parent"`
		Authority struct {
			Wan     int `json:"wan"`
			Pridisk int `json:"pridisk"`
			Admin   int `json:"admin"`
			Lan     int `json:"lan"`
		} `json:"authority"`
		Push   int    `json:"push"`
		Online int    `json:"online"`
		Name   string `json:"name"`
		Times  int    `json:"times"`
		IP     []struct {
			Downspeed string `json:"downspeed"`
			Online    string `json:"online"`
			Active    int    `json:"active"`
			Upspeed   string `json:"upspeed"`
			IP        string `json:"ip"`
		} `json:"ip"`
		Statistics struct {
			Downspeed string `json:"downspeed"`
			Online    string `json:"online"`
			Upspeed   string `json:"upspeed"`
		} `json:"statistics"`
		Icon string `json:"icon"`
		Type int    `json:"type"`
	} `json:"list"`
	Code int `json:"code"`
}

var DevStatus Status
var Mactoip MACtoIP

func GetIPtoMAC() {
	client := http.Client{}
	res, err := client.Get(fmt.Sprintf("http://%s/cgi-bin/luci/;stok=%s/api/misystem/devicelist",
		config.Config.IP, config.Token.Token))
	if err != nil {
		log.Println("请求路由器错误，可能原因：1.路由器掉线或者宕机")
		os.Exit(1)
	}
	body, err := ioutil.ReadAll(res.Body)
	count := 0
	if err = json.Unmarshal([]byte(body), &Mactoip); err != nil {
		log.Println("Token失效，正在重试获取")
		count++
		time.Sleep(1 * time.Minute)
		if count >= 5 {
			log.Println("获取状态错误，可能原因：1.账号或者密码错误，2.路由器鉴权错误")
			os.Exit(1)
		}
	}

}

func GetMiwifiStatus() {
	client := http.Client{}
	res, err := client.Get(fmt.Sprintf("http://%s/cgi-bin/luci/;stok=%s/api/misystem/status",
		config.Config.IP, config.Token.Token))

	if err != nil {
		log.Println("请求路由器错误，可能原因：1.路由器掉线或者宕机")
		os.Exit(1)
	}

	body, err := ioutil.ReadAll(res.Body)
	count := 0
	if err = json.Unmarshal([]byte(body), &DevStatus); err != nil {
		fmt.Println(DevStatus.Dev)
		log.Println("Token失效，正在重试获取")
		config.GetConfig()
		GetMiwifiStatus()
		count++
		time.Sleep(1 * time.Minute)
		if count >= 5 {
			log.Println("获取状态错误，可能原因：1.账号或者密码错误，2.路由器鉴权错误")
			os.Exit(1)
		}
	}
}
