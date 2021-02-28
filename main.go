package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

//Service 服务
type Service struct {
	// Other things
	ch        chan bool
	waitGroup *sync.WaitGroup
	client    MailClient
	ticker    *time.Ticker
}

//NewService 创建服务
func NewService() *Service {
	return &Service{
		// Init Other things
		ch:        make(chan bool),
		waitGroup: &sync.WaitGroup{},
		client:    NewClient(),
		ticker:    time.NewTicker(time.Duration(time.Second * 2)),
	}
}

//Stop 关闭服务
func (s *Service) Stop() {
	close(s.ch)
	s.waitGroup.Wait()
	s.ticker.Stop()
	s.client.DisConnect()
}

//Serve 开启服务
func (s *Service) Serve() {
	s.waitGroup.Add(1)
	defer s.waitGroup.Done()
	for {
		select {
		case <-s.ch:
			fmt.Println("stopping...")
			return
		case <-s.ticker.C:
			defer func() {
				log.Println("receover")
				if err := recover(); err != nil {
					s.client.DisConnect()
					log.Println("接收邮件出错:", err)
					s.client = NewClient()
				}
			}()
			s.client.Receive()
		}
	}
}

func main() {
	service := NewService()
	defer service.Stop()
	go service.Serve()

	// Handle SIGINT and SIGTERM.
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
}
