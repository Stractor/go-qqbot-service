package main

import (
	"fmt"

	"example.com/m/v2/src/controller"
	"example.com/m/v2/src/service"
	"go.uber.org/dig"
)

func main() {
	print("main start")

	container := dig.New()
	service.BindConfig(container, "Wss", &controller.WssConfig{})
	err := service.BindConfig(container, "OpenAI", &service.OpenAIConfig{})
	if err != nil {
		fmt.Printf("bind config error: %s", err)
		return
	}

	container.Provide(controller.NewController)
	container.Provide(service.NewOpenAIService)

	err = container.Invoke(func(ctr *controller.Controller) {
		ctr.Start()
	})
	if err != nil {
		fmt.Printf("invoke error: %s", err)
	}
}
