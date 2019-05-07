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

	cmdCtx, _ := context.WithTimeout(context.Background(), time.Second*20)
	cmd := exec.CommandContext(cmdCtx, "node", "services.js", natsUrl)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Start()
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
		userSvc := &UserService{make(chan bool)}
		bkr.Publish(userSvc)
		bkr.Start()
		time.Sleep(time.Second)

		r := <-bkr.Call("user.create", map[string]interface{}{
			"id":    10,
			"name":  "John",
			"email": "john@snow.com",
		})
		Expect(r.IsError()).Should(BeFalse())
		Expect(<-userSvc.profileCreated).Should(BeTrue())

		fmt.Println("<-userSvc.profileCreated")

		mistake := <-bkr.Call("profile.mistake", true)
		Expect(mistake.IsError()).Should(BeTrue())
		fmt.Println("mistake: ", mistake)
		Expect(mistake.Error().Error()).Should(Equal("Error from JS side! panixError: [this action will panic!] failError: [this actions returns an error!]"))

		notifierSvc := &NotifierSvc{make(chan bool)}
		bkr.Publish(notifierSvc)
		time.Sleep(time.Millisecond * 300)

		finish := <-bkr.Call("profile.finish", true)

		Expect(finish.String()).Should(Equal("JS side will explode in 500 miliseconds!"))

		Expect(<-notifierSvc.received).Should(BeTrue())
		Expect(<-jsEnded).Should(BeTrue())

		bkr.Stop()
	})
})

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
