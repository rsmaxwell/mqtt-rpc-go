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
	"context"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/autopaho/extensions/rpc"
	"github.com/eclipse/paho.golang/paho"
)

const qos = 0

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	server := flag.String("server", "mqtt://127.0.0.1:1883", "The full URL of the MQTT server to connect to")
	rTopic := flag.String("rtopic", "request", "Topic for requests to go to")
	username := flag.String("username", "", "A username to authenticate to the MQTT server")
	password := flag.String("password", "", "Password to match username")
	flag.Parse()

	serverUrl, err := url.Parse(*server)
	if err != nil {
		panic(err)
	}

	genericConfig := autopaho.ClientConfig{
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

	err = remoteProcedureCall(ctx, genericConfig, rTopic)
	if err != nil {
		log.Fatal(err)
	}

	topic := "diaries/pages"
	cm, err := subscribeToTopic(ctx, genericConfig, topic)
	if err != nil {
		log.Fatal(err)
	}
	defer unsubscribeFromTopic(ctx, cm, topic)

	// Wait a while
	time.Sleep(10 * time.Second)
}

func remoteProcedureCall(ctx context.Context, genericConfig autopaho.ClientConfig, rTopic *string) error {

	config := genericConfig
	config.ClientConfig.ClientID = "rpc-requestor"

	initialSubscriptionMade := make(chan struct{}) // Closed when subscription made (otherwise we might send request before subscription in place)
	var initialSubscriptionOnce sync.Once          // We only want to close the above once!

	config.OnConnectionUp = func(cm *autopaho.ConnectionManager, connAck *paho.Connack) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(5*time.Second))
		defer cancel()
		if _, err := cm.Subscribe(ctx, &paho.Subscribe{
			Subscriptions: []paho.SubscribeOptions{
				{Topic: fmt.Sprintf("response/%s", config.ClientID), QoS: qos},
			},
		}); err != nil {
			log.Printf("requestor failed to subscribe (%s). This is likely to mean no messages will be received.", err)
			return
		}
		initialSubscriptionOnce.Do(func() { close(initialSubscriptionMade) })
	}

	router := paho.NewStandardRouter()
	config.OnPublishReceived = []func(paho.PublishReceived) (bool, error){
		func(p paho.PublishReceived) (bool, error) {
			router.Route(p.Packet.Packet())
			return false, nil
		}}

	cm, err := autopaho.NewConnection(ctx, config)
	if err != nil {
		panic(err)
	}

	// Wait for the subscription to be made (otherwise we may miss the response!)
	connCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	select {
	case <-connCtx.Done():
		log.Fatalf("requestor failed to connect & subscribe: %s", err)
	case <-initialSubscriptionMade:
	}

	h, err := rpc.NewHandler(ctx, rpc.HandlerOpts{
		Conn:             cm,
		Router:           router,
		ResponseTopicFmt: "response/%s",
		ClientID:         config.ClientID,
	})

	if err != nil {
		log.Fatal(err)
	}

	request := `{"function":"calculator", "args": { "operation":"mul", "param1": 10, "param2": 5 } }`
	log.Printf("Sending request: %s\n", request)
	resp, err := h.Request(ctx, &paho.Publish{
		Topic:   *rTopic,
		Payload: []byte(request),
	})
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Received response: %s", string(resp.Payload))

	return nil
}

func subscribeToTopic(ctx context.Context, genericConfig autopaho.ClientConfig, topic string) (*autopaho.ConnectionManager, error) {

	log.Printf("Subscribing to: %s\n", topic)

	config := genericConfig

	initialSubscriptionMade := make(chan struct{}) // Closed when subscription made (otherwise we might send request before subscription in place)
	var initialSubscriptionOnce sync.Once          // We only want to close the above once!

	config.OnConnectionUp = func(cm *autopaho.ConnectionManager, connAck *paho.Connack) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(5*time.Second))
		defer cancel()
		if _, err := cm.Subscribe(ctx, &paho.Subscribe{
			Subscriptions: []paho.SubscribeOptions{
				{Topic: topic, QoS: qos},
			},
		}); err != nil {
			log.Printf("requestor failed to subscribe (%s). This is likely to mean no messages will be received.", err)
			return
		}
		initialSubscriptionOnce.Do(func() { close(initialSubscriptionMade) })
	}

	config.ClientConfig = paho.ClientConfig{
		// If you are using QOS 1/2, then it's important to specify a client id (which must be unique)
		ClientID: "subscriber",
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

	router := paho.NewStandardRouter()
	router.RegisterHandler(topic, func(p *paho.Publish) {
		log.Printf("'%s' received message with topic '%s'; body: %s (retain: %t)\n", topic, p.Topic, p.Payload, p.Retain)
	})
	router.DefaultHandler(func(p *paho.Publish) {
		log.Printf("defaulthandler received message with topic %s; body: %s (retain: %t)\n", p.Topic, p.Payload, p.Retain)
	})

	config.OnPublishReceived = []func(paho.PublishReceived) (bool, error){
		func(p paho.PublishReceived) (bool, error) {
			router.Route(p.Packet.Packet())
			return true, nil
		}}

	c, err := autopaho.NewConnection(ctx, config) // starts process; will reconnect until context cancelled
	if err != nil {
		panic(err)
	}

	// Wait for the connection to come up
	if err = c.AwaitConnection(ctx); err != nil {
		panic(err)
	}

	return c, nil
}

func unsubscribeFromTopic(ctx context.Context, cm *autopaho.ConnectionManager, topic string) error {

	log.Printf("Unsubscribing from: %s\n", topic)

	cm.Unsubscribe(ctx, &paho.Unsubscribe{
		Topics: []string{
			topic,
		},
	})

	return nil
}
