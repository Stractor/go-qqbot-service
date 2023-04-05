package service

import (
	"fmt"

	"github.com/sashabaranov/go-openai"
)

// 修改角色
func (svc *BotActionService) BotAction_changeRole(name string, metaData map[string]interface{}) string {
	userID := metaData["userID"].(int64)
	sessionID := fmt.Sprintf("%d_%s", userID, name)
	roleNow := svc.OpenAIService.getRoleNow(fmt.Sprintf("%d", userID))
	if roleNow == name {
		return fmt.Sprintf("cmd resonse: 我当前已是%s", name)
	}
	svc.OpenAIService.ChangeRole(fmt.Sprintf("%d", userID), sessionID)
	if !svc.OpenAIService.checkSessionExist(sessionID) {
		return svc.OpenAIService.CreateChatCompletion(sessionID, openai.ChatMessageRoleUser, "")
	}
	return fmt.Sprintf("cmd resonse: 我已切换成%s", name)
}

// 获取当前角色列表
func (svc *BotActionService) BotAction_roleList(param string, metaData map[string]interface{}) string {
	userID := metaData["userID"].(int64)
	defaultrole, roleList := svc.OpenAIService.QueryRoleList()
	roleNow := svc.OpenAIService.getRoleNow(fmt.Sprintf("%d", userID))
	respond := fmt.Sprintf("默认人设：%s\n当前人设：%s\n人设列表：\n", defaultrole, roleNow)
	for i := 0; i < len(roleList); i++ {
		respond = fmt.Sprintf("%s%d. %s\n", respond, i+1, roleList[i])
	}
	return respond
}

// 把现在的role全部保存到文件
func (svc *BotActionService) BotAction_saveRoleList(param string, metaData map[string]interface{}) string {
	return "暂不支持"
}

// 清空当前会话
func (svc *BotActionService) BotAction_resetSession(param string, metaData map[string]interface{}) string {
	userID := metaData["userID"].(int64)
	sessionID := svc.OpenAIService.getSessionIDByUser(fmt.Sprintf("%d", userID))
	return svc.OpenAIService.ResetCurrentSession(sessionID)
}

// 回退上一条设定
func (svc *BotActionService) BotAction_rollbackSystem(param string, metaData map[string]interface{}) string {
	userID := metaData["userID"].(int64)
	sessionID := svc.OpenAIService.getSessionIDByUser(fmt.Sprintf("%d", userID))
	return svc.OpenAIService.RollBackSystemMessage(sessionID, param)
}

// 回退上一条用户信息
func (svc *BotActionService) BotAction_rollbackUser(param string, metaData map[string]interface{}) string {
	userID := metaData["userID"].(int64)
	sessionID := svc.OpenAIService.getSessionIDByUser(fmt.Sprintf("%d", userID))
	return svc.OpenAIService.RollBackUserMessage(sessionID, param)
}

// 获取当前设定
func (svc *BotActionService) BotAction_describeCurrentSession(param string, metaData map[string]interface{}) string {
	userID := metaData["userID"].(int64)
	sessionID := svc.OpenAIService.getSessionIDByUser(fmt.Sprintf("%d", userID))
	roleNow := svc.OpenAIService.getRoleNow(fmt.Sprintf("%d", userID))
	systemMsg := svc.OpenAIService.DescribeCurrentSession(sessionID)
	respond := fmt.Sprintf("我的人设是：%s\n现在有%d条设定：\n", roleNow, len(systemMsg))
	for i := 0; i < len(systemMsg); i++ {
		respond = fmt.Sprintf("%s\n%d: %s\n", respond, i+1, systemMsg[i])
	}
	return respond
}

// 增加设定
func (svc *BotActionService) BotAction_addSystemMsg(param string, metaData map[string]interface{}) string {
	userID := metaData["userID"].(int64)
	sessionID := svc.OpenAIService.getSessionIDByUser(fmt.Sprintf("%d", userID))
	return svc.OpenAIService.CreateChatCompletion(sessionID, openai.ChatMessageRoleSystem, param)
}

// 清除用户聊天记录
func (svc *BotActionService) BotAction_clearUserMsg(param string, metaData map[string]interface{}) string {
	userID := metaData["userID"].(int64)
	sessionID := svc.OpenAIService.getSessionIDByUser(fmt.Sprintf("%d", userID))
	return svc.OpenAIService.ClearUserMessage(sessionID)
}

// 发消息到gpt
func (svc *BotActionService) BotAction_addUserMsg(param string, metaData map[string]interface{}) string {
	userID := metaData["userID"].(int64)
	sessionID := svc.OpenAIService.getSessionIDByUser(fmt.Sprintf("%d", userID))
	return svc.OpenAIService.CreateChatCompletion(sessionID, openai.ChatMessageRoleUser, param)
}
