package collector

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"

	"github.com/helloworlde/miwifi-exporter/config"
	"github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	metrics map[string]*prometheus.Desc
	mutex   sync.Mutex
}

func newGlobalMetric(namespace string, metricName string, docString string, labels []string) *prometheus.Desc {
	return prometheus.NewDesc(namespace+"_"+metricName, docString, labels, nil)
}

func NewMetrics(namespace string) *Metrics {
	return &Metrics{
		metrics: map[string]*prometheus.Desc{
			"cpu_cores": newGlobalMetric(namespace, "cpu_cores", "", []string{"host"}),
			"cpu_mhz":   newGlobalMetric(namespace, "cpu_mhz", "", []string{"host"}),
			"cpu_load":  newGlobalMetric(namespace, "cpu_load", "", []string{"host"}),

			"memory_total_mb": newGlobalMetric(namespace, "memory_total_mb", "", []string{"host"}),
			"memory_usage_mb": newGlobalMetric(namespace, "memory_usage_mb", "", []string{"host"}),
			"memory_usage":    newGlobalMetric(namespace, "memory_usage", "", []string{"host"}),

			"count_all":                 newGlobalMetric(namespace, "count_all", "", []string{"host"}),
			"count_online":              newGlobalMetric(namespace, "count_online", "", []string{"host"}),
			"count_all_without_mash":    newGlobalMetric(namespace, "count_all_without_mash", "", []string{"host"}),
			"count_online_without_mash": newGlobalMetric(namespace, "count_online_without_mash", "", []string{"host"}),

			"uptime": newGlobalMetric(namespace, "uptime", "", []string{"host"}),

			"platform": newGlobalMetric(namespace, "platform", "", []string{"platform"}),
			"version":  newGlobalMetric(namespace, "version", "", []string{"version"}),
			"sn":       newGlobalMetric(namespace, "sn", "", []string{"sn"}),
			"mac":      newGlobalMetric(namespace, "mac", "", []string{"mac"}),

			"ipv4":      newGlobalMetric(namespace, "ipv4", "", []string{"ipv4"}),
			"ipv4_mask": newGlobalMetric(namespace, "ipv4_mask", "", []string{"ipv4"}),
			"ipv6":      newGlobalMetric(namespace, "ipv6", "", []string{"ipv6"}),

			"wan_upload_speed":     newGlobalMetric(namespace, "wan_upload_speed", "", []string{"host"}),
			"wan_download_speed":   newGlobalMetric(namespace, "wan_download_speed", "", []string{"host"}),
			"wan_upload_traffic":   newGlobalMetric(namespace, "wan_upload_traffic", "", []string{"host"}),
			"wan_download_traffic": newGlobalMetric(namespace, "wan_download_traffic", "", []string{"host"}),

			"device_upload_traffic":   newGlobalMetric(namespace, "device_upload_traffic", "", []string{"ip", "mac", "device_name", "is_ap"}),
			"device_upload_speed":     newGlobalMetric(namespace, "device_upload_speed", "", []string{"ip", "mac", "device_name", "is_ap"}),
			"device_download_traffic": newGlobalMetric(namespace, "device_download_traffic", "", []string{"ip", "mac", "device_name", "is_ap"}),
			"device_download_speed":   newGlobalMetric(namespace, "device_download_speed", "", []string{"ip", "mac", "device_name", "is_ap"}),
			"device_online_time":      newGlobalMetric(namespace, "device_online_time", "", []string{"ip", "mac", "device_name", "is_ap"}),

			"wifi_detail": newGlobalMetric(namespace, "wifi_detail", "", []string{"ssid", "status", "band_list", "channel"}),

			// 新增：各频段 WiFi 在线终端数与 Mesh 节点数（newstatus）
			"wifi_sta_count": newGlobalMetric(namespace, "wifi_sta_count", "各频段 WiFi 在线终端数", []string{"host", "band"}),
			"mesh_cap_count": newGlobalMetric(namespace, "mesh_cap_count", "Mesh 主节点数", []string{"host"}),
			"mesh_re_count":  newGlobalMetric(namespace, "mesh_re_count", "Mesh 子节点数", []string{"host"}),

			// 新增：WAN 链路状态（pppoe_status）
			"wan_link_status":       newGlobalMetric(namespace, "wan_link_status", "WAN 链路状态码（2=已连接）", []string{"host"}),
			"wan_negotiation_speed": newGlobalMetric(namespace, "wan_negotiation_speed", "WAN 协商速率（Mbps）", []string{"host"}),
			"wan_info":              newGlobalMetric(namespace, "wan_info", "WAN 信息", []string{"host", "wan_ip", "gateway", "dns"}),

			// 新增：各 LAN 口链路状态（lan_info）
			"lan_port_link": newGlobalMetric(namespace, "lan_port_link", "LAN 口链路状态（0=断开，1=连接）", []string{"host", "port"}),
			"lan_uptime":    newGlobalMetric(namespace, "lan_uptime", "LAN uptime（秒）", []string{"host"}),

			// 新增：每台 WiFi 终端的信号强度（wifi_connect_devices）
			"wifi_device_signal": newGlobalMetric(namespace, "wifi_device_signal", "WiFi 终端信号强度（0-100）", []string{"mac", "band"}),

			// 新增：周围 WiFi 信道环境（wifi_list）
			"neighbor_wifi_rssi":  newGlobalMetric(namespace, "neighbor_wifi_rssi", "周围 WiFi 信号强度（dBm）", []string{"ssid", "bssid", "band", "channel"}),
			"neighbor_wifi_count": newGlobalMetric(namespace, "neighbor_wifi_count", "周围 WiFi 总数", []string{"host"}),

			// 新增：设备更丰富的维度（device_list）
			"device_wifi_quality":    newGlobalMetric(namespace, "device_wifi_quality", "设备 WiFi 信号等级（1-4）", []string{"ip", "mac", "device_name"}),
			"device_connection_type": newGlobalMetric(namespace, "device_connection_type", "设备连接类型", []string{"ip", "mac", "device_name", "type"}),
			"device_pcontrol":        newGlobalMetric(namespace, "device_pcontrol", "设备是否启用家长控制（0/1）", []string{"ip", "mac", "device_name"}),
		},
	}
}

func (c *Metrics) Describe(ch chan<- *prometheus.Desc) {
	for _, m := range c.metrics {
		ch <- m
	}
}

func (c *Metrics) Collect(ch chan<- prometheus.Metric) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	GetMiSystemStatus()
	GetMiSystemDeviceList()
	GetXQnetWorkWanInfo()
	GetXQnetWorkWifiDetailAll()
	GetMiSystemNewStatus()
	GetMiSystemDeviceListV2()
	GetXQnetWorkPppoeStatus()
	GetXQnetWorkLanInfo()
	GetXQnetWorkWifiConnectDevices()
	GetXQnetWorkWifiList()
	host := config.GetHost()

	routerCPUHz := StatusRepo.GetRouterCPUMhz()
	routerMemoryTotal := StatusRepo.GetRouterMemoryTotal()
	routerUpSpeed := StatusRepo.GetRouterUpSpeed()
	routerDownSpeed := StatusRepo.GetRouterDownSpeed()
	routerUpload := StatusRepo.GetRouterUpload()
	routerDownload := StatusRepo.GetRouterDownload()
	routerUptime := StatusRepo.GetRouterUptime()

	ch <- prometheus.MustNewConstMetric(c.metrics["cpu_cores"], prometheus.GaugeValue, float64(StatusRepo.CPU.Core), host)
	ch <- prometheus.MustNewConstMetric(c.metrics["cpu_mhz"], prometheus.GaugeValue, routerCPUHz, host)
	ch <- prometheus.MustNewConstMetric(c.metrics["cpu_load"], prometheus.GaugeValue, StatusRepo.CPU.Load, host)

	ch <- prometheus.MustNewConstMetric(c.metrics["memory_total_mb"], prometheus.GaugeValue, routerMemoryTotal, host)
	ch <- prometheus.MustNewConstMetric(c.metrics["memory_usage_mb"], prometheus.GaugeValue, StatusRepo.Mem.Usage*routerMemoryTotal, host)
	ch <- prometheus.MustNewConstMetric(c.metrics["memory_usage"], prometheus.GaugeValue, StatusRepo.Mem.Usage, host)

	ch <- prometheus.MustNewConstMetric(c.metrics["count_all"], prometheus.GaugeValue, float64(StatusRepo.Count.All), host)
	ch <- prometheus.MustNewConstMetric(c.metrics["count_online"], prometheus.GaugeValue, float64(StatusRepo.Count.Online), host)
	ch <- prometheus.MustNewConstMetric(c.metrics["count_all_without_mash"], prometheus.GaugeValue, float64(StatusRepo.Count.AllWithoutMash), host)
	ch <- prometheus.MustNewConstMetric(c.metrics["count_online_without_mash"], prometheus.GaugeValue, float64(StatusRepo.Count.OnlineWithoutMash), host)

	ch <- prometheus.MustNewConstMetric(c.metrics["uptime"], prometheus.GaugeValue, routerUptime, host)

	ch <- prometheus.MustNewConstMetric(c.metrics["platform"], prometheus.GaugeValue, 1, StatusRepo.Hardware.Platform)
	ch <- prometheus.MustNewConstMetric(c.metrics["version"], prometheus.GaugeValue, 1, StatusRepo.Hardware.Version)
	ch <- prometheus.MustNewConstMetric(c.metrics["sn"], prometheus.GaugeValue, 1, StatusRepo.Hardware.Sn)
	ch <- prometheus.MustNewConstMetric(c.metrics["mac"], prometheus.GaugeValue, 1, StatusRepo.Hardware.Mac)

	ch <- prometheus.MustNewConstMetric(c.metrics["wan_upload_speed"], prometheus.GaugeValue, routerUpSpeed, host)
	ch <- prometheus.MustNewConstMetric(c.metrics["wan_download_speed"], prometheus.GaugeValue, routerDownSpeed, host)
	ch <- prometheus.MustNewConstMetric(c.metrics["wan_upload_traffic"], prometheus.GaugeValue, routerUpload, host)
	ch <- prometheus.MustNewConstMetric(c.metrics["wan_download_traffic"], prometheus.GaugeValue, routerDownload, host)

	for _, ipv4 := range WanInfoRepo.Ipv4 {
		ch <- prometheus.MustNewConstMetric(c.metrics["ipv4"], prometheus.GaugeValue, 1, ipv4.IP)
		mask, _ := SubNetMaskToLen(ipv4.Mask)
		ch <- prometheus.MustNewConstMetric(c.metrics["ipv4_mask"], prometheus.GaugeValue, float64(mask), ipv4.IP)
	}
	for _, ipv6 := range WanInfoRepo.Ipv6Info.IP6Addr {
		ch <- prometheus.MustNewConstMetric(c.metrics["ipv6"], prometheus.GaugeValue, 1, ipv6)
	}

	for _, dev := range StatusRepo.Dev {
		devUpload, err := InterfaceToFloat64(dev.Upload)
		if err != nil {
			log.Println("err: ", err)
		}
		devDownload, err := InterfaceToFloat64(dev.Download)
		if err != nil {
			log.Println("err: ", err)
		}

		var devIP string
		var devMac string
		var devName string
		var devIsAP string

		devMac = dev.Mac
		devName = dev.DevName
		for _, d := range DeviceListRepo.List {
			if d.Mac == dev.Mac && len(d.IP) > 0 {
				devIP = d.IP[0].IP
				devIsAP = strconv.Itoa(d.IsAP)
				break
			}
		}

		ch <- prometheus.MustNewConstMetric(c.metrics["device_upload_traffic"], prometheus.GaugeValue, devUpload, devIP, devMac, devName, devIsAP)
		ch <- prometheus.MustNewConstMetric(c.metrics["device_download_traffic"], prometheus.GaugeValue, devDownload, devIP, devMac, devName, devIsAP)
	}

	for _, dev := range DeviceListRepo.List {
		if len(dev.IP) > 0 {
			devMac := dev.Mac
			devName := dev.Name
			devIsAP := strconv.Itoa(dev.IsAP)
			devOnlineTime, err := InterfaceToFloat64(dev.Statistics.Online)
			if err != nil {
				log.Println("err: ", err)
			}
			devIP := dev.IP[0].IP
			devUpSpeed, err := InterfaceToFloat64(dev.Statistics.UpSpeed)
			if err != nil {
				log.Println("err: ", err)
			}
			devDownSpeed, err := InterfaceToFloat64(dev.Statistics.DownSpeed)
			if err != nil {
				log.Println("err: ", err)
			}

			ch <- prometheus.MustNewConstMetric(c.metrics["device_upload_speed"], prometheus.GaugeValue, devUpSpeed, devIP, devMac, devName, devIsAP)
			ch <- prometheus.MustNewConstMetric(c.metrics["device_download_speed"], prometheus.GaugeValue, devDownSpeed, devIP, devMac, devName, devIsAP)
			ch <- prometheus.MustNewConstMetric(c.metrics["device_online_time"], prometheus.GaugeValue, devOnlineTime, devIP, devMac, devName, devIsAP)
		}
	}

	for _, info := range WifiDetailAllRepo.Info {
		status, err := InterfaceToFloat64(info.Status)
		if err != nil {
			log.Println("err: ", err)
		}
		bandList := ""
		for i, band := range info.ChannelInfo.BandList {
			bandList += band
			if i != len(info.ChannelInfo.BandList)-1 {
				bandList += "/"
			} else {
				bandList += "MHz"
			}
		}
		channel := fmt.Sprintf("%v", info.ChannelInfo.Channel)
		ch <- prometheus.MustNewConstMetric(c.metrics["wifi_detail"], prometheus.GaugeValue, status, info.Ssid, info.Status, bandList, channel)
	}

	// 各频段 WiFi 在线终端数 + Mesh 节点数（newstatus）
	ch <- prometheus.MustNewConstMetric(c.metrics["wifi_sta_count"], prometheus.GaugeValue, float64(NewStatusRepo.G2.OnlineStaCount), host, "2.4GHz")
	ch <- prometheus.MustNewConstMetric(c.metrics["wifi_sta_count"], prometheus.GaugeValue, float64(NewStatusRepo.G5.OnlineStaCount), host, "5GHz")
	ch <- prometheus.MustNewConstMetric(c.metrics["wifi_sta_count"], prometheus.GaugeValue, float64(NewStatusRepo.G5H.OnlineStaCount), host, "5GHz-high")
	ch <- prometheus.MustNewConstMetric(c.metrics["mesh_cap_count"], prometheus.GaugeValue, float64(NewStatusRepo.CapCount), host)
	ch <- prometheus.MustNewConstMetric(c.metrics["mesh_re_count"], prometheus.GaugeValue, float64(NewStatusRepo.ReCount), host)

	// WAN 链路状态（pppoe_status）
	ch <- prometheus.MustNewConstMetric(c.metrics["wan_link_status"], prometheus.GaugeValue, float64(PppoeStatusRepo.Status), host)
	if wanSpeed, err := strconv.ParseFloat(PppoeStatusRepo.WanSpeed, 64); err == nil {
		ch <- prometheus.MustNewConstMetric(c.metrics["wan_negotiation_speed"], prometheus.GaugeValue, wanSpeed, host)
	}
	dns := strings.Join(PppoeStatusRepo.DNS, ",")
	ch <- prometheus.MustNewConstMetric(c.metrics["wan_info"], prometheus.GaugeValue, 1, host, PppoeStatusRepo.IP.Address, PppoeStatusRepo.GW, dns)

	// 各 LAN 口链路状态（lan_info）
	for i, link := range LanInfoRepo.LinkList {
		ch <- prometheus.MustNewConstMetric(c.metrics["lan_port_link"], prometheus.GaugeValue, float64(link), host, strconv.Itoa(i))
	}
	ch <- prometheus.MustNewConstMetric(c.metrics["lan_uptime"], prometheus.GaugeValue, float64(LanInfoRepo.Info.Uptime), host)

	// 每台 WiFi 终端的信号强度（wifi_connect_devices）
	for _, d := range WifiConnectDevicesRepo.List {
		band := wifiIndexToBand(d.WifiIndex)
		ch <- prometheus.MustNewConstMetric(c.metrics["wifi_device_signal"], prometheus.GaugeValue, float64(d.Signal), d.Mac, band)
	}

	// 周围 WiFi 信道环境（wifi_list）
	for _, w := range WifiListRepo.List {
		ch <- prometheus.MustNewConstMetric(c.metrics["neighbor_wifi_rssi"], prometheus.GaugeValue, float64(w.RSSI), w.SSID, w.BSSID, w.Band, strconv.Itoa(w.Channel))
	}
	ch <- prometheus.MustNewConstMetric(c.metrics["neighbor_wifi_count"], prometheus.GaugeValue, float64(len(WifiListRepo.List)), host)

	// 设备更丰富的维度（device_list）
	for _, dev := range DeviceListV2Repo.Devices {
		devIP := dev.IP
		devMac := dev.Mac
		devName := dev.Name
		if devName == "" {
			devName = dev.OName
		}
		ch <- prometheus.MustNewConstMetric(c.metrics["device_wifi_quality"], prometheus.GaugeValue, float64(dev.WifiQuality), devIP, devMac, devName)
		ch <- prometheus.MustNewConstMetric(c.metrics["device_connection_type"], prometheus.GaugeValue, 1, devIP, devMac, devName, strconv.Itoa(dev.ConnectionType))
		ch <- prometheus.MustNewConstMetric(c.metrics["device_pcontrol"], prometheus.GaugeValue, float64(dev.PControl.Enabled), devIP, devMac, devName)
	}
}

func InterfaceToFloat64(n interface{}) (float64, error) {
	switch x := n.(type) {
	case string:
		return strconv.ParseFloat(n.(string), 64)
	case float32:
		return float64(x), nil
	case float64:
		return x, nil
	case int64:
		return float64(x), nil
	case int32:
		return float64(x), nil
	case int:
		return float64(x), nil
	case uint64:
		return float64(x), nil
	case uint32:
		return float64(x), nil
	case uint:
		return float64(x), nil
	default:
		return 0.0, nil
	}
}
