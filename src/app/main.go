package main

import (
	"fmt"

	"example.com/m/v2/src/controller"
	"example.com/m/v2/src/service"
	"go.uber.org/dig"
)

func main() {
	println("main start")

	container := dig.New()
	service.BindConfig(container, "Wss", &controller.WssConfig{})
	service.BindConfig(container, "OpenAI", &service.OpenAIConfig{})
	service.BindConfig(container, "QQBot", &service.QQBotActionConfig{})
	service.BindPrivicyConfig(container, "OpenAI", &service.PrivacyConfig{})
	err := service.BindPrivicyConfig(container, "Role", &service.RoleConfig{})
	if err != nil {
		fmt.Printf("bind config error: %s", err)
		return
	}

	container.Provide(controller.NewController)
	container.Provide(service.NewOpenAIService)
	container.Provide(service.NewBotActionService)

	err = container.Invoke(func(ctr *controller.Controller) {
		ctr.Start()
	})
	if err != nil {
		fmt.Printf("invoke error: %s", err)
	}
}
