package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/helloworlde/miwifi-exporter/collector"
	"github.com/helloworlde/miwifi-exporter/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	metricsPath      = flag.String("web.telemetry-path", "/metrics", "A path under which to expose metrics.")
	metricsNamespace = flag.String("metric.namespace", "miwifi", "Prometheus metrics namespace, as the prefix of metrics name")
)

func main() {
	log.Println("欢迎使用小米路由器监控prometheus客户端，项目名miwifi_exporter，作者：Huck，欢迎提交issues、PullRequest")
	log.Println("初始化程序")
	flag.Parse()
	config.GetConfig()
	log.Println("初始化完成")

	log.Println("初始化监控指标")
	metrics := collector.NewMetrics(*metricsNamespace)
	registry := prometheus.NewRegistry()
	registry.MustRegister(metrics)
	log.Println("监控指标初始化注册完成")

	log.Println("启动服务器，监听端口为:" + strconv.Itoa(config.Configs.Port))
	http.Handle(*metricsPath, promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte(`<html>
			<head><title>A Prometheus Exporter</title></head>
			<body>
			<h1>A Prometheus Exporter</h1>
			<p><a href='/metrics'>Metrics</a></p>
			</body>
			</html>`))
		if err != nil {
			log.Println("运行exporter错误", err)
			os.Exit(1)
		}
	})

	log.Printf("监控Metrics位置： http://localhost:%d%s", config.Configs.Port, *metricsPath)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(config.Configs.Port), nil))
}
