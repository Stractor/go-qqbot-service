package controller

import "example.com/m/v2/src/service"

type Controller struct {
	wssService *service.WssService
}

func NewController(
	wss *service.WssService,
) *Controller {
	return &Controller{
		wssService: wss,
	}
}

func (ctr *Controller) Start() {
	ctr.wssService.Start()
}
