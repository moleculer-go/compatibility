package main

import (
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/moleculer-go/moleculer"
	"github.com/moleculer-go/moleculer/broker"
	"github.com/moleculer-go/moleculer/payload"
)

type DataService struct {
	inputMessages []moleculer.Payload
}

func (s *DataService) Name() string {
	return "data"
}

func (s *DataService) Create(ctx moleculer.Context, user moleculer.Payload) moleculer.Payload {
	ctx.Logger().Info("user.create called! - user: ", user)
	ctx.Emit("user.created", user)
	return user
}

func (s *DataService) Get(ctx moleculer.Context, user moleculer.Payload) moleculer.Payload {
	ctx.Logger().Info("user.get called! - user: ", user)
	return user
}

var lastAction = "multiply"

func (s *DataService) Events() []moleculer.Event {
	return []moleculer.Event{
		{
			Name: "profile.created",
			Handler: func(ctx moleculer.Context, profile moleculer.Payload) {
				ctx.Logger().Info("profile.created event! profile: ", profile)
				syntheticData := createSyntheticData(ctx, profile)
				ctx.Logger().Info("account.bulkUpdate syntheticData size: ", len(syntheticData))

				action := "multiply"
				if lastAction == "multiply" {
					action = "divide"
				}
				lastAction = action

				startTime := time.Now()
				result := <-ctx.Call("account.bulkUpdate", payload.Empty().Add("data:", syntheticData).Add("action", action))
				ctx.Logger().Info("account.bulkUpdate result size: ", result.Len())
				if result.IsError() {
					ctx.Logger().Error("Not expected Error -> account.bulkUpdate error: ", result.Error())
					return
				}
				ctx.Logger().Debug("account.bulkUpdate took: ", time.Since(startTime))

				startTime = time.Now()
				result = <-ctx.Call("user.bulkUpdate", payload.Empty().Add("data:", syntheticData).Add("action", action))
				ctx.Logger().Info("user.bulkUpdate result size: ", result.Len())
				if result.IsError() {
					ctx.Logger().Error("Not expected Error -> account.bulkUpdate error: ", result.Error())
					return
				}
				ctx.Logger().Debug("user.bulkUpdate took: ", time.Since(startTime))

				startTime = time.Now()
				ctx.Logger().Debug("will call account.bulkUpdate and user.bulkUpdate for each record in syntheticData")
				for _, data := range syntheticData {
					<-ctx.Call("account.bulkUpdate", payload.Empty().Add("data:", []moleculer.Payload{data}).Add("action", "multiply"))
					<-ctx.Call("user.bulkUpdate", payload.Empty().Add("data:", []moleculer.Payload{data}).Add("action", "multiply"))
				}
				ctx.Logger().Debug("mutiple calls -> account.bulkUpdate and user.bulkUpdate took: ", time.Since(startTime))

			},
		},
	}
}

var previouId = ""

func createSyntheticData(ctx moleculer.Context, profile moleculer.Payload) []moleculer.Payload {
	user := profile.Get("user")
	profileType := profile.Get("type")
	count := 1 + rand.Intn(1000)
	data := make([]moleculer.Payload, count)

	for i := 0; i < count; i++ {
		previousUser := payload.Empty()
		if previouId != "" {
			previousUser = <-ctx.Call("user.get", previouId)
		}
		previouId = previousUser.Get("id").String()

		data[i] = payload.Empty().
			Add("currentUser", user).
			Add("previousUser", previousUser).
			Add("profileType", profileType.String()).
			Add("batchSize", count).
			Add("dateTime", time.Now()).
			Add("rand", rand.Intn(1000))
	}
	return data
}

func main() {
	bkr := broker.New(&moleculer.Config{
		Transporter:                "TCP",
		WaitForDependenciesTimeout: 10 * time.Second,
		LogLevel:                   "INFO",
		DiscoverNodeID: func() string {
			return "data-service-node"
		}})

	bkr.Publish(&DataService{inputMessages: []moleculer.Payload{}})
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
