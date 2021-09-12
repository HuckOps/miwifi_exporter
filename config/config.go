package config

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
)

type ConfigStruct struct {
	IP string `json:"ip"`
	Stok string	`json:"stok"`
}

var Config ConfigStruct

func GetConfig(){
	config := ConfigStruct{}
	f, err := os.Open("config.json")
	if err != nil {
		log.Println("读取配置文件失败")
		os.Exit(1)
	}
	byteValue, _ := ioutil.ReadAll(f)
	json.Unmarshal([]byte(byteValue), &config)
	Config = config
}