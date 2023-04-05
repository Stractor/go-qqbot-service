package service

import (
	"fmt"
	"log"
	"reflect"
	"strings"
)

type BotActionService struct {
	config        *QQBotActionConfig
	botCmdList    map[string]*BotCmdItem
	OpenAIService *OpenAIService
}

func NewBotActionService(
	config *QQBotActionConfig,
	openAIService *OpenAIService,
) *BotActionService {

	botaction := &BotActionService{
		config:        config,
		OpenAIService: openAIService,
	}

	// 初始化bot命令
	botCmdList := map[string]*BotCmdItem{}

	for i := 0; i < len(config.CmdList); i++ {
		cmdItem := config.CmdList[i]
		register, err := botaction.registerCmd(cmdItem.Func)
		if err != nil {
			log.Printf("无法注册服务：%s，错误是：%s\n", cmdItem.Name, err)
			continue
		}
		// log.Printf("服务注册成功，%s:%s", cmdItem.Name, cmdItem.Func)
		botCmdList[cmdItem.Name] = &BotCmdItem{
			Name:   cmdItem.Name,
			Desc:   cmdItem.Desc,
			Action: register,
		}
	}

	botaction.botCmdList = botCmdList

	return botaction
}

type QQBotActionConfig struct {
	CmdPrefix string
	CmdList   []CmdListItem
}

type CmdListItem struct {
	Name string
	Func string
	Desc string
}

type BotCommandType struct {
	Name  string
	Param string
}

type BotCmdItem struct {
	Name   string
	Desc   string
	Action BotCmdAction
}

type BotCmdAction func(string, map[string]interface{}) string

func (service *BotActionService) parseCmd(msg string, info map[string]interface{}) *BotCommandType {
	if fmt.Sprintf("%c", msg[0]) != service.config.CmdPrefix {
		return nil
	}
	res := strings.SplitN(msg[1:], " ", 2)
	if len(res) <= 1 {
		return &BotCommandType{
			Name: res[0],
		}
	}
	return &BotCommandType{
		Name:  res[0],
		Param: res[1],
	}
}

func (service *BotActionService) registerCmd(name string) (BotCmdAction, error) {
	value := reflect.ValueOf(service)
	methodValue := value.MethodByName(fmt.Sprintf("BotAction_%s", name))
	invalidResponse := func(string, map[string]interface{}) string {
		return "method invalid"
	}
	if !methodValue.IsValid() {
		// 如果方法不存在，返回错误
		return invalidResponse, fmt.Errorf("method invalid")
	}
	responseFunc := func(name string, param map[string]interface{}) string {
		return methodValue.Call([]reflect.Value{
			reflect.ValueOf(name),
			reflect.ValueOf(param),
		})[0].String()
	}
	return BotCmdAction(responseFunc), nil
}

func (service *BotActionService) TriggerCmd(msg string, param map[string]interface{}) string {
	cmd := service.parseCmd(msg, param)
	cmdItem, ok := service.botCmdList[cmd.Name]
	if !ok {
		return "cmd response: 没有这个方法"
	}
	return cmdItem.Action(cmd.Param, param)
}
