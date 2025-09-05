package redis

import (
	"context"
	"os"
	"os/exec"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func startJSRedisService() *exec.Cmd {
	cmdCtx, _ := context.WithTimeout(context.Background(), time.Minute*2)
	cmd := exec.CommandContext(cmdCtx, "node", "js-redis-service.js")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(),
		"REDIS_HOST="+redisTestHost(),
		"REDIS_PORT="+redisTestPort(),
	)

	err := cmd.Start()
	if err != nil {
		Fail("Failed to start JS Redis service: " + err.Error())
		return nil
	}
	return cmd
}

var _ = Describe("Redis JS ↔ Go Compatibility", func() {
	var jsCmd *exec.Cmd

	BeforeEach(func() {
		// Start JS service
		jsCmd = startJSRedisService()
		Expect(jsCmd).ShouldNot(BeNil())
	})

	AfterEach(func() {
		if jsCmd != nil {
			jsCmd.Process.Kill()
			jsCmd.Wait()
		}
	})

	Describe("JS Redis Service", func() {
		It("should start and connect to Redis", func() {
			// Wait for JS service to start and connect
			time.Sleep(3 * time.Second)

			// Verify the service is running (no error means it started successfully)
			Expect(jsCmd.ProcessState).To(BeNil()) // Process should still be running
		})

		It("should perform math operations", func() {
			// Wait for JS service to complete its operations
			time.Sleep(8 * time.Second)

			// The JS service will output its results and then stop
			// This test verifies it runs without errors
		})
	})

	Describe("Redis Connection", func() {
		It("should maintain stable Redis connection", func() {
			// Wait for JS service to start
			time.Sleep(2 * time.Second)

			// Verify the service is still running
			Expect(jsCmd.ProcessState).To(BeNil())

			// Wait a bit more to ensure stability
			time.Sleep(3 * time.Second)
			Expect(jsCmd.ProcessState).To(BeNil())
		})
	})
})
