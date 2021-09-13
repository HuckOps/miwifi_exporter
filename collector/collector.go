package collector

import (
	"encoding/json"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"io/ioutil"
	"log"
	"miwifi-exporter/config"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
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
			"dev_upload":           newGlobalMetric(namespace, "dev_upload", "", []string{"host"}),
			"dev_upspeed":          newGlobalMetric(namespace, "dev_upspeed", "", []string{"host"}),
			"dev_download":         newGlobalMetric(namespace, "dev_download", "", []string{"host"}),
			"dev_downspeed":        newGlobalMetric(namespace, "dev_downspeed", "", []string{"host"}),
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
	ch <- prometheus.MustNewConstMetric(c.metrics["memory_usage_percent"], prometheus.CounterValue, float64(status.Mem.Usage*100), "miwifi")
	memory_total, _ := strconv.ParseFloat(strings.Split(status.Mem.Total, "MB")[0], 64)
	ch <- prometheus.MustNewConstMetric(c.metrics["memory_usage"], prometheus.GaugeValue, float64(status.Mem.Usage*memory_total), "miwifi")
	ch <- prometheus.MustNewConstMetric(c.metrics["count_all"], prometheus.GaugeValue, float64(status.Count.All), "miwifi")
	ch <- prometheus.MustNewConstMetric(c.metrics["count_online"], prometheus.GaugeValue, float64(status.Count.Online), "miwifi")
	ch <- prometheus.MustNewConstMetric(c.metrics["load_percent"], prometheus.GaugeValue, float64(status.CPU.Load*100), "miwifi")
	upspeed, _ := strconv.ParseFloat(status.Wan.Upspeed, 64)
	downspeed, _ := strconv.ParseFloat(status.Wan.Downspeed, 64)
	up, _ := strconv.ParseFloat(status.Wan.Upload, 64)
	down, _ := strconv.ParseFloat(status.Wan.Download, 64)
	ch <- prometheus.MustNewConstMetric(c.metrics["wan_upspeed"], prometheus.GaugeValue, upspeed, "miwifi")
	ch <- prometheus.MustNewConstMetric(c.metrics["wan_downspeed"], prometheus.GaugeValue, downspeed, "miwifi")
	ch <- prometheus.MustNewConstMetric(c.metrics["wan_up"], prometheus.GaugeValue, up, "miwifi")
	ch <- prometheus.MustNewConstMetric(c.metrics["wan_down"], prometheus.GaugeValue, down, "miwifi")
	count := 0
	for _, dev := range status.Dev {
		upload, _ := strconv.ParseFloat(dev.Upload, 64)
		download, _ := strconv.ParseFloat(dev.Download, 64)
		devupspeed, _ := strconv.ParseFloat(dev.Upspeed, 64)
		devdownspeed, _ := strconv.ParseFloat(dev.Downspeed, 64)
		var Devname string
		if dev.Devname == "Unknown" {
			Devname = dev.Devname + strconv.Itoa(count)
			count = count + 1
		} else {
			Devname = dev.Devname
		}
		ch <- prometheus.MustNewConstMetric(c.metrics["dev_upload"], prometheus.GaugeValue, upload, Devname)
		ch <- prometheus.MustNewConstMetric(c.metrics["dev_download"], prometheus.GaugeValue, download, Devname)
		ch <- prometheus.MustNewConstMetric(c.metrics["dev_upspeed"], prometheus.GaugeValue, devupspeed, Devname)
		ch <- prometheus.MustNewConstMetric(c.metrics["dev_downspeed"], prometheus.GaugeValue, devdownspeed, Devname)
	}
}

type Status struct {
	Dev []struct {
		Mac              string `json:"mac"`
		Maxdownloadspeed string `json:"maxdownloadspeed"`
		Upload           string `json:"upload"`
		Upspeed          string `json:"upspeed"`
		Downspeed        string `json:"downspeed"`
		Online           string `json:"online"`
		Devname          string `json:"devname"`
		Maxuploadspeed   string `json:"maxuploadspeed"`
		Download         string `json:"download"`
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

var status Status

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
	if err = json.Unmarshal([]byte(body), &status); err != nil {
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
