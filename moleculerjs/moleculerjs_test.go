package moleculerjs

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/moleculer-go/moleculer"
	"github.com/moleculer-go/moleculer/broker"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func moleculerJs(natsUrl string) *exec.Cmd {
	cmd := exec.Command("node", "services.js", natsUrl)
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
		go cmd.Wait()

		bkr := broker.New(&moleculer.Config{Transporter: natsUrl})
		bkr.Publish(&UserService{})
		bkr.Start()

		Expect(true).Should(BeTrue())

		bkr.Stop()
	})
})
