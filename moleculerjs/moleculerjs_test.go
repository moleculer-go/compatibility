package moleculerjs

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/moleculer-go/moleculer/payload"

	"github.com/moleculer-go/moleculer"
	"github.com/moleculer-go/moleculer/broker"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func moleculerJs(natsUrl string) *exec.Cmd {

	cmdCtx, _ := context.WithTimeout(context.Background(), time.Minute*2)
	cmd := exec.CommandContext(cmdCtx, "npm", "install")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		fmt.Println("Failed on npm install - error: ", err)
	}

	cmdCtx, _ = context.WithTimeout(context.Background(), time.Second*20)
	cmd = exec.CommandContext(cmdCtx, "node", "services.js", natsUrl)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Start()
	if err != nil {
		fmt.Println("error starting node - error: ", err)
		return nil
	}
	fmt.Println("node started")
	return cmd
}

func natsTestHost() string {
	env := os.Getenv("NATS_HOST")
	if env == "" {
		return "localhost"
	}
	return env
}

var natsUrl = "nats://" + natsTestHost() + ":4222"

var _ = Describe("Moleculerjs", func() {

	It("should discover and call a moleculer JS service over NATS", func() {
		cmd := moleculerJs(natsUrl)
		Expect(cmd).ShouldNot(BeNil())
		jsEnded := make(chan bool)
		go func() {
			cmd.Wait()
			jsEnded <- true
		}()

		bkr := broker.New(&moleculer.Config{Transporter: natsUrl})
		userSvc := &UserService{profileCreated: make(chan bool)}
		bkr.Publish(userSvc)
		bkr.Start()
		time.Sleep(time.Second)

		r := <-bkr.Call("user.create", map[string]interface{}{
			"id":    10,
			"name":  "John",
			"email": "john@snow.com",
		})
		Expect(r.Error()).Should(BeNil())
		Expect(<-userSvc.profileCreated).Should(BeTrue())

		//test moleculer JS sending meta info on action to moleculer go
		onPanixCalled := false
		userSvc.OnPanix = func(ctx moleculer.Context) {
			Expect(ctx.Meta().Get("name").String()).Should(Equal("John"))
			Expect(ctx.Meta().Get("sword").String()).Should(Equal("Valyrian Steel"))
			onPanixCalled = true
		}

		mistake := <-bkr.Call("profile.mistake", true)
		Expect(mistake.IsError()).Should(BeTrue())
		fmt.Println("mistake: ", mistake)
		Expect(mistake.Error().Error()).Should(Equal("Error from JS side! panixError: [this action will panic!] failError: [this actions returns an error!]"))

		Expect(onPanixCalled).Should(BeTrue())

		// test moleculer Go sending meta info on action to moleculer JS //
		r = <-bkr.Call("profile.metarepeat", nil, moleculer.Options{
			Meta: payload.Empty().Add("country", "NZ").Add("cached", "maybe"),
		})
		fmt.Println("meta test: ", r)
		Expect(r.Get("meta").Exists()).Should(BeTrue())
		Expect(r.Get("meta").Get("country").String()).Should(Equal("NZ"))
		Expect(r.Get("meta").Get("cached").String()).Should(Equal("maybe"))

		Expect(r.Get("params").Exists()).Should(BeTrue())

		fmt.Println("checkAvailableServices - shuold bring all Moleculer js services")
		checkAvailableServices(bkr, []string{"account", "$node", "user", "profile"})

		r = <-bkr.Call("account.unregister", nil)
		Expect(r.Error()).Should(BeNil())

		time.Sleep(time.Millisecond * 300) // wait for local register to update

		fmt.Println("checkAvailableServices - after account service was unpublished from JS side")
		checkAvailableServices(bkr, []string{"$node", "user", "profile"})

		notifierSvc := &NotifierSvc{make(chan bool)}
		bkr.Publish(notifierSvc)
		time.Sleep(time.Millisecond * 300)

		finish := <-bkr.Call("profile.finish", true)

		Expect(finish.String()).Should(Equal("JS side will explode in 500 miliseconds!"))

		Expect(<-notifierSvc.received).Should(BeTrue())
		Expect(<-jsEnded).Should(BeTrue())

		// time.Sleep(time.Millisecond * 700) // wait for JS to exit and local register to update

		fmt.Println("checkAvailableServices - after JS broker ended")
		checkAvailableServices(bkr, []string{"$node", "user", "notifier"})

		// For the available services, we call
		// $node.services onlyAvailable:true and withEndpoints:true

		// But this returns services that have already been "unpublished".
		// Same thing happens on a node restart (when the nodeID stays the same):
		// We still see the service published by the previous instance

		//check that the moleculer go registry does not have the service aymore
		//

		bkr.Stop()
	})
})

func checkAvailableServices(bkr *broker.ServiceBroker, expectedServices []string) {
	services := <-bkr.Call("$node.services", map[string]interface{}{
		"onlyAvailable": true,
		"withEndpoints": true,
	})
	Expect(services.Error()).Should(BeNil())
	list := services.MapArray()
	fmt.Println("$node.services results: ")
	matches := 0
	for _, item := range list {
		name := item["name"].(string)
		for _, expected := range expectedServices {
			if expected == name {
				matches++
			}
		}
		fmt.Println("Name: ", name)
		fmt.Println("Available: ", item["available"])
		fmt.Println("HasLocal: ", item["hasLocal"])
		fmt.Println("Endpoints: ")
		for _, endpoint := range item["endpoints"].([]map[string]interface{}) {
			fmt.Println("  Available: ", endpoint["available"])
			fmt.Println("  NodeID: ", endpoint["nodeID"])
		}
		fmt.Println(" ")
	}
	fmt.Println("matches:", matches, " expected: ", len(list))
	Expect(matches).Should(Equal(len(list)))
}

type NotifierSvc struct {
	received chan bool
}

func (s *NotifierSvc) Name() string {
	return "notifier"
}

func (s *NotifierSvc) Send(ctx moleculer.Context, params moleculer.Payload) moleculer.Payload {
	ctx.Logger().Info("[notifier.send] params: ", params)

	n := payload.Empty().Add(
		"notificationId", "10").Add(
		"content", params)

	ctx.Emit("notifier.sent", n)

	go func() {
		s.received <- true
	}()
	return n
}
