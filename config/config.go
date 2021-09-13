package config

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"miwifi-exporter/token"
	"os"
)

type ConfigStruct struct {
	IP       string `json:"ip"`
	Password string `json:"password"`
	Port     int    `json:"port"`
	//Stok string	`json:"stok"`
}

var Config ConfigStruct
var Token token.Auth

func GetConfig() {
	config := ConfigStruct{}
	f, err := os.Open("config.json")
	if err != nil {
		log.Println("读取配置文件失败")
		os.Exit(1)
	}
	byteValue, _ := ioutil.ReadAll(f)
	json.Unmarshal([]byte(byteValue), &config)
	Config = config
	Token = token.GetToken(Config.IP, config.Password)
}
