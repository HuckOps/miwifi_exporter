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
			"memory_usage":         newGlobalMetric(namespace, "memory_usage", "", []string{"host"}),
			"memory_usage_percent": newGlobalMetric(namespace, "memory_usage_percent", "", []string{"host"}),
			"count_online":         newGlobalMetric(namespace, "count_online", "", []string{"host"}),
			"count_all":            newGlobalMetric(namespace, "count_all", "", []string{"host"}),
			"load_percent":         newGlobalMetric(namespace, "load_percent", "", []string{"host"}),
			"wan_upspeed":          newGlobalMetric(namespace, "wan_upspeed", "", []string{"host"}),
			"wan_downspeed":        newGlobalMetric(namespace, "wan_downspeed", "", []string{"host"}),
			"wan_up":               newGlobalMetric(namespace, "wan_up", "", []string{"host"}),
			"wan_down":             newGlobalMetric(namespace, "wan_down", "", []string{"host"}),
			"dev_upload":           newGlobalMetric(namespace, "dev_upload", "", []string{"ip"}),
			"dev_upspeed":          newGlobalMetric(namespace, "dev_upspeed", "", []string{"ip"}),
			"dev_download":         newGlobalMetric(namespace, "dev_download", "", []string{"ip"}),
			"dev_downspeed":        newGlobalMetric(namespace, "dev_downspeed", "", []string{"ip"}),
			"platform":             newGlobalMetric(namespace, "platform", "", []string{"platform"}),
			"version":              newGlobalMetric(namespace, "version", "", []string{"version"}),
			"sn":                   newGlobalMetric(namespace, "sn", "", []string{"sn"}),
			"mac":                  newGlobalMetric(namespace, "mac", "", []string{"mac"}),
			"ipv4":                 newGlobalMetric(namespace, "ipv4", "", []string{"ipv4"}),
			"ipv4_mask":            newGlobalMetric(namespace, "ipv4_mask", "", []string{"ipv4"}),
			"ipv6":                 newGlobalMetric(namespace, "ipv6", "", []string{"ipv6"}),
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
	ch <- prometheus.MustNewConstMetric(c.metrics["memory_usage_percent"], prometheus.CounterValue, float64(DevStatus.Mem.Usage*100), "miwifi")
	memory_total, _ := strconv.ParseFloat(strings.Split(DevStatus.Mem.Total, "MB")[0], 64)
	ch <- prometheus.MustNewConstMetric(c.metrics["memory_usage"], prometheus.GaugeValue, float64(DevStatus.Mem.Usage*memory_total), "miwifi")
	ch <- prometheus.MustNewConstMetric(c.metrics["count_all"], prometheus.GaugeValue, float64(DevStatus.Count.All), "miwifi")
	ch <- prometheus.MustNewConstMetric(c.metrics["count_online"], prometheus.GaugeValue, float64(DevStatus.Count.Online), "miwifi")
	ch <- prometheus.MustNewConstMetric(c.metrics["load_percent"], prometheus.GaugeValue, float64(DevStatus.CPU.Load*100), "miwifi")
	upspeed, _ := strconv.ParseFloat(DevStatus.Wan.Upspeed, 64)
	downspeed, _ := strconv.ParseFloat(DevStatus.Wan.Downspeed, 64)
	up, _ := strconv.ParseFloat(DevStatus.Wan.Upload, 64)
	down, _ := strconv.ParseFloat(DevStatus.Wan.Download, 64)
	ch <- prometheus.MustNewConstMetric(c.metrics["wan_upspeed"], prometheus.GaugeValue, upspeed, "miwifi")
	ch <- prometheus.MustNewConstMetric(c.metrics["wan_downspeed"], prometheus.GaugeValue, downspeed, "miwifi")
	ch <- prometheus.MustNewConstMetric(c.metrics["wan_up"], prometheus.GaugeValue, up, "miwifi")
	ch <- prometheus.MustNewConstMetric(c.metrics["wan_down"], prometheus.GaugeValue, down, "miwifi")
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
		ch <- prometheus.MustNewConstMetric(c.metrics["dev_upload"], prometheus.GaugeValue, upload, ip)
		ch <- prometheus.MustNewConstMetric(c.metrics["dev_download"], prometheus.GaugeValue, download, ip)
		ch <- prometheus.MustNewConstMetric(c.metrics["dev_upspeed"], prometheus.GaugeValue, devupspeed, ip)
		ch <- prometheus.MustNewConstMetric(c.metrics["dev_downspeed"], prometheus.GaugeValue, devdownspeed, ip)
	}
}
