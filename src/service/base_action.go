package service

import "fmt"

func (service *BotActionService) BotAction_help(msg string, param map[string]interface{}) string {
	response := fmt.Sprintf("命令格式：%s命令 命令参数\n\n", service.config.CmdPrefix)
	i := 0
	for k, v := range service.botCmdList {
		i++
		response += fmt.Sprintf("%d:%s%s %s\n", i, service.config.CmdPrefix, k, v.Desc)
	}
	return response
}
