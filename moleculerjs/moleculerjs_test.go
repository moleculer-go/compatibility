package moleculerjs

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/moleculer-go/moleculer"
	"github.com/moleculer-go/moleculer/broker"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func moleculerJs(natsUrl string) *exec.Cmd {

	cmd := exec.CommandContext(context.Background(), "node", "services.js", natsUrl)
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
		time.Sleep(time.Second * 2)

		r := <-bkr.Call("user.create", map[string]interface{}{
			"id":    10,
			"name":  "John",
			"email": "john@snow.com",
		})
		Expect(r.IsError()).Should(BeFalse())
		Expect(<-userSvc.profileCreated).Should(BeTrue())

		fmt.Println("<-userSvc.profileCreated")

		<-bkr.Call("profile.finish", true)
		Expect(<-jsEnded).Should(BeTrue())

		bkr.Stop()
	})
})
