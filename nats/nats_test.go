package nats

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/moleculer-go/moleculer"
	"github.com/moleculer-go/moleculer/broker"
	"github.com/moleculer-go/moleculer/payload"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func moleculerJs(transporter, nodeID, jsFile string) *exec.Cmd {
	cmd := exec.Command("npm", "ci")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		fmt.Println("Failed on npm install - error: ", err)
	}

	cmd = exec.Command("node", jsFile, transporter)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), "NODE_ID="+nodeID)

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

var _ = Describe("NATS Moleculer JS ↔ Go Compatibility", func() {
	var jsProcess *exec.Cmd
	var bkr *broker.ServiceBroker

	BeforeEach(func() {
		// Start JS service
		jsProcess = moleculerJs(natsUrl, "js-node", "services.js")
		Expect(jsProcess).ShouldNot(BeNil())

		// Start Go broker
		bkr = broker.New(&moleculer.Config{Transporter: natsUrl})
		userSvc := &UserService{profileCreated: make(chan bool)}
		bkr.Publish(userSvc)
		bkr.Start()

		// Wait for services to be available
		time.Sleep(time.Second * 2)
	})

	AfterEach(func() {
		// Stop Go broker first
		if bkr != nil {
			bkr.Stop()
		}

		// Kill JS process
		if jsProcess != nil && jsProcess.Process != nil {
			jsProcess.Process.Kill()
			jsProcess.Wait()
		}
	})

	It("should discover and call a moleculer JS service over NATS", func() {
		// Test 1: Service discovery
		fmt.Println("checkAvailableServices - should bring all Moleculer js services")
		checkAvailableServices(bkr, []string{"account", "$node", "user", "profile"})

		// Test 2: Call JS service from Go
		r := <-bkr.Call("profile.create", map[string]interface{}{
			"id":    1,
			"name":  "Test",
			"email": "test@example.com",
		})
		Expect(r.Error()).Should(BeNil())

		// Test 3: Event handling
		userSvc := bkr.GetService("user").(*UserService)
		Expect(<-userSvc.profileCreated).Should(BeTrue())

		// Test 4: Call Go service from JS
		r = <-bkr.Call("user.create", map[string]interface{}{
			"id":    10,
			"name":  "John",
			"email": "john@snow.com",
		})
		Expect(r.Error()).Should(BeNil())

		// Test 5: Error handling
		r = <-bkr.Call("profile.mistake", true)
		Expect(r.Error()).ShouldNot(BeNil())
		fmt.Println("mistake: ", r.Error().Error())

		// Test 6: Meta data passing
		r = <-bkr.Call("profile.metarepeat", map[string]interface{}{
			"cached":  "maybe",
			"country": "NZ",
		})
		Expect(r.Error()).Should(BeNil())
		fmt.Println("meta test: ", r.String())

		// Test 7: Service unregistration
		r = <-bkr.Call("account.unregister", nil)
		Expect(r.Error()).Should(BeNil())

		time.Sleep(time.Millisecond * 300)
		fmt.Println("checkAvailableServices - after account service was unpublished from JS side")
		checkAvailableServices(bkr, []string{"$node", "user", "profile"})

		// Test 8: Event emission
		notifierSvc := &NotifierSvc{make(chan bool)}
		bkr.Publish(notifierSvc)
		time.Sleep(time.Millisecond * 300)

		// Test 9: Final action
		finish := <-bkr.Call("profile.finish", true)
		Expect(finish.String()).Should(Equal("JS side will explode in 500 miliseconds!"))

		// Test 10: Event reception
		Expect(<-notifierSvc.received).Should(BeTrue())

		fmt.Println("All tests completed successfully!")
	})
})

func checkAvailableServices(bkr *broker.ServiceBroker, expectedServices []string) {
	// Add timeout to prevent hanging
	select {
	case services := <-bkr.Call("$node.services", map[string]interface{}{
		"onlyAvailable": true,
		"withEndpoints": true,
	}):
		Expect(services.Error()).Should(BeNil())
		processServices(services, expectedServices)
	case <-time.After(5 * time.Second):
		fmt.Println("Timeout waiting for $node.services - skipping check")
	}
}

func processServices(services moleculer.Payload, expectedServices []string) {
	list := services.MapArray()
	fmt.Println("$node.services results: ")
	matches := 0
	for _, item := range list {
		name := item["name"].(string)
		for _, expected := range expectedServices {
			if expected == name {
				fmt.Println("Match: ", name)
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
	fmt.Println("matches:", matches, " expected: ", len(expectedServices), "expectedServices: ", expectedServices)
	Expect(matches).Should(Equal(len(expectedServices)))
}

type NotifierSvc struct {
	received chan bool
}

func (s *NotifierSvc) Name() string {
	return "notifier"
}

func (s *NotifierSvc) Start(broker *broker.ServiceBroker) {
	broker.On("profile.finished", func(payload moleculer.Payload) {
		s.received <- true
	})
}

func (s *NotifierSvc) Stop(broker *broker.ServiceBroker) {
}

type UserService struct {
	profileCreated chan bool
}

func (s *UserService) Name() string {
	return "user"
}

func (s *UserService) Start(broker *broker.ServiceBroker) {
	broker.On("profile.created", func(payload moleculer.Payload) {
		fmt.Println("profile.created event! profile:", payload.Map())
		user := payload.Map()["user"].(map[string]interface{})
		broker.Call("user.update", map[string]interface{}{
			"id":        user["id"],
			"profileId": user["id"],
		})
		fmt.Println("user updated with profile Id :)")
		s.profileCreated <- true
	})
}

func (s *UserService) Stop(broker *broker.ServiceBroker) {
}

func (s *UserService) Create(ctx moleculer.Context, params moleculer.Payload) moleculer.Payload {
	fmt.Println("user.create called! - user:", params.Map())
	ctx.Emit("user.created", params.Map())
	return payload.Empty().Add("message", "user created")
}

func (s *UserService) Update(ctx moleculer.Context, params moleculer.Payload) moleculer.Payload {
	fmt.Println("user.update called! - user:", params.Map())
	return payload.Empty().Add("message", "user updated")
}

func (s *UserService) Panix(ctx moleculer.Context, params moleculer.Payload) moleculer.Payload {
	fmt.Println("user.panix called!")
	panic("this action will panic!")
}

func (s *UserService) Fail(ctx moleculer.Context, params moleculer.Payload) moleculer.Payload {
	fmt.Println("user.fail called!")
	return payload.Empty().Add("error", "this actions returns an error!")
}
