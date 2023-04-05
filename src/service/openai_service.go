package service

import (
	"context"
	"log"

	openai "github.com/sashabaranov/go-openai"
)

type OpenAIService struct {
	client   *openai.Client
	config   *OpenAIConfig
	sessions map[string][]openai.ChatCompletionMessage
	roleNow  string
	roleList map[string]string
}

func NewOpenAIService(config *OpenAIConfig) *OpenAIService {
	client := openai.NewClient(config.Token)

	sessions := map[string][]openai.ChatCompletionMessage{}
	return &OpenAIService{
		client:   client,
		config:   config,
		sessions: sessions,
		roleNow:  "",
		roleList: config.RoleList,
	}
}

type OpenAIConfig struct {
	Model           string
	Token           string
	RoleList        map[string]string
	MaxToken        int
	Temperature     float32
	TopP            float32
	UserMemory      int
	AssistantMemory int
	SystemMemory    int
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
		Model:     service.config.Model,
		Prompt:    prompt,
		MaxTokens: service.config.MaxToken,
	}
	if service.config.Temperature > -0.000001 {
		request.Temperature = service.config.Temperature
	}
	if service.config.TopP > -0.000001 {
		request.TopP = service.config.TopP
	}
	response, err := service.client.CreateCompletion(context.TODO(), request)
	if err != nil {
		return openai.CompletionChoice{}, err
	}
	return response.Choices[0], nil
}

// sessionID目前打算是群ID/userID+uuid
// 发空消息可以更新初始人设的返回
func (service *OpenAIService) CreateChatCompletion(sessionID string, gptRole, message string) (string, error) {
	completionMsg, ok := service.sessions[sessionID]
	if !ok {
		completionMsg = []openai.ChatCompletionMessage{}
		if service.roleNow != "" {
			if _, ok := service.roleList[service.roleNow]; ok {
				completionMsg = append(completionMsg, openai.ChatCompletionMessage{
					Role:    openai.ChatMessageRoleSystem,
					Content: service.roleList[service.roleNow],
					Name:    "",
				})
			}
		}
	}
	if message != "" {
		completionMsg = append(completionMsg, openai.ChatCompletionMessage{
			Role:    gptRole,
			Content: message,
		})
		completionMsg = service.ShortenChatToken(completionMsg, gptRole)
	}
	if len(completionMsg) < 1 {
		return "cmd response: 请发消息", nil
	}
	request := openai.ChatCompletionRequest{
		Model:     service.config.Model,
		Messages:  completionMsg,
		MaxTokens: service.config.MaxToken,
	}
	if service.config.Temperature > -0.000001 {
		request.Temperature = service.config.Temperature
	}
	if service.config.TopP > -0.000001 {
		request.TopP = service.config.TopP
	}
	response, err := service.client.CreateChatCompletion(context.TODO(), request)
	if err != nil {
		return "", err
	}
	completionMsg = append(completionMsg, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleAssistant,
		Content: response.Choices[0].Message.Content,
	})
	completionMsg = service.ShortenChatToken(completionMsg, openai.ChatMessageRoleAssistant)

	service.sessions[sessionID] = completionMsg
	return response.Choices[0].Message.Content, nil
}

// 回滚上句对话，把上一个user对话删除，并删除user对话之后的assistant
func (service *OpenAIService) RoleBackUserMessage(sessionID string, msg string) (string, error) {
	completionMsg, ok := service.sessions[sessionID]
	if !ok {
		completionMsg = []openai.ChatCompletionMessage{}
		if service.roleNow != "" {
			if _, ok := service.roleList[service.roleNow]; ok {
				completionMsg = append(completionMsg, openai.ChatCompletionMessage{
					Role:    openai.ChatMessageRoleSystem,
					Content: service.roleList[service.roleNow],
					Name:    "",
				})
			}
		}
		service.sessions[sessionID] = completionMsg
	}

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
	if endflag {
		service.sessions[sessionID] = newCompletionMsg
		return service.CreateChatCompletion(sessionID, openai.ChatMessageRoleUser, msg)
	}
	log.Printf("找不到第一条user类型的消息，无法回滚")
	return "cmd response: 找不到第一条user类型的消息，无法回滚", nil
}

// 回滚system角色的消息
func (service *OpenAIService) RoleBackSystemMessage(sessionID string, msg string) (string, error) {
	completionMsg, ok := service.sessions[sessionID]
	if !ok || len(completionMsg) < 1 {
		return "cmd response: 系统现在没有人设", nil
	}

	endflag := false
	for i := len(completionMsg) - 1; i >= 0; i-- {
		if !endflag && completionMsg[i].Role == openai.ChatMessageRoleSystem {
			endflag = true
			service.removeElement(completionMsg, i)
			break
		}
	}
	if endflag {
		service.sessions[sessionID] = completionMsg
		return service.CreateChatCompletion(sessionID, openai.ChatMessageRoleSystem, msg)
	}
	log.Printf("找不到设定，无法回滚")
	return "cmd response: 找不到设定，无法回滚", nil
}

func (service *OpenAIService) ChangeRole(name string, userID string) {
	service.roleNow = name
}

func (service *OpenAIService) QueryRoleList(userID string) map[string]string {
	return service.roleList
}

// 减少某个role的聊天记录，用于节省token花费
func (service *OpenAIService) ShortenChatToken(messages []openai.ChatCompletionMessage, gptRole string) []openai.ChatCompletionMessage {
	n, i := service.returnNumAndFirstIndex(messages, gptRole)
	switch gptRole {
	case openai.ChatMessageRoleAssistant:
		if n > 0 && n > service.config.AssistantMemory {
			messages = service.removeElement(messages, i)
		}
	case openai.ChatMessageRoleUser:
		if n > 0 && n > service.config.UserMemory {
			messages = service.removeElement(messages, i)
		}
	case openai.ChatMessageRoleSystem:
		if n > 0 && n > service.config.SystemMemory {
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
