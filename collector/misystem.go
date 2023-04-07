package collector

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/helloworlde/miwifi-exporter/config"
)

type Status struct {
	Dev []struct {
		Mac              string      `json:"mac"`
		MaxDownloadSpeed string      `json:"maxdownloadspeed"`
		Upload           interface{} `json:"upload"`
		UpSpeed          interface{} `json:"upspeed"`
		DownSpeed        interface{} `json:"downspeed"`
		Online           string      `json:"online"`
		DevName          string      `json:"devname"`
		MaxUploadSpeed   string      `json:"maxuploadspeed"`
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
		All               int `json:"all"`
		Online            int `json:"online"`
		AllWithoutMash    int `json:"all_without_mash"`
		OnlineWithoutMash int `json:"online_without_mash"`
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
		DownSpeed        string `json:"downspeed"`
		MaxDownloadSpeed string `json:"maxdownloadspeed"`
		History          string `json:"history"`
		DevName          string `json:"devname"`
		Upload           string `json:"upload"`
		UpSpeed          string `json:"upspeed"`
		MaxUploadSpeed   string `json:"maxuploadspeed"`
		Download         string `json:"download"`
	} `json:"wan"`
}

type DeviceList struct {
	Mac  string `json:"mac"`
	List []struct {
		Mac       string `json:"mac"`
		OnNme     string `json:"oname"`
		IsAP      int    `json:"isap"`
		Parent    string `json:"parent"`
		Authority struct {
			Wan     int `json:"wan"`
			PriDisk int `json:"pridisk"`
			Admin   int `json:"admin"`
			Lan     int `json:"lan"`
		} `json:"authority"`
		Push   int    `json:"push"`
		Online int    `json:"online"`
		Name   string `json:"name"`
		Times  int    `json:"times"`
		IP     []struct {
			DownSpeed string `json:"downspeed"`
			Online    string `json:"online"`
			Active    int    `json:"active"`
			UpSpeed   string `json:"upspeed"`
			IP        string `json:"ip"`
		} `json:"ip"`
		Statistics struct {
			DownSpeed string `json:"downspeed"`
			Online    string `json:"online"`
			UpSpeed   string `json:"upspeed"`
		} `json:"statistics"`
		Icon string `json:"icon"`
		Type int    `json:"type"`
	} `json:"list"`
	Code int `json:"code"`
}

var StatusRepo Status
var DeviceListRepo DeviceList

func GetMiSystemDeviceList() {
	client := http.Client{}
	res, err := client.Get(fmt.Sprintf("http://%s/cgi-bin/luci/;stok=%s/api/misystem/devicelist",
		config.Configs.IP, config.Token.Token))
	if err != nil {
		log.Println("请求路由器错误，可能原因：1.路由器掉线或者宕机", err)
		os.Exit(1)
	}
	body, err := ioutil.ReadAll(res.Body)
	count := 0
	if err = json.Unmarshal(body, &DeviceListRepo); err != nil {
		log.Println("Token失效，正在重试获取")
		count++
		time.Sleep(1 * time.Minute)
		if count >= 5 {
			log.Println("获取状态错误，可能原因：1.账号或者密码错误，2.路由器鉴权错误", err)
			os.Exit(1)
		}
	}

}

func GetMiSystemStatus() {
	client := http.Client{}
	res, err := client.Get(fmt.Sprintf("http://%s/cgi-bin/luci/;stok=%s/api/misystem/status",
		config.Configs.IP, config.Token.Token))

	if err != nil {
		log.Println("请求路由器错误，可能原因：1.路由器掉线或者宕机", err)
		os.Exit(1)
	}

	body, err := ioutil.ReadAll(res.Body)
	count := 0
	if err = json.Unmarshal(body, &StatusRepo); err != nil {
		log.Println("Token失效，正在重试获取", err)
		config.GetConfig()
		GetMiSystemStatus()
		count++
		time.Sleep(1 * time.Minute)
		if count >= 5 {
			log.Println("获取状态错误，可能原因：1.账号或者密码错误，2.路由器鉴权错误", err)
			os.Exit(1)
		}
	}
}

func (r *Status) GetRouterUptime() float64 {
	var n float64

	n, err := strconv.ParseFloat(r.UpTime, 64)
	if err != nil {
		log.Println("err: ", err)
	}

	return n
}

func (r *Status) GetRouterCPUMhz() float64 {
	var n float64

	switch {
	case strings.HasSuffix(StatusRepo.CPU.Hz, "GHz"):
		x, err := strconv.ParseFloat(strings.Split(StatusRepo.CPU.Hz, "GHz")[0], 64)
		if err != nil {
			log.Println("err: ", err)
		}
		n = x * 1000
	case strings.HasSuffix(StatusRepo.CPU.Hz, "MHz"):
		x, err := strconv.ParseFloat(strings.Split(StatusRepo.CPU.Hz, "MHz")[0], 64)
		if err != nil {
			log.Println("err: ", err)
		}
		n = x
	}

	return n
}

func (r *Status) GetRouterMemoryTotal() float64 {
	var n float64

	if len(strings.Split(StatusRepo.Mem.Total, "MB")) < 1 {
		return n
	}

	n, err := strconv.ParseFloat(strings.Split(StatusRepo.Mem.Total, "MB")[0], 64)
	if err != nil {
		log.Println("err: ", err)
	}

	return n
}

func (r *Status) GetRouterUpSpeed() float64 {
	var n float64

	n, err := strconv.ParseFloat(r.Wan.UpSpeed, 64)
	if err != nil {
		log.Println("err: ", err)
	}

	return n
}

func (r *Status) GetRouterDownSpeed() float64 {
	var n float64

	n, err := strconv.ParseFloat(r.Wan.DownSpeed, 64)
	if err != nil {
		log.Println("err: ", err)
	}

	return n
}

func (r *Status) GetRouterUpload() float64 {
	var n float64

	n, err := strconv.ParseFloat(r.Wan.Upload, 64)
	if err != nil {
		log.Println("err: ", err)
	}

	return n
}

func (r *Status) GetRouterDownload() float64 {
	var n float64

	n, err := strconv.ParseFloat(r.Wan.Download, 64)
	if err != nil {
		log.Println("err: ", err)
	}

	return n
}
