package collector

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
)

type WanInfo struct {
	Info any `json:"info"`
	Code int `json:"code"`
}

type Info struct {
	Mac     string `json:"mac"`
	Mtu     string `json:"mtu"`
	Details struct {
		Username string `json:"username"`
		IfName   string `json:"ifname"`
		WanType  string `json:"wanType"`
		Service  string `json:"service"`
		Password string `json:"password"`
		PeerDns  string `json:"peerdns"`
	} `json:"details"`
	GateWay  string `json:"gateWay"`
	DnsAddr1 string `json:"dnsAddrs1"`
	Status   int    `json:"status"`
	Uptime   int    `json:"uptime"`
	DNSAddr  string `json:"dnsAddrs"`
	Ipv6Info struct {
		WanType      string        `json:"wanType"`
		IfName       string        `json:"ifname"`
		DNS          []interface{} `json:"dns"`
		IP6Addr      []string      `json:"ip6addr"`
		PeerDns      string        `json:"peerdns"`
		LanIP6Prefix []interface{} `json:"lan_ip6prefix"`
		LanIP6Addr   []interface{} `json:"lan_ip6addr"`
	} `json:"ipv6_info"`
	Ipv6Show int `json:"ipv6_show"`
	Link     int `json:"link"`
	Ipv4     []struct {
		Mask string `json:"mask"`
		IP   string `json:"ip"`
	} `json:"ipv4"`
}

type WifiDetailAll struct {
	Bsd  int `json:"bsd"`
	Info []struct {
		IfName      string `json:"ifname"`
		ChannelInfo struct {
			Bandwidth string   `json:"bandwidth"`
			BandList  []string `json:"bandList"`
			Channel   any      `json:"channel"`
		} `json:"channelInfo"`
		Encryption    string `json:"encryption"`
		Bandwidth     string `json:"bandwidth"`
		KickThreshold string `json:"kickthreshold"`
		Status        string `json:"status"`
		Mode          string `json:"mode"`
		Bsd           string `json:"bsd"`
		Ssid          string `json:"ssid"`
		WeakThreshold string `json:"weakthreshold"`
		Device        string `json:"device"`
		Ax            string `json:"ax"`
		Hidden        any    `json:"hidden"`
		Password      string `json:"password"`
		Channel       string `json:"channel"`
		TxPWR         string `json:"txpwr"`
		WeakEnable    string `json:"weakenable"`
		TxBF          string `json:"txbf"`
		Signal        int    `json:"signal"`
	} `json:"info"`
	Code int `json:"code"`
}

var WanInfoRepo Info
var WifiDetailAllRepo WifiDetailAll

func SubNetMaskToLen(netmask string) (int, error) {
	ipSplitArr := strings.Split(netmask, ".")
	if len(ipSplitArr) != 4 {
		return 0, fmt.Errorf("netmask:%v is not valid, pattern should like: 255.255.255.0", netmask)
	}
	ipv4MaskArr := make([]byte, 4)
	for i, value := range ipSplitArr {
		intValue, err := strconv.Atoi(value)
		if err != nil {
			return 0, fmt.Errorf("ipMaskToInt call strconv.Atoi error:[%v] string value is: [%s]", err, value)
		}
		if intValue > 255 {
			return 0, fmt.Errorf("netmask cannot greater than 255, current value is: [%s]", value)
		}
		ipv4MaskArr[i] = byte(intValue)
	}

	ones, _ := net.IPv4Mask(ipv4MaskArr[0], ipv4MaskArr[1], ipv4MaskArr[2], ipv4MaskArr[3]).Size()
	return ones, nil

}

func GetXQnetWorkWanInfo() {
	// wan_info 的 info 字段在 WAN 未连接时是 bool（false），连接时是对象，
	// 因此先用 WanInfo（info 为 any）解析，再做二次解析到 WanInfoRepo。
	wanInfoRepo := WanInfo{}
	if err := fetchAPI("api/xqnetwork/wan_info", &wanInfoRepo); err != nil {
		log.Println("获取 WAN 信息失败：", err)
		return
	}

	// WAN 未连接时 info 为 bool，此时没有可用信息。
	if _, ok := wanInfoRepo.Info.(bool); ok {
		return
	}

	m, err := json.Marshal(wanInfoRepo.Info)
	if err != nil {
		log.Println("解析 WAN 信息失败：", err)
		return
	}
	if err := json.Unmarshal(m, &WanInfoRepo); err != nil {
		log.Println("解析 WAN 信息失败：", err)
	}
}

func GetXQnetWorkWifiDetailAll() {
	if err := fetchAPI("api/xqnetwork/wifi_detail_all", &WifiDetailAllRepo); err != nil {
		log.Println("获取 WiFi 详情失败：", err)
	}
}

// PppoeStatus 对应 /api/xqnetwork/pppoe_status，提供 WAN 链路状态、协商速率、网关与 DNS。
type PppoeStatus struct {
	Code     int      `json:"code"`
	Status   int      `json:"status"`   // 2 = 已连接
	WanSpeed string   `json:"wanSpeed"` // WAN 协商速率（Mbps）
	GW       string   `json:"gw"`       // 网关
	Proto    string   `json:"proto"`    // 接入协议：dhcp / pppoe 等
	DNS      []string `json:"dns"`
	IP       struct {
		Mask    string `json:"mask"`
		Address string `json:"address"`
	} `json:"ip"`
}

// LanInfo 对应 /api/xqnetwork/lan_info，提供各 LAN 口链路状态与 LAN uptime。
type LanInfo struct {
	Code int `json:"code"`
	Info struct {
		Status int    `json:"status"`
		Mac    string `json:"mac"`
		Uptime int    `json:"uptime"`
		Ipv4   []struct {
			Mask string `json:"mask"`
			IP   string `json:"ip"`
		} `json:"ipv4"`
	} `json:"info"`
	LinkList []int `json:"linkList"` // 每个 LAN 口的链路状态（0/1），索引即端口号
}

// WifiConnectDevices 对应 /api/xqnetwork/wifi_connect_devices，
// 提供每台 WiFi 终端的信号强度（signal，0-100）与其所在频段（wifiIndex）。
type WifiConnectDevices struct {
	Code int `json:"code"`
	List []struct {
		Mac       string `json:"mac"`
		Signal    int    `json:"signal"`
		WifiIndex int    `json:"wifiIndex"` // 1=2.4G, 2=5G, 3=5Gh
	} `json:"list"`
}

// WifiList 对应 /api/xqnetwork/wifi_list，提供周围 WiFi 的扫描结果，
// 用于信道占用与干扰分析。
type WifiList struct {
	Code int `json:"code"`
	List []struct {
		Band       string `json:"band"`
		SSID       string `json:"ssid"`
		BSSID      string `json:"bssid"`
		RSSI       int    `json:"rssi"` // 信号强度（dBm，负值，越接近 0 越强）
		Channel    int    `json:"channel"`
		Signal     int    `json:"signal"`
		Quality    int    `json:"quality"`
		Encryption string `json:"encryption"`
	} `json:"list"`
}

var PppoeStatusRepo PppoeStatus
var LanInfoRepo LanInfo
var WifiConnectDevicesRepo WifiConnectDevices
var WifiListRepo WifiList

func GetXQnetWorkPppoeStatus() {
	if err := fetchAPI("api/xqnetwork/pppoe_status", &PppoeStatusRepo); err != nil {
		log.Println("获取 PPPoE/WAN 状态失败：", err)
	}
}

func GetXQnetWorkLanInfo() {
	if err := fetchAPI("api/xqnetwork/lan_info", &LanInfoRepo); err != nil {
		log.Println("获取 LAN 信息失败：", err)
	}
}

func GetXQnetWorkWifiConnectDevices() {
	if err := fetchAPI("api/xqnetwork/wifi_connect_devices", &WifiConnectDevicesRepo); err != nil {
		log.Println("获取 WiFi 终端信号失败：", err)
	}
}

func GetXQnetWorkWifiList() {
	if err := fetchAPI("api/xqnetwork/wifi_list", &WifiListRepo); err != nil {
		log.Println("获取周围 WiFi 列表失败：", err)
	}
}

// wifiIndexToBand 将路由器返回的 wifiIndex 映射为可读频段名。
func wifiIndexToBand(idx int) string {
	switch idx {
	case 1:
		return "2.4GHz"
	case 2:
		return "5GHz"
	case 3:
		return "5GHz-high"
	default:
		return "unknown"
	}
}
