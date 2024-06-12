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
	"fmt"
	"log"
	"net/url"
	"sync"
	"time"

	"github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/paho"

	"github.com/rsmaxwell/diaries/server-go/cmd/internal/request"
)

const qos = 0

type Handler func(*map[string]interface{}) (interface{}, error)

var (
	handlers = map[string]Handler{
		"calculator": calculator,
		"getPages":   getPages,
	}
)

func publisher(ctx context.Context, config autopaho.ClientConfig) error {
	log.Printf("Publishing to 'diaries/pages'")

	config.KeepAlive = 20 // Keepalive message should be sent every 20 seconds
	// CleanStartOnInitialConnection defaults to false. Setting this to true will clear the session on the first connection.
	config.CleanStartOnInitialConnection = false
	// SessionExpiryInterval - Seconds that a session will survive after disconnection.
	// It is important to set this because otherwise, any queued messages will be lost if the connection drops and
	// the server will not queue messages while it is down. The specific setting will depend upon your needs
	// (60 = 1 minute, 3600 = 1 hour, 86400 = one day, 0xFFFFFFFE = 136 years, 0xFFFFFFFF = don't expire)
	config.SessionExpiryInterval = 60
	config.OnConnectionUp = func(cm *autopaho.ConnectionManager, connAck *paho.Connack) {
		log.Println("mqtt connection up")
		// Subscribing in the OnConnectionUp callback is recommended (ensures the subscription is reestablished if
		// the connection drops)
		topic := "diaries/pages"
		if _, err := cm.Subscribe(context.Background(), &paho.Subscribe{
			Subscriptions: []paho.SubscribeOptions{
				{Topic: topic, QoS: 1},
			},
		}); err != nil {
			log.Printf("failed to subscribe (%s). This is likely to mean no messages will be received.", err)
		}
		log.Printf("Subscribed to: %s\n", topic)
	}
	config.OnConnectError = func(err error) { log.Printf("error whilst attempting connection: %s\n", err) }
	// eclipse/paho.golang/paho provides base mqtt functionality, the below config will be passed in for each connection
	config.ClientConfig = paho.ClientConfig{
		// If you are using QOS 1/2, then it's important to specify a client id (which must be unique)
		ClientID: "publisher",
		// OnPublishReceived is a slice of functions that will be called when a message is received.
		// You can write the function(s) yourself or use the supplied Router
		OnPublishReceived: []func(paho.PublishReceived) (bool, error){
			func(pr paho.PublishReceived) (bool, error) {
				// log.Printf("received message on topic %s; body: %s (retain: %t)\n", pr.Packet.Topic, pr.Packet.Payload, pr.Packet.Retain)
				return true, nil
			}},
		OnClientError: func(err error) { log.Printf("client error: %s\n", err) },
		OnServerDisconnect: func(d *paho.Disconnect) {
			if d.Properties != nil {
				log.Printf("server requested disconnect: %s\n", d.Properties.ReasonString)
			} else {
				log.Printf("server requested disconnect; reason code: %d\n", d.ReasonCode)
			}
		},
	}
	config.CleanStartOnInitialConnection = false

	c, err := autopaho.NewConnection(ctx, config) // starts process; will reconnect until context cancelled
	if err != nil {
		panic(err)
	}
	// Wait for the connection to come up
	if err = c.AwaitConnection(ctx); err != nil {
		panic(err)
	}

	_, err = c.Publish(ctx, &paho.Publish{
		Topic:   "diaries/pages",
		Payload: []byte(`[ "one", "two", "three" ]`),
		Retain:  true,
	})
	if err != nil {
		log.Fatal(err)
	}

	return nil
}

func listener(ctx context.Context, config autopaho.ClientConfig, topic string, qos byte) {
	initialSubscriptionMade := make(chan struct{}) // Closed when subscription made (otherwise we might send request before subscription in place)
	var initialSubscriptionOnce sync.Once          // We only want to close the above once!

	config.ClientConfig.ClientID = "rpc-listener"
	// Subscribing in OnConnectionUp is the recommended approach because this ensures the subscription is reestablished
	// following reconnection (the subscription should survive `cliCfg.SessionExpiryInterval` after disconnection,
	// but in this case that is 0, and it's safer if we don't assume the session survived anyway).
	config.OnConnectionUp = func(cm *autopaho.ConnectionManager, connAck *paho.Connack) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(5*time.Second))
		defer cancel()
		if _, err := cm.Subscribe(ctx, &paho.Subscribe{
			Subscriptions: []paho.SubscribeOptions{
				{Topic: topic, QoS: qos},
			},
		}); err != nil {
			log.Printf("listener failed to subscribe (%s). This is likely to mean no messages will be received.", err)
			return
		}
		initialSubscriptionOnce.Do(func() { close(initialSubscriptionMade) })
	}
	config.OnPublishReceived = []func(paho.PublishReceived) (bool, error){
		func(received paho.PublishReceived) (bool, error) {
			if received.Packet.Properties != nil && received.Packet.Properties.CorrelationData != nil && received.Packet.Properties.ResponseTopic != "" {
				log.Printf("Received message with response topic %s and correl id %s\n%s", received.Packet.Properties.ResponseTopic, string(received.Packet.Properties.CorrelationData), string(received.Packet.Payload))

				var r request.Request

				if err := json.NewDecoder(bytes.NewReader(received.Packet.Payload)).Decode(&r); err != nil {
					log.Printf("Failed to decode Request: %v", err)
				}

				handler := handlers[r.Function]
				if handler == nil {
					log.Fatalf("Handler not found: %s", r.Function)
				}

				result, err := handler(r.Args)
				if err != nil {
					log.Fatalf("handler '%s' failed: %s", r.Function, err)
				}

				body, _ := json.Marshal(result)
				_, err = received.Client.Publish(ctx, &paho.Publish{
					Properties: &paho.PublishProperties{
						CorrelationData: received.Packet.Properties.CorrelationData,
					},
					Topic:   received.Packet.Properties.ResponseTopic,
					Payload: body,
				})
				if err != nil {
					log.Fatalf("failed to publish message: %s", err)
				}
			}
			return true, nil
		}}

	_, err := autopaho.NewConnection(ctx, config)
	if err != nil {
		panic(err)
	}

	// Connection must be up, and subscription made, within a reasonable time period.
	// In a real app you would probably not wait for the subscription, but it's important here because otherwise the
	// request could be sent before the subscription is in place.
	connCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	select {
	case <-connCtx.Done():
		log.Fatalf("listener failed to connect & subscribe: %s", err)
	case <-initialSubscriptionMade:
	}

	// Wait forever, or until ctrl-C
	for {
		time.Sleep(time.Duration(1<<63 - 1))
	}
}

func main() {
	// ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	// defer cancel()

	ctx := context.Background()

	server := flag.String("server", "mqtt://127.0.0.1:1883", "The full URL of the MQTT server to connect to")
	requestTopic := flag.String("rtopic", "rpc/request", "Topic for requests to go to")
	username := flag.String("username", "", "A username to authenticate to the MQTT server")
	password := flag.String("password", "", "Password to match username")
	flag.Parse()

	serverUrl, err := url.Parse(*server)
	if err != nil {
		panic(err)
	}

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

	publisher(ctx, config)

	listener(ctx, config, *requestTopic, qos) // Start the listener (this will respond to requests)
}

func GetStringArgument(key string, data *map[string]interface{}) (string, error) {

	object, ok := (*data)[key]
	if !ok {
		return "", fmt.Errorf("could not find the key [%s]", key)
	}
	value, ok := object.(string)
	if !ok {
		return "", fmt.Errorf("unexpected type for the key [%s]: %#v", key, object)
	}

	return value, nil
}

func GetIntegerArgument(key string, data *map[string]interface{}) (int, error) {

	object, ok := (*data)[key]
	if !ok {
		return 0, fmt.Errorf("could not find the key [%s]", key)
	}

	value, ok := object.(float64)
	if !ok {
		return 0, fmt.Errorf("unexpected type for the key [%s]: %#v", key, object)
	}

	return int(value), nil
}

func calculator(args *map[string]interface{}) (interface{}, error) {

	operation, err := GetStringArgument("operation", args)
	if err != nil {
		log.Fatalf("could not find 'operation' in arguments: %s", err)
	}

	param1, err := GetIntegerArgument("param1", args)
	if err != nil {
		log.Fatalf("could not find 'param1' in arguments: %s", err)
	}

	param2, err := GetIntegerArgument("param2", args)
	if err != nil {
		log.Fatalf("could not find 'param2' in arguments: %s", err)
	}

	value := 0

	switch operation {
	case "add":
		value = param1 + param2
	case "mul":
		value = param1 * param2
	case "div":
		value = param1 / param2
	case "sub":
		value = param1 - param2
	}

	result := interface{}(value)
	return result, nil
}

func getPages(args *map[string]interface{}) (interface{}, error) {
	value := "[ 'one', 'two', 'three' ]"
	result := interface{}(value)
	return result, nil
}
