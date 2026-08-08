package collector

import (
	"log"
	"strconv"
	"strings"
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
	if err := fetchAPI("api/misystem/devicelist", &DeviceListRepo); err != nil {
		log.Println("获取设备列表失败：", err)
	}
}

func GetMiSystemStatus() {
	if err := fetchAPI("api/misystem/status", &StatusRepo); err != nil {
		log.Println("获取路由器状态失败：", err)
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

// NewStatus 对应 /api/misystem/newstatus，提供各频段 WiFi 在线终端数与 Mesh 节点数。
// 注意：该接口会返回各频段密码（passwd），监控场景下不采集该字段。
type NewStatus struct {
	BSD      string `json:"bsd"`
	Code     int    `json:"code"`
	Count    int    `json:"count"`
	CapCount int    `json:"cap_count"` // Mesh 主节点数
	ReCount  int    `json:"re_count"`  // Mesh 子节点数
	G2       struct {
		OnlineStaCount int `json:"online_sta_count"`
	} `json:"2g"`
	G5 struct {
		OnlineStaCount int `json:"online_sta_count"`
	} `json:"5g"`
	G5H struct {
		OnlineStaCount int `json:"online_sta_count"`
	} `json:"5gh"`
}

// DeviceListV2 对应 /api/misystem/device_list，比 devicelist 提供更丰富的设备维度：
// wifi_quality（信号等级）、connectionType（连接类型）、pcontrol（家长控制）。
// 注意：该接口的 ip 字段是字符串（如 "192.168.31.172"），与 devicelist 接口的数组形式不同。
type DeviceListV2 struct {
	WanRX   int `json:"wanRX"`
	Devices []struct {
		Mac            string `json:"mac"`
		Name           string `json:"name"`
		OName          string `json:"oname"`
		IsAP           int    `json:"isap"`
		ConnectionType int    `json:"connectionType"`
		WifiQuality    int    `json:"wifi_quality"`
		IP             string `json:"ip"`
		PControl       struct {
			Enabled int `json:"enabled"`
			Total   int `json:"total"`
		} `json:"pcontrol"`
	} `json:"devices"`
	Code int `json:"code"`
}

var NewStatusRepo NewStatus
var DeviceListV2Repo DeviceListV2

func GetMiSystemNewStatus() {
	if err := fetchAPI("api/misystem/newstatus", &NewStatusRepo); err != nil {
		log.Println("获取 WiFi 频段状态失败：", err)
	}
}

func GetMiSystemDeviceListV2() {
	if err := fetchAPI("api/misystem/device_list", &DeviceListV2Repo); err != nil {
		log.Println("获取设备详情列表失败：", err)
	}
}
