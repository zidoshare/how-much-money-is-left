package main

import (
	"crypto/tls"
	"log"

	"gopkg.in/mail.v2"
)

//Sender 邮件发送器
type Sender struct {
	dialer *mail.Dialer
}

//Send 发送邮件
func (sender *Sender) Send(title, content string) {
	m := mail.NewMessage()
	m.SetHeader("From", GlobalConfig.Sender.Email)
	m.SetHeader("To", GlobalConfig.Sender.Targets...)
	m.SetHeader("Subject", title)
	m.SetBody("text/plain", content)
	if err := sender.dialer.DialAndSend(m); err != nil {
		log.Print(err)
	}
}

//NewSender 创建邮件发送器
func NewSender() *Sender {
	d := mail.NewDialer(GlobalConfig.Sender.Remote, GlobalConfig.Sender.Port, GlobalConfig.Sender.Email, GlobalConfig.Sender.Password)
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	return &Sender{
		dialer: d,
	}
}
