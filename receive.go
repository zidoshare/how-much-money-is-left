package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/emersion/go-imap"
	id "github.com/emersion/go-imap-id"
	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-message/charset"
	"github.com/emersion/go-message/mail"
	//"gopkg.in/mail.v2"
)

//MailClient 邮件客户端
type MailClient struct {
	client     *client.Client
	sender     *Sender
	httpClient *http.Client
}

func init() {

	imap.CharsetReader = charset.Reader
}

//NewClient 创建客户端
func NewClient() MailClient {
	log.Println("连接收信服务器...")
	c, err := client.Dial(GlobalConfig.Receive.Remote)
	if err != nil {
		log.Panicln(err)
	}
	log.Println("连接成功")

	// Login
	if err := c.Login(GlobalConfig.Receive.Email, GlobalConfig.Receive.Password); err != nil {
		log.Panicln(err)
	}
	log.Println("登录成功")
	clt := id.NewClient(c)
	clt.ID(map[string]string{
		"name":       "loveMail",
		"version":    "1.0.0",
		"os":         "Linux",
		"os-version": "5.1.1",
		"vendor":     "xxx",
		"contact":    "wuhongxu1208@Gmail.com",
	})
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	return MailClient{
		client:     c,
		sender:     NewSender(),
		httpClient: &http.Client{Transport: tr},
	}
}

//DisConnect 断开客户端链接
func (mailClient MailClient) DisConnect() {
	err := mailClient.client.Logout()
	if err != nil {
		fmt.Println("断开连接失败：", err)
	}
}

//Receive 接受
func (mailClient MailClient) Receive() {
	c := mailClient.client
	// INBOX, Sent Messages, 其他文件夹/QQ邮件订阅
	// Select INBOX 只读
	mbox, err := c.Select(GlobalConfig.Receive.Box, true)
	if err != nil {
		log.Panicln("选择收信文件夹失败：", err)
	}
	last := mbox.Messages
	to := last
	from, err := GetMessageID()
	if err != nil {
		log.Println("获取消息ID失败：", err)
		return
	}
	if from >= to {
		return
	}
	seqset := new(imap.SeqSet)
	log.Printf("最新邮件范围：%d,%d\n", from, to)
	seqset.AddRange(from+1, to)
	last = from

	messages := make(chan *imap.Message, to-from)
	done := make(chan error, 1)
	var section imap.BodySectionName
	items := []imap.FetchItem{section.FetchItem()}
	go func() {
		done <- c.Fetch(seqset, items, messages)
	}()

	for msg := range messages {
		r := msg.GetBody(&section)
		if r == nil {
			log.Printf("获取邮件内容失败:%+v\n", msg.Envelope)
			continue
		}
		mr, err := mail.CreateReader(r)
		if err != nil {
			log.Panicln("读取邮件内容错误：", err)
		}
		// Print some info about the message
		header := mr.Header
		if subject, err := header.Subject(); err == nil {
			if subject != "一卡通账户变动通知" {
				continue
			}
			log.Println(subject)
		} else {
			continue
		}
		if date, err := header.Date(); err == nil {
			log.Println("Date:", date)
		}
		// Process each message's part
		for {
			p, err := mr.NextPart()
			if err == io.EOF {
				break
			} else if err != nil {
				log.Panicln("错误", err)
			}

			switch p.Header.(type) {
			case *mail.InlineHeader:
				b, err := ioutil.ReadAll(p.Body)
				if err != nil {
					log.Println(err)
					continue
				}
				content := string(b)
				if strings.HasPrefix(content, "您账户"+GlobalConfig.Account) {
					log.Println("匹配到账户:" + GlobalConfig.Account)
					for _, oprator := range GetMatchers() {
						call := oprator.Match(content)
						if call != nil {
							balance, err := GetBalance()
							if err != nil {
								log.Println(err)
								continue
							}
							lastBalance := call(balance)
							log.Println("计算余额为：", lastBalance)
							SetBalance(lastBalance)
							one := ""
							resp, err := mailClient.httpClient.Get("http://v1.jinrishici.com/all.txt")
							if err != nil {
								log.Println(err)
							} else {
								bytes, err := ioutil.ReadAll(resp.Body)
								if err != nil {
									log.Println(err)
								}
								one = string(bytes) + "\n\n"
							}
							if lastBalance > balance {
								mailClient.sender.Send("咱们的钱变多啦！！！", fmt.Sprintf("%s%s，账户余额为：%.2f", one, content, lastBalance))
							} else if lastBalance < balance {
								mailClient.sender.Send("咱们的钱变少了？？？", fmt.Sprintf("%s%s，账户余额为：%.2f", one, content, lastBalance))
							}
						}
					}
				}
			}
		}
	}
	if err := <-done; err != nil {
		log.Panicln("收信失败", err)
	}
	if err := SetMessageID(to); err != nil {
		log.Println("设置收信ID失败：", err)
	}
	log.Println("当前批次处理完成")
}
