package service

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/gorilla/websocket"
)

type WssService struct {
	config *WssConfig
}

type WssConfig struct {
	Url  string
	Port int
}

func NewWssService(config *WssConfig) *WssService {
	return &WssService{
		config: config,
	}
}

func (service *WssService) Start() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	// 连接到WebSocket服务器
	addr := fmt.Sprintf("%s:%d", service.config.Url, service.config.Port)
	conn, _, err := websocket.DefaultDialer.Dial(addr, nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer conn.Close()

	// 处理连接建立事件
	fmt.Println("Connected to WebSocket server")

	done := make(chan struct{})

	defer close(done)
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			return
		}
		fmt.Printf("Received message: %s\n", message)
	}

	// // 发送消息
	// ticker := time.NewTicker(1 * time.Second)
	// defer ticker.Stop()

	// for {
	// 	select {
	// 	case <-done:
	// 		return
	// 	case t := <-ticker.C:
	// 		err := conn.WriteMessage(websocket.TextMessage, []byte(t.String()))
	// 		if err != nil {
	// 			log.Println("write:", err)
	// 			return
	// 		}
	// 		fmt.Printf("Sent message: %s\n", t.String())
	// 	case <-interrupt:
	// 		log.Println("interrupt")

	// 		// 优雅地关闭连接
	// 		err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	// 		if err != nil {
	// 			log.Println("write close:", err)
	// 			return
	// 		}
	// 		select {
	// 		case <-done:
	// 		case <-time.After(time.Second):
	// 		}
	// 		return
	// 	}
	// }
}

func (service *WssService) handleMessage() error {
	return nil
}
