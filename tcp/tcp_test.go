package tcp

import (
	"context"
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
	cmdCtx, _ := context.WithTimeout(context.Background(), time.Minute*2)
	cmd := exec.CommandContext(cmdCtx, "npm", "install")
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

	BeforeEach(func() {
		// Start JS service
		jsProcess = moleculerJs("TCP", "js-node-1", "services.js")
		Expect(jsProcess).ShouldNot(BeNil())
	})

	AfterEach(func() {
		// Kill JS process (same as Redis test)
		if jsProcess != nil && jsProcess.Process != nil {
			jsProcess.Process.Kill()
			jsProcess.Wait()
		}
	})

	It("should discover and call a moleculer JS service over TCP", func() {

		bkr := broker.New(&moleculer.Config{
			Transporter:                "TCP",
			WaitForDependenciesTimeout: 10 * time.Second,
			LogLevel:                   "DEBUG",
			DiscoverNodeID: func() string {
				return "moleculer-go"
			},
		})

		userSvc := &UserService{profileCreated: make(chan bool)}
		bkr.Publish(userSvc)
		bkr.Start()
		fmt.Println("waiting for profile service")
		bkr.WaitFor("profile")
		fmt.Println("profile service is available")

		r := <-bkr.Call("user.create", map[string]interface{}{
			"id":    10,
			"name":  "John",
			"email": "john@snow.com",
		})
		Expect(r.Error()).Should(BeNil())
		Expect(<-userSvc.profileCreated).Should(BeTrue())

		//get the internal state of the moleculer broker
		r = <-bkr.Call("profile.listServices", nil)
		Expect(r.Error()).Should(BeNil())
		fmt.Println("listServices: ", r.MapArray())

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

		fmt.Println("checkAvailableServices - should bring all Moleculer js services")
		checkAvailableServices(bkr, []string{"account", "$node", "user", "profile"})

		r = <-bkr.Call("account.unregister", nil)
		Expect(r.Error()).Should(BeNil())

		time.Sleep(time.Millisecond * 400) // wait for local register to update

		fmt.Println("checkAvailableServices - after account service was unpublished from JS side")
		checkAvailableServices(bkr, []string{"$node", "user", "profile"})

		notifierSvc := &NotifierSvc{make(chan bool)}
		bkr.Publish(notifierSvc)

		time.Sleep(time.Second * 2)

		finish := <-bkr.Call("profile.finish", true)
		Expect(finish.String()).Should(Equal("JS side will explode in 500 miliseconds!"))

		Expect(<-notifierSvc.received).Should(BeTrue())

		fmt.Println("checkAvailableServices - after JS broker ended")
		checkAvailableServices(bkr, []string{"$node", "user", "notifier"})

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
	OnPanix        func(moleculer.Context)
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
	if s.OnPanix != nil {
		s.OnPanix(ctx)
	}
	panic("this action will panic!")
}

func (s *UserService) Fail(ctx moleculer.Context) interface{} {
	ctx.Logger().Info("user.fail called! ")
	return fmt.Errorf("this actions returns an error!")
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

type ProfileService struct{}

func (s *ProfileService) Name() string {
	return "profile"
}

func (s *ProfileService) Create(ctx moleculer.Context, user moleculer.Payload) moleculer.Payload {
	ctx.Logger().Info("[moleculer-Go] profile.create action user: ", user)
	id := user.Get("id").Int()
	name := user.Get("name").String()
	email := user.Get("email").String()
	profile := payload.Empty().
		Add("user", payload.Empty().Add("id", id).Add("name", name).Add("email", email)).
		Add("type", "web-user")
	ctx.Emit("profile.created", profile)
	return profile
}

func (s *ProfileService) ListServices(ctx moleculer.Context, params moleculer.Payload) moleculer.Payload {
	return <-ctx.Call("$node.services", nil)
}

func (s *ProfileService) Metarepeat(ctx moleculer.Context, params moleculer.Payload) moleculer.Payload {
	ctx.Logger().Info("[moleculer-Go] profile.metarepeat ctx.meta: ", ctx.Meta())
	return payload.Empty().
		Add("meta", ctx.Meta()).
		Add("params", params)
}

func (s *ProfileService) Mistake(ctx moleculer.Context, params moleculer.Payload) moleculer.Payload {
	ctx.Logger().Info("[moleculer-Go] profile.mistake called with: ", params)
	panixError := <-ctx.Call("user.panix", payload.Empty().
		Add("name", ctx.Meta().Get("name").String()).
		Add("sword", ctx.Meta().Get("sword").String()))
	failError := <-ctx.Call("user.fail", nil)
	return payload.Empty().Add("message", fmt.Sprintf("Error from Go side! panixError: [%v] failError: [%v]", panixError, failError))
}

func (s *ProfileService) Finish(ctx moleculer.Context, params moleculer.Payload) moleculer.Payload {
	ctx.Logger().Info("[moleculer-Go] profile.finish called with: ", params)
	ctx.Emit("profile.finished", payload.Empty().Add("message", "Go side will explode in 500 miliseconds!"))
	go func() {
		time.Sleep(500 * time.Millisecond)
		// In a real scenario, this would stop the broker
	}()
	return payload.Empty().Add("message", "Go side will explode in 500 miliseconds!")
}

func (s *ProfileService) Unregister(ctx moleculer.Context, params moleculer.Payload) moleculer.Payload {
	ctx.Logger().Info("[moleculer-Go] profile.unregister called")
	// In a real scenario, this would unregister the account service
	return payload.Empty().Add("message", "account service unregistered")
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
