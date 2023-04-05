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
	config          *WssConfig
	actionService   *service.BotActionService
	botActionConfig *service.QQBotActionConfig
}

func NewController(
	config *WssConfig,
	actionService *service.BotActionService,
	botActionConfig *service.QQBotActionConfig,
) *Controller {
	return &Controller{
		config:          config,
		actionService:   actionService,
		botActionConfig: botActionConfig,
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
		// fmt.Printf("Received message: %s\n", message)
		go ctr.handleMessage(conn, message)
	}
}

func (ctr *Controller) handleMessage(conn *websocket.Conn, byteMsg []byte) error {
	var msg service.QQMessageType
	err := json.Unmarshal(byteMsg, &msg)
	if err != nil {
		log.Printf("decode message error: %s\n ", err)
		return err
	}
	if service.QQ_MSG_POST_TYPE(msg.PostType) == service.POST_TYPE_MESSAGE {
		fmt.Printf("Received message: %s\n", string(byteMsg))
		switch service.QQ_MSG_MESSAGE_TYPE(msg.MessageType) {
		case service.MESSAGE_TYPE_GROUP:
		case service.MESSAGE_TYPE_PRIVATE:
			finalMsg := msg.Message
			if msg.Message[0:1] != ctr.botActionConfig.CmdPrefix {
				finalMsg = fmt.Sprintf("%s发送消息到gpt %s", ctr.botActionConfig.CmdPrefix, msg.Message)
			}
			ctr.pushMessageBackToQQ(conn, service.MESSAGE_TYPE_PRIVATE, msg.UserID,
				ctr.actionService.TriggerCmd(finalMsg, map[string]interface{}{
					"userID": msg.UserID,
				}))
		}
	}
	return nil
}

func (ctr *Controller) pushMessageBackToQQ(conn *websocket.Conn, messageType service.QQ_MSG_MESSAGE_TYPE, userID int64, message string) error {
	sendMsg, err := json.Marshal(service.QQMessageSendModel{
		Action: "send_private_msg",
		Params: service.QQMessageSendDetail{
			MessageType: messageType,
			UserId:      userID,
			Message:     message,
		},
	})
	err = conn.WriteMessage(websocket.TextMessage, sendMsg)
	if err != nil {
		log.Printf("发送消息到出错，消息内容：%s\n错误是：%s", sendMsg, err)
		return err
	}
	fmt.Printf("Sent message: %s\n", string(sendMsg))
	return nil
}
