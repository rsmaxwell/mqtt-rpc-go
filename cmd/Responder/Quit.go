package main

import (
	"fmt"
	"log"
	"sync"

	"github.com/rsmaxwell/mqtt-rpc-go/internal/request"
	"github.com/rsmaxwell/mqtt-rpc-go/internal/response"
)

type Quit struct {
}

func (h *Quit) Handle(wg *sync.WaitGroup, req request.Request) (*response.Response, bool, error) {
	log.Printf("quit")

	quit, err := req.GetBoolean("quit")
	if err != nil {
		resp := response.BadRequest()
		resp.PutMessage(fmt.Sprintf("could not find 'quit' in arguments: %s", err))
		return resp, false, nil
	}

	resp := response.Success()
	return resp, quit, nil
}
