package service

import (
	"context"
	"log"

	openai "github.com/sashabaranov/go-openai"
)

type OpenAIService struct {
	client         *openai.Client
	config         *OpenAIConfig
	sessions       map[string][]openai.ChatCompletionMessage
	roleNow        string
	roleList       map[string]string
	totalUsageChat map[string]map[string]int
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
	Model       string
	Token       string
	RoleList    map[string]string
	MaxToken    int
	Temperature float32
	TopP        int
}

func (service *OpenAIService) ListModel() ([]openai.Model, error) {
	modelList, err := service.client.ListModels(context.TODO())
	if err != nil {
		return []openai.Model{}, err
	}
	return modelList.Models, nil
}

// 根据prompt续写
func (service *OpenAIService) CreateCompletion(prompt string, maxToken int,
	temperature float32, topP float32) (openai.CompletionChoice, error) {
	response, err := service.client.CreateCompletion(context.TODO(), openai.CompletionRequest{
		Model:       service.config.Model,
		Prompt:      prompt,
		MaxTokens:   maxToken,
		Temperature: temperature,
		TopP:        topP,
	})
	if err != nil {
		return openai.CompletionChoice{}, err
	}
	return response.Choices[0], nil
}

// sessionID目前打算是群ID/userID+uuid
func (service *OpenAIService) CreateChatCompletion(sessionID string, message string) (string, error) {
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
	response, err := service.client.CreateChatCompletion(context.TODO(), openai.ChatCompletionRequest{
		Model:       service.config.Model,
		Messages:    completionMsg,
		MaxTokens:   service.config.MaxToken,
		Temperature: service.config.Temperature,
		TopP:        float32(service.config.TopP),
	})
	if err != nil {
		return "", err
	}
	log.Printf("prompt_token:%d, completion_token:%d, total_token:%d", response.Usage.PromptTokens, response.Usage.CompletionTokens, response.Usage.TotalTokens)
	completionMsg = append(completionMsg, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleAssistant,
		Content: response.Choices[0].Message.Content,
	})
	service.RemoveUntilMaxTokenShort(sessionID, response.Usage.PromptTokens, response.Usage.CompletionTokens)
	service.sessions[sessionID] = completionMsg
	return response.Choices[0].Message.Content, nil
}

func (service *OpenAIService) RoleBackMessage(sessionID string) (string, error) {
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
		return "你们还没对话过，不用回滚", nil
	}
	index := 1
	length := len(completionMsg)
	for ; index < length; index++ {

		str := completionMsg[length-index : length-index+1]
		if str[0].Role == openai.ChatMessageRoleUser {
			completionMsg = completionMsg[:length-index]
			service.sessions[sessionID] = completionMsg
			return service.CreateChatCompletion(sessionID, str[0].Content)
		}
	}
	log.Printf("找不到第一条user类型的消息")
	return "会话出错，请重试", nil
}

func (service *OpenAIService) ChangeRole(name string, userID string) {
	service.roleNow = name
}

func (service *OpenAIService) QueryRoleList(userID string) map[string]string {
	return service.roleList
}

func (service *OpenAIService) RemoveUntilMaxTokenShort(sessionID string, prompt, completion int) {
	// // fixme: 用量要重新计算，对接
	// usages, ok := service.totalUsageChat[sessionID]
	// if !ok {
	// 	usages = map[string]int{
	// 		"prompt":     prompt,
	// 		"completion": completion,
	// 		"total":      prompt + completion,
	// 	}
	// 	service.totalUsageChat[sessionID] = usages
	// }
	// if usages["total"] < service.config.MaxToken {
	// 	return
	// }
	// completionMsg, ok := service.sessions[sessionID]
	// if !ok {
	// 	log.Fatalf("出错了，检查一下:%s %d:%d", sessionID, prompt, completion)
	// }
	// count := 3
	// newCompletionMsg := []openai.ChatCompletionMessage{}
	// for _, v := range completionMsg {
	// 	if v.Role != openai.ChatMessageRoleSystem && count > 0 {
	// 		continue
	// 	}
	// 	newCompletionMsg = append(newCompletionMsg, v)
	// }
}
