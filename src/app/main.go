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
	err := service.BindConfig(container, "Wss", &service.WssConfig{})
	if err != nil {
		fmt.Printf("bind config error: %s", err)
		return
	}

	container.Provide(service.NewWssService)

	container.Provide(controller.NewController)

	err = container.Invoke(func(ctr *controller.Controller) {
		ctr.Start()
	})
	if err != nil {
		fmt.Printf("invoke error: %s", err)
	}
}
