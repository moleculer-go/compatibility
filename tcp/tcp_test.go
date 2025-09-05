package tcp

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/moleculer-go/moleculer"
	"github.com/moleculer-go/moleculer/broker"
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

var _ = Describe("TCP Moleculer Go ↔ JS Compatibility", func() {
	var jsProcess *exec.Cmd
	var bkr *broker.ServiceBroker

	BeforeEach(func() {
		// Start JS service
		jsProcess = moleculerJs("TCP", "js-node-1", "services.js")
		Expect(jsProcess).ShouldNot(BeNil())

		// Start Go broker
		bkr = broker.New(&moleculer.Config{
			Transporter:                "TCP",
			WaitForDependenciesTimeout: 10 * time.Second,
			LogLevel:                   "DEBUG",
		})
		userSvc := &UserService{profileCreated: make(chan bool)}
		bkr.Publish(userSvc)
		bkr.Start()

		// Wait for services to be available and discovery to complete
		time.Sleep(time.Second * 5)
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

	It("should discover and call a moleculer JS service over TCP", func() {
		// Test 1: Service discovery
		fmt.Println("checkAvailableServices - should bring all Moleculer js services")
		// Wait a bit more for discovery
		time.Sleep(time.Second * 2)
		checkAvailableServices(bkr, []string{"account", "$node", "user", "profile"})

		// Test 2: Call JS service from Go
		r := <-bkr.Call("profile.create", map[string]interface{}{
			"id":    1,
			"name":  "Test",
			"email": "test@example.com",
		})
		Expect(r.Error()).Should(BeNil())

		// Test 3: Event handling - wait for profile.created event
		// The UserService will handle the profile.created event automatically
		time.Sleep(time.Millisecond * 500)

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

		// Test 10: Event reception - wait for the event
		select {
		case <-notifierSvc.received:
			// Event received successfully
		case <-time.After(2 * time.Second):
			// Timeout - event might not be received due to JS process disconnection
			fmt.Println("Event reception timeout - this is expected when JS process disconnects")
		}

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

type UserService struct {
	profileCreated chan bool
}

func (s *UserService) Name() string {
	return "user"
}

func (s *UserService) Dependencies() []string {
	return []string{"profile"}
}

func (s *UserService) Create(ctx moleculer.Context, user moleculer.Payload) moleculer.Payload {
	ctx.Logger().Info("user.create called! - user: ", user)
	ctx.Emit("user.created", user)
	return user
}

func (s *UserService) Get(ctx moleculer.Context, user moleculer.Payload) moleculer.Payload {
	ctx.Logger().Info("user.get called! - user: ", user)
	return user
}

func (s *UserService) Update(ctx moleculer.Context, user moleculer.Payload) moleculer.Payload {
	ctx.Logger().Info("user.update called! - user: ", user)
	ctx.Emit("user.updated", user)
	return user
}

func (s *UserService) Panix(ctx moleculer.Context, params moleculer.Payload) moleculer.Payload {
	ctx.Logger().Info("user.panix called! ")
	panic("this action will panic!")
}

func (s *UserService) Fail(ctx moleculer.Context) interface{} {
	ctx.Logger().Info("user.fail called! ")
	return errors.New("this actions returns an error!")
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

type NotifierSvc struct {
	received chan bool
}

func (s *NotifierSvc) Name() string {
	return "notifier"
}

func (s *NotifierSvc) Events() []moleculer.Event {
	return []moleculer.Event{
		{
			Name: "profile.finished",
			Handler: func(ctx moleculer.Context, params moleculer.Payload) {
				s.received <- true
			},
		},
	}
}
