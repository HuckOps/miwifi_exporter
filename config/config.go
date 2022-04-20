package config

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	"github.com/helloworlde/miwifi-exporter/token"
)

type ConfigStruct struct {
	IP       string `json:"ip"`
	Password string `json:"password"`
	Port     int    `json:"port"`
	// Stok string	`json:"stok"`
}

var Config ConfigStruct
var Token token.Auth

func GetConfig() {
	Config := ConfigStruct{}

	routerIp := os.Getenv("ROUTER_IP")
	routerPassword := os.Getenv("ROUTER_PASSWORD")

	if routerIp != "" && routerPassword != "" {
		Config.IP = routerIp
		Config.Password = routerPassword
	} else {
		log.Println("从 config.json 读取配置")
		f, err := os.Open("config.json")
		if err != nil {
			log.Println("读取配置文件失败")
			os.Exit(1)
		}
		byteValue, _ := ioutil.ReadAll(f)
		_ = json.Unmarshal(byteValue, &Config)
	}

	log.Println("获取到的地址: ", Config.IP, "密码: ", Config.Password)

	Token = token.GetToken(Config.IP, Config.Password)
}
