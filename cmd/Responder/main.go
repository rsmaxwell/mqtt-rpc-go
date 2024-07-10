/*
 * Copyright (c) 2024 Contributors to the Eclipse Foundation
 *
 *  All rights reserved. This program and the accompanying materials
 *  are made available under the terms of the Eclipse Public License v2.0
 *  and Eclipse Distribution License v1.0 which accompany this distribution.
 *
 * The Eclipse Public License is available at
 *    https://www.eclipse.org/legal/epl-2.0/
 *  and the Eclipse Distribution License is available at
 *    http://www.eclipse.org/org/documents/edl-v10.php.
 *
 *  SPDX-License-Identifier: EPL-2.0 OR BSD-3-Clause
 */

/* see:
 *    https://github.com/eclipse/paho.golang/blob/v0.21.0/autopaho/examples/basics/basics.go
 *    https://github.com/eclipse/paho.golang/blob/master/autopaho/examples/rpc/main.go
 */

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"log"
	"net/url"
	"sync"
	"time"

	"github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/paho"
	"github.com/rsmaxwell/mqtt-rpc-go/internal/request"
	"github.com/rsmaxwell/mqtt-rpc-go/internal/response"
)

const qos = 0

type Handler interface {
	Handle(request.Request) (*response.Response, bool, error)
}

var (
	requestHandlers = map[string]Handler{
		"buildinfo":  new(BuildInfoHandler),
		"calculator": new(CalculatorHandler),
		"getPages":   new(GetPagesHandler),
		"quit":       new(QuitHandler),
	}
)

func main() {

	log.Printf("Responder")

	server := flag.String("server", "mqtt://127.0.0.1:1883", "The URL of the MQTT server")
	requestTopic := flag.String("rtopic", "request", "Topic for requests to go to")
	username := flag.String("username", "", "A username to authenticate to the MQTT server")
	password := flag.String("password", "", "Password to match username")
	flag.Parse()

	serverUrl, err := url.Parse(*server)
	if err != nil {
		panic(err)
	}

	var wg sync.WaitGroup
	wg.Add(1)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	config := autopaho.ClientConfig{
		ServerUrls:        []*url.URL{serverUrl},
		KeepAlive:         30,
		ConnectRetryDelay: 2 * time.Second,
		ConnectTimeout:    5 * time.Second,
		OnConnectError:    func(err error) { log.Printf("error whilst attempting connection: %s\n", err) },
		ClientConfig: paho.ClientConfig{
			OnClientError: func(err error) { log.Printf("requested disconnect: %s\n", err) },
			OnServerDisconnect: func(d *paho.Disconnect) {
				if d.Properties != nil {
					log.Printf("requested disconnect: %s\n", d.Properties.ReasonString)
				} else {
					log.Printf("requested disconnect; reason code: %d\n", d.ReasonCode)
				}
			},
		},
		ConnectUsername: *username,
		ConnectPassword: []byte(*password),
	}

	config.ClientConfig.ClientID = "listener"
	// Subscribing in OnConnectionUp is the recommended approach because this ensures the subscription is reestablished
	// following reconnection (the subscription should survive `cliCfg.SessionExpiryInterval` after disconnection,
	// but in this case that is 0, and it's safer if we don't assume the session survived anyway).
	config.OnConnectionUp = func(cm *autopaho.ConnectionManager, connAck *paho.Connack) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(5*time.Second))
		defer cancel()
		if _, err := cm.Subscribe(ctx, &paho.Subscribe{
			Subscriptions: []paho.SubscribeOptions{
				{Topic: *requestTopic, QoS: qos},
			},
		}); err != nil {
			log.Printf("listener failed to subscribe (%s). This is likely to mean no messages will be received.", err)
			return
		}
	}
	config.OnPublishReceived = []func(paho.PublishReceived) (bool, error){
		func(received paho.PublishReceived) (bool, error) {
			if received.Packet.Properties != nil && received.Packet.Properties.CorrelationData != nil && received.Packet.Properties.ResponseTopic != "" {
				log.Printf("Received request: %s", string(received.Packet.Payload))

				var req request.Request

				if err := json.NewDecoder(bytes.NewReader(received.Packet.Payload)).Decode(&req); err != nil {
					log.Printf("discarding request because message could not be decoded: %v", err)
				}

				handler := requestHandlers[req.Function]
				if handler == nil {
					log.Fatalf("discarding request because handler not found: %s", req.Function)
				}

				resp, quit, err := handler.Handle(req)
				if err != nil {
					log.Fatalf("discarding request because handler '%s' failed: %s", req.Function, err)
				}

				body, _ := json.Marshal(resp)
				log.Printf("Sending reply: %s", body)

				_, err = received.Client.Publish(ctx, &paho.Publish{
					Properties: &paho.PublishProperties{
						CorrelationData: received.Packet.Properties.CorrelationData,
					},
					Topic:   received.Packet.Properties.ResponseTopic,
					Payload: body,
				})
				if err != nil {
					log.Fatalf("failed to publish response: %s", err)
				}

				if quit {
					wg.Done()
				}
			}
			return true, nil
		}}

	_, err = autopaho.NewConnection(ctx, config)
	if err != nil {
		panic(err)
	}

	// Wait till asked to quit
	wg.Wait()
	log.Printf("Quitting")
}
