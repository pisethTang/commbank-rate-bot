package handlers

import (
	"conversion-bot/viewmodels"
)

type ApiHandler struct {
	APiList viewmodels.APIs
}

func (h ApiHandler) GetAPIs() *viewmodels.APIs {
	return &h.APiList
}