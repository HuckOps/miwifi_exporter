package collector

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/helloworlde/miwifi-exporter/config"
)

type WanInfo struct {
	Info struct {
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
	} `json:"info"`
	Code int `json:"code"`
}

type WifiDetailAll struct {
	Bsd  int `json:"bsd"`
	Info []struct {
		IfName      string `json:"ifname"`
		ChannelInfo struct {
			Bandwidth string   `json:"bandwidth"`
			BandList  []string `json:"bandList"`
			Channel   int      `json:"channel"`
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
		Hidden        string `json:"hidden"`
		Password      string `json:"password"`
		Channel       string `json:"channel"`
		TxPWR         string `json:"txpwr"`
		WeakEnable    string `json:"weakenable"`
		TxBF          string `json:"txbf"`
		Signal        int    `json:"signal"`
	} `json:"info"`
	Code int `json:"code"`
}

var WanInfoRepo WanInfo
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
	client := http.Client{}
	res, err := client.Get(fmt.Sprintf("http://%s/cgi-bin/luci/;stok=%s/api/xqnetwork/wan_info",
		config.Configs.IP, config.Token.Token))
	if err != nil {
		log.Println("请求路由器错误，可能原因：1.路由器掉线或者宕机", err)
		os.Exit(1)
	}
	body, err := ioutil.ReadAll(res.Body)
	count := 0
	if err = json.Unmarshal(body, &WanInfoRepo); err != nil {
		log.Println("Token失效，正在重试获取")
		config.GetConfig()
		count++
		time.Sleep(1 * time.Minute)
		if count >= 5 {
			log.Println("获取状态错误，可能原因：1.账号或者密码错误，2.路由器鉴权错误", err)
			os.Exit(1)
		}
	}
}

func GetXQnetWorkWifiDetailAll() {
	client := http.Client{}
	res, err := client.Get(fmt.Sprintf("http://%s/cgi-bin/luci/;stok=%s/api/xqnetwork/wifi_detail_all",
		config.Configs.IP, config.Token.Token))
	if err != nil {
		log.Println("请求路由器错误，可能原因：1.路由器掉线或者宕机", err)
		os.Exit(1)
	}
	body, err := ioutil.ReadAll(res.Body)
	count := 0
	if err = json.Unmarshal(body, &WifiDetailAllRepo); err != nil {
		log.Println("Token失效，正在重试获取")
		config.GetConfig()
		count++
		time.Sleep(1 * time.Minute)
		if count >= 5 {
			log.Println("获取状态错误，可能原因：1.账号或者密码错误，2.路由器鉴权错误", err)
			os.Exit(1)
		}
	}
}
