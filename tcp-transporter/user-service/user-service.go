package main

import (
	"errors"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/moleculer-go/moleculer"
	"github.com/moleculer-go/moleculer/broker"
	"github.com/moleculer-go/moleculer/payload"
)

type UserService struct {
	profileCreated chan bool
	OnPanix        func(moleculer.Context)
}

func (s *UserService) Name() string {
	return "user"
}

func (s *UserService) Dependencies() []string {
	return []string{"profile"}
}

var userStore = make(map[string]moleculer.Payload)

func (s *UserService) Create(ctx moleculer.Context, user moleculer.Payload) moleculer.Payload {
	ctx.Logger().Info("user.create called! - user: ", user)
	user = user.Add("id", rand.Intn(1000000))
	ctx.Emit("user.created", user)
	userStore[user.Get("id").String()] = user
	return user
}

func (s *UserService) Get(ctx moleculer.Context, userId moleculer.Payload) moleculer.Payload {
	ctx.Logger().Info("user.get called! - userId: ", userId.Get("id").String())
	return userStore[userId.Get("id").String()]
}

func (s *UserService) Update(ctx moleculer.Context, user moleculer.Payload) moleculer.Payload {
	ctx.Logger().Info("user.update called! - user: ", user)
	user = user.Add("updated", time.Now())
	ctx.Emit("user.updated", user)
	userStore[user.Get("id").String()] = user
	return user
}

func (s *UserService) Panix(ctx moleculer.Context, params moleculer.Payload) moleculer.Payload {
	ctx.Logger().Info("user.panix called! ")
	s.OnPanix(ctx)

	panic("this action will panic!")
}

func (s *UserService) Fail(ctx moleculer.Context) interface{} {
	ctx.Logger().Info("user.fail called! ")
	return errors.New("this actions returns an error!")
}

func (s *UserService) BulkUpdate(ctx moleculer.Context) moleculer.Payload {
	params := ctx.Payload()
	data := params.Get("data").Array()
	action := params.Get("action").String()
	ctx.Logger().Debugf("user.bulkUpdate action: %s data.length: %d\n", action, len(data))
	var result []moleculer.Payload
	if action == "divide" {
		half := len(data) / 2
		seen := make(map[interface{}]bool)
		for len(result) < half {
			randomIndex := rand.Intn(len(data))
			randomItem := data[randomIndex]
			if !seen[randomItem] {
				result = append(result, randomItem)
				seen[randomItem] = true
				<-ctx.Call("account.update", randomItem)
			}
		}
	} else if action == "multiply" {
		count := make(map[interface{}]int)
		for len(result) < len(data)*2 {
			randomIndex := rand.Intn(len(data))
			randomItem := data[randomIndex]
			if count[randomItem] < 2 {
				result = append(result, randomItem)
				count[randomItem]++
				<-ctx.Call("account.update", randomItem)
			}
		}
	}
	ctx.Emit("user.bulkUpdated", payload.Empty().Add("size", len(result)).Add("action", action))
	return payload.New(result)
}

func (s *UserService) Events() []moleculer.Event {
	return []moleculer.Event{
		{
			Name: "profile.loopevent",
			Handler: func(ctx moleculer.Context, params moleculer.Payload) {
				ctx.Logger().Info("profile.loopevent arrived: ", params)
			},
		},
		{
			Name: "profile.created",
			Handler: func(ctx moleculer.Context, profile moleculer.Payload) {
				ctx.Logger().Info("profile.created event! profile: ", profile)
				user := map[string]interface{}{
					"id":        profile.Get("user").Get("id").String(),
					"profileId": profile.Get("id").String(),
				}
				<-ctx.Call("user.update", user)
				ctx.Logger().Info("user updated with profile Id :) ")

				go func() {
					s.profileCreated <- true
				}()
			},
		},
	}
}

func main() {
	bkr := broker.New(&moleculer.Config{
		Transporter:                "TCP",
		WaitForDependenciesTimeout: 10 * time.Second,
		LogLevel:                   "DEBUG",
		DiscoverNodeID: func() string {
			return "user-service-node"
		}})

	bkr.Publish(&UserService{profileCreated: make(chan bool)})
	bkr.Start()

	// Create a channel to receive OS signals
	sig := make(chan os.Signal, 1)
	// Notify the channel for SIGINT and SIGTERM signals
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	// Wait for a signal
	<-sig
	// Stop the broker after receiving the signal
	bkr.Stop()
}
