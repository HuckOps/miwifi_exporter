package collector

import (
	"fmt"
	"strconv"
	"strings"
	"sync"

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
			"memory_usage":            newGlobalMetric(namespace, "memory_usage", "", []string{"host"}),
			"memory_usage_percent":    newGlobalMetric(namespace, "memory_usage_percent", "", []string{"host"}),
			"online_device_amount":    newGlobalMetric(namespace, "online_device_amount", "", []string{"host"}),
			"history_device_amount":   newGlobalMetric(namespace, "history_device_amount", "", []string{"host"}),
			"load_percent":            newGlobalMetric(namespace, "load_percent", "", []string{"host"}),
			"wan_upload_speed":        newGlobalMetric(namespace, "wan_upload_speed", "", []string{"host"}),
			"wan_download_speed":      newGlobalMetric(namespace, "wan_download_speed", "", []string{"host"}),
			"wan_upload_traffic":      newGlobalMetric(namespace, "wan_upload_traffic", "", []string{"host"}),
			"wan_download_traffic":    newGlobalMetric(namespace, "wan_download_traffic", "", []string{"host"}),
			"device_upload_traffic":   newGlobalMetric(namespace, "device_upload_traffic", "", []string{"ip"}),
			"device_upload_speed":     newGlobalMetric(namespace, "device_upload_speed", "", []string{"ip"}),
			"device_download_traffic": newGlobalMetric(namespace, "device_download_traffic", "", []string{"ip"}),
			"device_download_speed":   newGlobalMetric(namespace, "device_download_speed", "", []string{"ip"}),
			"platform":                newGlobalMetric(namespace, "platform", "", []string{"platform"}),
			"version":                 newGlobalMetric(namespace, "version", "", []string{"version"}),
			"sn":                      newGlobalMetric(namespace, "sn", "", []string{"sn"}),
			"mac":                     newGlobalMetric(namespace, "mac", "", []string{"mac"}),
			"ipv4":                    newGlobalMetric(namespace, "ipv4", "", []string{"ipv4"}),
			"ipv4_mask":               newGlobalMetric(namespace, "ipv4_mask", "", []string{"ipv4"}),
			"ipv6":                    newGlobalMetric(namespace, "ipv6", "", []string{"ipv6"}),
		},
	}
}

func (c *Metrics) Describe(ch chan<- *prometheus.Desc) {
	for _, m := range c.metrics {
		ch <- m
	}
}

func (c *Metrics) Collect(ch chan<- prometheus.Metric) {
	c.mutex.Lock() // 加锁
	defer c.mutex.Unlock()
	GetMiwifiStatus()
	GetIPtoMAC()
	GetWAN()
	ch <- prometheus.MustNewConstMetric(c.metrics["memory_usage_percent"], prometheus.CounterValue, DevStatus.Mem.Usage*100, "miwifi")
	memory_total, _ := strconv.ParseFloat(strings.Split(DevStatus.Mem.Total, "MB")[0], 64)
	ch <- prometheus.MustNewConstMetric(c.metrics["memory_usage"], prometheus.GaugeValue, DevStatus.Mem.Usage*memory_total, "miwifi")
	ch <- prometheus.MustNewConstMetric(c.metrics["history_device_amount"], prometheus.GaugeValue, float64(DevStatus.Count.All), "miwifi")
	ch <- prometheus.MustNewConstMetric(c.metrics["online_device_amount"], prometheus.GaugeValue, float64(DevStatus.Count.Online), "miwifi")
	ch <- prometheus.MustNewConstMetric(c.metrics["load_percent"], prometheus.GaugeValue, DevStatus.CPU.Load*100, "miwifi")
	upspeed, _ := strconv.ParseFloat(DevStatus.Wan.Upspeed, 64)
	downspeed, _ := strconv.ParseFloat(DevStatus.Wan.Downspeed, 64)
	up, _ := strconv.ParseFloat(DevStatus.Wan.Upload, 64)
	down, _ := strconv.ParseFloat(DevStatus.Wan.Download, 64)
	ch <- prometheus.MustNewConstMetric(c.metrics["wan_upload_speed"], prometheus.GaugeValue, upspeed, "miwifi")
	ch <- prometheus.MustNewConstMetric(c.metrics["wan_download_speed"], prometheus.GaugeValue, downspeed, "miwifi")
	ch <- prometheus.MustNewConstMetric(c.metrics["wan_upload_traffic"], prometheus.GaugeValue, up, "miwifi")
	ch <- prometheus.MustNewConstMetric(c.metrics["wan_download_traffic"], prometheus.GaugeValue, down, "miwifi")
	ch <- prometheus.MustNewConstMetric(c.metrics["platform"], prometheus.GaugeValue, 1, DevStatus.Hardware.Platform)
	ch <- prometheus.MustNewConstMetric(c.metrics["version"], prometheus.GaugeValue, 1, DevStatus.Hardware.Version)
	ch <- prometheus.MustNewConstMetric(c.metrics["sn"], prometheus.GaugeValue, 1, DevStatus.Hardware.Sn)
	ch <- prometheus.MustNewConstMetric(c.metrics["mac"], prometheus.GaugeValue, 1, DevStatus.Hardware.Mac)
	count := 0

	for _, ipv4 := range WANInfo.Info.Ipv4 {
		ch <- prometheus.MustNewConstMetric(c.metrics["ipv4"], prometheus.GaugeValue, 1, ipv4.IP)
		mask, _ := SubNetMaskToLen(ipv4.Mask)
		ch <- prometheus.MustNewConstMetric(c.metrics["ipv4_mask"], prometheus.GaugeValue, float64(mask), ipv4.IP)
	}
	for _, ipv6 := range WANInfo.Info.Ipv6Info.IP6Addr {
		ch <- prometheus.MustNewConstMetric(c.metrics["ipv6"], prometheus.GaugeValue, 1, ipv6)
	}

	for _, dev := range DevStatus.Dev {
		defer func() {
			if err := recover(); err != nil {
				fmt.Println("存在不正常数据，请检查API", err)
			}
		}()
		upload, _ := strconv.ParseFloat(dev.Upload.(string), 64)
		download, _ := strconv.ParseFloat(dev.Download.(string), 64)
		devupspeed, _ := strconv.ParseFloat(dev.Upspeed.(string), 64)
		devdownspeed, _ := strconv.ParseFloat(dev.Downspeed.(string), 64)
		var ip string
		for _, d := range Mactoip.List {
			if d.Mac == dev.Mac {
				ip = d.IP[0].IP
				break
			}
		}

		if ip == "" {
			ip = fmt.Sprintf("未知设备%d", count)
			count++
		}
		ch <- prometheus.MustNewConstMetric(c.metrics["device_upload_traffic"], prometheus.GaugeValue, upload, ip)
		ch <- prometheus.MustNewConstMetric(c.metrics["device_download_traffic"], prometheus.GaugeValue, download, ip)
		ch <- prometheus.MustNewConstMetric(c.metrics["device_upload_speed"], prometheus.GaugeValue, devupspeed, ip)
		ch <- prometheus.MustNewConstMetric(c.metrics["device_download_speed"], prometheus.GaugeValue, devdownspeed, ip)
	}
}
