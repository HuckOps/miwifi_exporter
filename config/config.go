package config

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	"github.com/helloworlde/miwifi-exporter/token"
)

type Config struct {
	IP       string `json:"ip"`
	Password string `json:"password"`
	Port     int    `json:"port"`
	Host     string `json:"host"`
}

var Configs Config
var Token token.Auth

func GetConfig() {
	config := Config{}
	config.Port = 9001
	config.Host = "miwifi"

	routerIp := os.Getenv("ROUTER_IP")
	routerPassword := os.Getenv("ROUTER_PASSWORD")
	routerHost := os.Getenv("ROUTER_HOST")
	if routerHost != "" {
		config.Host = routerHost
	}

	if routerIp != "" && routerPassword != "" {
		config.IP = routerIp
		config.Password = routerPassword
	} else {
		log.Println("从 config.json 读取配置")
		f, err := os.Open("config.json")
		if err != nil {
			log.Println("读取配置文件失败")
			os.Exit(1)
		}
		byteValue, _ := ioutil.ReadAll(f)
		_ = json.Unmarshal(byteValue, &config)
	}

	Configs = config
	log.Println("获取到的地址: ", Configs.IP, "密码: ", Configs.Password, "端口: ", Configs.Port)
	Token = token.GetToken(Configs.IP, Configs.Password)
}

func GetHost() string {
	return Configs.Host
}
