package collector

import (
	"log"
	"strconv"
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

			"online_device_amount":  newGlobalMetric(namespace, "online_device_amount", "", []string{"host"}),
			"history_device_amount": newGlobalMetric(namespace, "history_device_amount", "", []string{"host"}),

			"platform":             newGlobalMetric(namespace, "platform", "", []string{"platform"}),
			"version":              newGlobalMetric(namespace, "version", "", []string{"version"}),
			"sn":                   newGlobalMetric(namespace, "sn", "", []string{"sn"}),
			"mac":                  newGlobalMetric(namespace, "mac", "", []string{"mac"}),
			"ipv4":                 newGlobalMetric(namespace, "ipv4", "", []string{"ipv4"}),
			"ipv4_mask":            newGlobalMetric(namespace, "ipv4_mask", "", []string{"ipv4"}),
			"ipv6":                 newGlobalMetric(namespace, "ipv6", "", []string{"ipv6"}),
			"wan_upload_speed":     newGlobalMetric(namespace, "wan_upload_speed", "", []string{"host"}),
			"wan_download_speed":   newGlobalMetric(namespace, "wan_download_speed", "", []string{"host"}),
			"wan_upload_traffic":   newGlobalMetric(namespace, "wan_upload_traffic", "", []string{"host"}),
			"wan_download_traffic": newGlobalMetric(namespace, "wan_download_traffic", "", []string{"host"}),

			"device_upload_traffic":   newGlobalMetric(namespace, "device_upload_traffic", "", []string{"ip", "mac", "device_name", "is_ap"}),
			"device_upload_speed":     newGlobalMetric(namespace, "device_upload_speed", "", []string{"ip", "mac", "device_name", "is_ap"}),
			"device_download_traffic": newGlobalMetric(namespace, "device_download_traffic", "", []string{"ip", "mac", "device_name", "is_ap"}),
			"device_download_speed":   newGlobalMetric(namespace, "device_download_speed", "", []string{"ip", "mac", "device_name", "is_ap"}),
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

	GetMiWifiStatus()
	GetMiWifiDeviceList()
	GetWifiWanInfo()
	host := config.GetHost()

	routerCPUHz := StatusRepo.GetRouterCPUMhz()
	routerMemoryTotal := StatusRepo.GetRouterMemoryTotal()
	routerUpSpeed := StatusRepo.GetRouterUpSpeed()
	routerDownSpeed := StatusRepo.GetRouterDownSpeed()
	routerUpload := StatusRepo.GetRouterUpload()
	routerDownload := StatusRepo.GetRouterDownload()

	ch <- prometheus.MustNewConstMetric(c.metrics["cpu_cores"], prometheus.GaugeValue, float64(StatusRepo.CPU.Core), host)
	ch <- prometheus.MustNewConstMetric(c.metrics["cpu_mhz"], prometheus.GaugeValue, routerCPUHz, host)
	ch <- prometheus.MustNewConstMetric(c.metrics["cpu_load"], prometheus.GaugeValue, StatusRepo.CPU.Load, host)

	ch <- prometheus.MustNewConstMetric(c.metrics["memory_total_mb"], prometheus.GaugeValue, routerMemoryTotal, host)
	ch <- prometheus.MustNewConstMetric(c.metrics["memory_usage_mb"], prometheus.GaugeValue, StatusRepo.Mem.Usage*routerMemoryTotal, host)
	ch <- prometheus.MustNewConstMetric(c.metrics["memory_usage"], prometheus.GaugeValue, StatusRepo.Mem.Usage, host)

	ch <- prometheus.MustNewConstMetric(c.metrics["history_device_amount"], prometheus.GaugeValue, float64(StatusRepo.Count.All), host)
	ch <- prometheus.MustNewConstMetric(c.metrics["online_device_amount"], prometheus.GaugeValue, float64(StatusRepo.Count.Online), host)

	ch <- prometheus.MustNewConstMetric(c.metrics["wan_upload_speed"], prometheus.GaugeValue, routerUpSpeed, host)
	ch <- prometheus.MustNewConstMetric(c.metrics["wan_download_speed"], prometheus.GaugeValue, routerDownSpeed, host)
	ch <- prometheus.MustNewConstMetric(c.metrics["wan_upload_traffic"], prometheus.GaugeValue, routerUpload, host)
	ch <- prometheus.MustNewConstMetric(c.metrics["wan_download_traffic"], prometheus.GaugeValue, routerDownload, host)
	ch <- prometheus.MustNewConstMetric(c.metrics["platform"], prometheus.GaugeValue, 1, StatusRepo.Hardware.Platform)
	ch <- prometheus.MustNewConstMetric(c.metrics["version"], prometheus.GaugeValue, 1, StatusRepo.Hardware.Version)
	ch <- prometheus.MustNewConstMetric(c.metrics["sn"], prometheus.GaugeValue, 1, StatusRepo.Hardware.Sn)
	ch <- prometheus.MustNewConstMetric(c.metrics["mac"], prometheus.GaugeValue, 1, StatusRepo.Hardware.Mac)

	for _, ipv4 := range WanInfoRepo.Info.Ipv4 {
		ch <- prometheus.MustNewConstMetric(c.metrics["ipv4"], prometheus.GaugeValue, 1, ipv4.IP)
		mask, _ := SubNetMaskToLen(ipv4.Mask)
		ch <- prometheus.MustNewConstMetric(c.metrics["ipv4_mask"], prometheus.GaugeValue, float64(mask), ipv4.IP)
	}
	for _, ipv6 := range WanInfoRepo.Info.Ipv6Info.IP6Addr {
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
		devUpSpeed, err := InterfaceToFloat64(dev.UpSpeed)
		if err != nil {
			log.Println("err: ", err)
		}
		devDownSpeed, err := InterfaceToFloat64(dev.DownSpeed)
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
		ch <- prometheus.MustNewConstMetric(c.metrics["device_upload_speed"], prometheus.GaugeValue, devUpSpeed, devIP, devMac, devName, devIsAP)
		ch <- prometheus.MustNewConstMetric(c.metrics["device_download_speed"], prometheus.GaugeValue, devDownSpeed, devIP, devMac, devName, devIsAP)
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
