package controller

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"

	"example.com/m/v2/src/service"
	"github.com/gorilla/websocket"
)

type Controller struct {
	config *WssConfig
}

func NewController(
	config *WssConfig,
) *Controller {
	return &Controller{
		config: config,
	}
}

type WssConfig struct {
	Url  string
	Port int
}

func (ctr *Controller) Start() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	// 连接到WebSocket服务器
	addr := fmt.Sprintf("%s:%d", ctr.config.Url, ctr.config.Port)
	conn, _, err := websocket.DefaultDialer.Dial(addr, nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer conn.Close()

	// 处理连接建立事件
	fmt.Println("Connected to WebSocket server")

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			return
		}
		fmt.Printf("Received message: %s\n", message)
		ctr.handleMessage(message)
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

func (ctr *Controller) handleMessage(byteMsg []byte) error {
	var msg service.QQMessageType
	err := json.Unmarshal(byteMsg, &msg)
	if err != nil {
		log.Printf("decode message error: %s\n ", err)
		return err
	}
	if service.QQ_MSG_POST_TYPE(msg.PostType) == service.POST_TYPE_MESSAGE {
		switch service.QQ_MSG_MESSAGE_TYPE(msg.MessageType) {
		case service.MESSAGE_TYPE_GROUP:
		case service.MESSAGE_TYPE_PRIVATE:
		}
	}
	return nil
}
