package service

import (
	"context"
	"fmt"
	"log"
	"strings"

	openai "github.com/sashabaranov/go-openai"
)

type OpenAIService struct {
	client         *openai.Client
	openAIConfig   *OpenAIConfig
	sessions       map[string][]openai.ChatCompletionMessage
	userSessionNow map[string]string
	roleConfig     RoleConfig
}

func NewOpenAIService(
	openAIConfig *OpenAIConfig,
	roleConfig *RoleConfig,
	privicyConfig *PrivacyConfig,
) *OpenAIService {
	client := openai.NewClient(privicyConfig.Token)

	sessions := map[string][]openai.ChatCompletionMessage{}
	return &OpenAIService{
		client:         client,
		openAIConfig:   openAIConfig,
		sessions:       sessions,
		userSessionNow: map[string]string{},
		roleConfig:     *roleConfig,
	}
}

type OpenAIConfig struct {
	Model           string
	MaxToken        int
	Temperature     float32
	TopP            float32
	UserMemory      int
	AssistantMemory int
	SystemMemory    int
}

type PrivacyConfig struct {
	Token string
}

type RoleConfig struct {
	DefaultRole string
	RoleList    map[string][]string
}

func (service *OpenAIService) ListModel() ([]openai.Model, error) {
	modelList, err := service.client.ListModels(context.TODO())
	if err != nil {
		return []openai.Model{}, err
	}
	return modelList.Models, nil
}

// 根据prompt续写
func (service *OpenAIService) CreateCompletion(prompt string) (openai.CompletionChoice, error) {
	request := openai.CompletionRequest{
		Model:     service.openAIConfig.Model,
		Prompt:    prompt,
		MaxTokens: service.openAIConfig.MaxToken,
	}
	if service.openAIConfig.Temperature > -0.000001 {
		request.Temperature = service.openAIConfig.Temperature
	}
	if service.openAIConfig.TopP > -0.000001 {
		request.TopP = service.openAIConfig.TopP
	}
	response, err := service.client.CreateCompletion(context.TODO(), request)
	if err != nil {
		return openai.CompletionChoice{}, err
	}
	return response.Choices[0], nil
}

// sessionID目前打算是群ID/userID+uuid
// 发空消息可以更新初始人设的返回
func (service *OpenAIService) CreateChatCompletion(sessionID string, gptRole, message string) string {
	completionMsg := service.getChatSessionByID(sessionID)
	if message != "" {
		completionMsg = append(completionMsg, openai.ChatCompletionMessage{
			Role:    gptRole,
			Content: message,
		})
		completionMsg = service.ShortenChatToken(completionMsg, gptRole)
	}
	if len(completionMsg) < 1 {
		return "cmd response: 空白人设切换成功"
	}
	request := openai.ChatCompletionRequest{
		Model:    service.openAIConfig.Model,
		Messages: completionMsg,
		// MaxTokens: service.config.MaxToken,
	}
	if service.openAIConfig.Temperature > -0.000001 {
		request.Temperature = service.openAIConfig.Temperature
	}
	if service.openAIConfig.TopP > -0.000001 {
		request.TopP = service.openAIConfig.TopP
	}
	qingqiu := "消息体为：\n"
	for i := 0; i < len(completionMsg); i++ {
		qingqiu = fmt.Sprintf("%s%d. %s:%s\n", qingqiu, i+1, completionMsg[i].Role, completionMsg[i].Content)
	}
	log.Printf("请求内容为：%s", qingqiu)
	response, err := service.client.CreateChatCompletion(context.TODO(), request)
	if err != nil {
		return fmt.Sprintf("response error: %s", err)
	}
	completionMsg = append(completionMsg, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleAssistant,
		Content: response.Choices[0].Message.Content,
	})
	completionMsg = service.ShortenChatToken(completionMsg, openai.ChatMessageRoleAssistant)

	service.sessions[sessionID] = completionMsg
	return response.Choices[0].Message.Content
}

// 回滚上句对话，把上一个user对话删除，并删除user对话之后的assistant
func (service *OpenAIService) RollBackUserMessage(sessionID string, msg string) string {
	completionMsg := service.getChatSessionByID(sessionID)

	endflag := false
	newCompletionMsg := []openai.ChatCompletionMessage{}
	for i := len(completionMsg) - 1; i >= 0; i-- {
		if !endflag && completionMsg[i].Role == openai.ChatMessageRoleUser {
			endflag = true
			continue
		}
		if !endflag && completionMsg[i].Role == openai.ChatMessageRoleAssistant {
			continue
		}
		newCompletionMsg = append(newCompletionMsg, completionMsg[i])
	}
	completionMsg = []openai.ChatCompletionMessage{}
	for i := len(newCompletionMsg) - 1; i >= 0; i-- {
		completionMsg = append(completionMsg, newCompletionMsg[i])
	}
	if endflag {
		service.sessions[sessionID] = completionMsg
		return service.CreateChatCompletion(sessionID, openai.ChatMessageRoleUser, msg)
	}
	log.Printf("找不到第一条user类型的消息，无法回滚")
	return "cmd response: 找不到第一条user类型的消息，无法回滚"
}

// 回滚system角色的消息，如果下一条信息是assistant，就一并删除
func (service *OpenAIService) RollBackSystemMessage(sessionID string, msg string) string {
	completionMsg := service.getChatSessionByID(sessionID)
	if len(completionMsg) < 1 {
		return "cmd response: 系统现在没有设定，回滚不了"
	}

	endflag := false
	for i := len(completionMsg) - 1; i >= 0; i-- {
		if !endflag && completionMsg[i].Role == openai.ChatMessageRoleSystem {
			endflag = true
			startIndex := i
			endIndex := i + 1
			if i+1 < len(completionMsg) && completionMsg[i+1].Role == openai.ChatMessageRoleAssistant {
				endIndex = endIndex + 1
			}
			completionMsg = append(completionMsg[:startIndex], completionMsg[endIndex:]...)
			break
		}
	}
	if endflag {
		service.sessions[sessionID] = completionMsg
		return service.CreateChatCompletion(sessionID, openai.ChatMessageRoleSystem, msg)
	}
	log.Printf("找不到设定，无法回滚")
	return "cmd response: 找不到设定，无法回滚"
}

func (service *OpenAIService) ChangeRole(userID string, sessionID string) {
	service.userSessionNow[userID] = sessionID
}

func (service *OpenAIService) QueryRoleList() (string, []string) {
	list := []string{}
	for k := range service.roleConfig.RoleList {
		list = append(list, k)
	}
	return service.roleConfig.DefaultRole, list
}

// 减少某个role的聊天记录，用于节省token花费
func (service *OpenAIService) ShortenChatToken(messages []openai.ChatCompletionMessage, gptRole string) []openai.ChatCompletionMessage {
	n, i := service.returnNumAndFirstIndex(messages, gptRole)
	switch gptRole {
	case openai.ChatMessageRoleAssistant:
		if n > 0 && n > service.openAIConfig.AssistantMemory {
			messages = service.removeElement(messages, i)
		}
	case openai.ChatMessageRoleUser:
		if n > 0 && n > service.openAIConfig.UserMemory {
			messages = service.removeElement(messages, i)
		}
	case openai.ChatMessageRoleSystem:
		if n > 0 && n > service.openAIConfig.SystemMemory {
			messages = service.removeElement(messages, i)
		}
	}
	return messages
}

func (service *OpenAIService) returnNumAndFirstIndex(arr []openai.ChatCompletionMessage, gptRole string) (int, int) {
	n := 0
	s := 999
	for i := len(arr) - 1; i >= 0; i-- {
		if arr[i].Role == gptRole {
			n++
			s = i
		}
	}
	return n, s
}

func (service *OpenAIService) removeElement(arr []openai.ChatCompletionMessage, p int) []openai.ChatCompletionMessage {
	newArr := make([]openai.ChatCompletionMessage, 0, len(arr)-1)
	newArr = append(newArr, arr[:p]...)
	newArr = append(newArr, arr[p+1:]...)
	return newArr
}

func (service *OpenAIService) getChatSessionByID(sessionID string) []openai.ChatCompletionMessage {
	completionMsg, ok := service.sessions[sessionID]
	if !ok {
		completionMsg = []openai.ChatCompletionMessage{}
		roleNow := strings.Split(sessionID, "_")[1]
		detail2, ok2 := service.roleConfig.RoleList[roleNow]
		if !ok2 {
			// 如果不存在这个role的设定
			roleNow = service.roleConfig.DefaultRole
			detail3, ok3 := service.roleConfig.RoleList[roleNow]
			if !ok3 {
				// 如果连默认role都不存在
				detail3 = []string{}
			}
			detail2 = detail3
		}
		if len(detail2) < 1 || (len(detail2) == 1 && detail2[0] == "") {
			// 如果这个role设定存在，但是没有设定，就什么都不做
		} else {
			for i := 0; i < len(detail2); i++ {
				completionMsg = append(completionMsg, openai.ChatCompletionMessage{
					Role:    openai.ChatMessageRoleSystem,
					Content: detail2[i],
					Name:    "",
				})
			}
		}
	}
	return completionMsg
}

func (service *OpenAIService) getRoleNow(userID string) string {
	sessionID := service.userSessionNow[userID]
	if sessionID == "" {
		return service.roleConfig.DefaultRole
	}
	strs := strings.Split(sessionID, "_")
	return strs[1]
}

func (service *OpenAIService) ResetCurrentSession(sessionID string) string {
	delete(service.sessions, sessionID)
	msg := service.CreateChatCompletion(sessionID, openai.ChatMessageRoleUser, "")
	return fmt.Sprintf("cmd response: 重置成功\n%s", msg)
}

func (service *OpenAIService) getSessionIDByUser(userID string) string {
	sessionID := service.userSessionNow[userID]
	if sessionID == "" {
		sessionID = fmt.Sprintf("%s_%s", userID, service.getRoleNow(userID))
	}
	return sessionID
}

func (service *OpenAIService) checkSessionExist(sessionID string) bool {
	completionMsg, ok := service.sessions[sessionID]
	if ok && len(completionMsg) > 0 {
		return true
	}
	return false
}

func (service *OpenAIService) ClearUserMessage(sessionID string) string {
	completionMsg := service.getChatSessionByID(sessionID)
	newCompletionMsg := []openai.ChatCompletionMessage{}
	for i := 0; i < len(completionMsg); i++ {
		if completionMsg[i].Role == openai.ChatMessageRoleSystem {
			newCompletionMsg = append(newCompletionMsg, completionMsg[i])
		}
	}
	service.sessions[sessionID] = newCompletionMsg
	msg := service.CreateChatCompletion(sessionID, openai.ChatMessageRoleUser, "")
	return fmt.Sprintf("cmd response: 重置会话成功\n%s", msg)
}

func (service *OpenAIService) DescribeCurrentSession(sessionID string) []string {
	completionMsg := service.getChatSessionByID(sessionID)
	systemMsg := []string{}
	for i := 0; i < len(completionMsg); i++ {
		if completionMsg[i].Role == openai.ChatMessageRoleSystem {
			systemMsg = append(systemMsg, completionMsg[i].Content)
		}
	}
	return systemMsg
}
