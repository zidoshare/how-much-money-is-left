package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

//SetBalance 设置余额
func SetBalance(balance float64) error {
	messageID, err := GetMessageID()
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filepath.Join(configPath, "balance"), []byte(fmt.Sprintf("%.2f\n%d", balance, messageID)), os.ModePerm)
}

//GetBalance 获取余额
func GetBalance() (balance float64, err error) {
	bytes, err := ioutil.ReadFile(filepath.Join(configPath, "balance"))
	content := strings.Split(strings.ReplaceAll(string(bytes), "\r", ""), "\n")[0]
	if err != nil {
		return
	}
	balance, err = strconv.ParseFloat(content, 64)
	return
}

//GetMessageID 当前最新已读消息ID
func GetMessageID() (ID uint32, err error) {
	bytes, err := ioutil.ReadFile(filepath.Join(configPath, "balance"))
	if err != nil {
		return
	}
	id, err := strconv.Atoi(strings.Split(strings.ReplaceAll(string(bytes), "\r", ""), "\n")[1])
	ID = uint32(id)
	return
}

//SetMessageID 设置当前已读消息ID
func SetMessageID(ID uint32) error {
	balance, err := GetBalance()
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filepath.Join(configPath, "balance"), []byte(fmt.Sprintf("%.2f\n%d", balance, ID)), os.ModePerm)

}

//MatcherConfig 匹配器配置类
type MatcherConfig struct {
	Pattern string
	Op      string
}

//ClientConfig 接收邮件客户端配置
type ClientConfig struct {
	Remote   string `json:"remote"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Box      string `json:"box"`
}

//SenderConfig 发送邮件客户端配置
type SenderConfig struct {
	Remote   string   `json:"remote"`
	Port     int      `json:"port"`
	Email    string   `json:"email"`
	Password string   `json:"password"`
	Targets  []string `json:"targets"`
}

//JSONConfig JSON配置聚合
type JSONConfig struct {
	Account  string          `json:"account"`
	Patterns []MatcherConfig `json:"patterns"`
	Receive  ClientConfig    `json:"receive"`
	Sender   SenderConfig    `json:"sender"`
}

//GlobalConfig 全局配置
var GlobalConfig = &JSONConfig{}
var globalMatchers []OpMatcher

var configPath string

func init() {
	flag.StringVar(&configPath, "config", ".", "指定配置目录")
	flag.Parse()
	bytes, err := ioutil.ReadFile(filepath.Join(configPath, "config.json"))
	if err != nil {
		log.Panicln("读取配置文件出错：", err)
	}
	if err = json.Unmarshal(bytes, GlobalConfig); err != nil {
		log.Panicln("读取配置文件出错：", err)
	}
}
