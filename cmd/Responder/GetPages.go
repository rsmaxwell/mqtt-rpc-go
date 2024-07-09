package main

import (
	"log"
	"sync"

	"github.com/rsmaxwell/mqtt-rpc-go/internal/request"
	"github.com/rsmaxwell/mqtt-rpc-go/internal/response"
)

type GetPages struct {
}

func (h *GetPages) Handle(wg *sync.WaitGroup, req request.Request) (*response.Response, bool, error) {
	log.Printf("getPages")

	resp := response.Success()
	resp.PutString("result", "[ 'one', 'two', 'three' ]")
	return resp, false, nil
}
