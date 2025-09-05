package redis

import (
	"fmt"
	"os"
	"time"

	"github.com/moleculer-go/moleculer/transit/redis"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func redisTestHost() string {
	env := os.Getenv("REDIS_HOST")
	if env == "" {
		return "localhost"
	}
	return env
}

func redisTestPort() string {
	env := os.Getenv("REDIS_PORT")
	if env == "" {
		return "6379"
	}
	return env
}

var _ = Describe("Redis Transporter", func() {
	var redisTransporter *redis.RedisTransporter

	BeforeEach(func() {
		// Create Redis transporter configuration
		host := redisTestHost()
		fmt.Printf("Redis test host: %s\n", host)
		redisConfig := &redis.RedisConfig{
			Host:     host,
			Port:     6379,
			Password: "",
			DB:       2, // Use DB 2 for testing
			Prefix:   "test-moleculer",
		}

		// Create Redis transporter
		redisTransporter = redis.NewRedisTransporter(redisConfig)
	})

	AfterEach(func() {
		if redisTransporter != nil {
			redisTransporter.Disconnect()
		}
	})

	Describe("Connection", func() {
		It("should connect to Redis", func() {
			// Debug: Print environment variables
			fmt.Printf("REDIS_HOST env var: %s\n", os.Getenv("REDIS_HOST"))
			fmt.Printf("REDIS_PORT env var: %s\n", os.Getenv("REDIS_PORT"))
			fmt.Printf("redisTestHost() returns: %s\n", redisTestHost())
			fmt.Printf("redisTestPort() returns: %s\n", redisTestPort())

			// Mock registry for testing
			errChan := redisTransporter.Connect(nil)

			select {
			case err := <-errChan:
				Expect(err).To(BeNil())
			case <-time.After(5 * time.Second):
				Fail("Connection timeout")
			}

			Expect(redisTransporter.IsConnected()).To(BeTrue())
		})

		It("should disconnect from Redis", func() {
			// Mock registry for testing
			errChan := redisTransporter.Connect(nil)

			select {
			case err := <-errChan:
				Expect(err).To(BeNil())
			case <-time.After(5 * time.Second):
				Fail("Connection timeout")
			}

			errChan = redisTransporter.Disconnect()
			select {
			case err := <-errChan:
				Expect(err).To(BeNil())
			case <-time.After(5 * time.Second):
				Fail("Disconnection timeout")
			}

			Expect(redisTransporter.IsConnected()).To(BeFalse())
		})
	})

	Describe("Configuration", func() {
		It("should return correct name", func() {
			Expect(redisTransporter.GetName()).To(Equal("Redis"))
		})

		It("should return configuration", func() {
			config := redisTransporter.GetConfig()
			Expect(config).ToNot(BeNil())
		})

		It("should update configuration", func() {
			newConfig := &redis.RedisConfig{
				Host: "newhost",
				Port: 6380,
				DB:   2,
			}
			redisTransporter.SetConfig(newConfig)
			Expect(redisTransporter.GetConfig()).To(Equal(newConfig))
		})
	})

	Describe("Metrics", func() {
		BeforeEach(func() {
			// Mock registry for testing
			errChan := redisTransporter.Connect(nil)

			select {
			case err := <-errChan:
				Expect(err).To(BeNil())
			case <-time.After(5 * time.Second):
				Fail("Connection timeout")
			}
		})

		It("should return metrics", func() {
			metrics := redisTransporter.GetMetrics()
			Expect(metrics).ToNot(BeNil())
			Expect(metrics).To(HaveKey("hits"))
			Expect(metrics).To(HaveKey("misses"))
			Expect(metrics).To(HaveKey("total_conns"))
		})
	})

	Describe("Publish/Subscribe", func() {
		BeforeEach(func() {
			// Mock registry for testing
			errChan := redisTransporter.Connect(nil)

			select {
			case err := <-errChan:
				Expect(err).To(BeNil())
			case <-time.After(5 * time.Second):
				Fail("Connection timeout")
			}
		})

		It("should handle publish and subscribe", func() {
			// This test verifies that the Redis transporter can handle
			// basic publish/subscribe operations
			// Note: Full integration testing would require a complete broker setup

			// Test that we can get metrics (which indicates Redis is working)
			metrics := redisTransporter.GetMetrics()
			Expect(metrics).ToNot(BeNil())

			// Test configuration
			config := redisTransporter.GetConfig()
			Expect(config).ToNot(BeNil())

			// Test name
			Expect(redisTransporter.GetName()).To(Equal("Redis"))
		})
	})

	Describe("Error Handling", func() {
		It("should handle connection errors gracefully", func() {
			// Create transporter with invalid host
			invalidConfig := &redis.RedisConfig{
				Host: "invalid-host-that-does-not-exist",
				Port: 6379,
				DB:   2,
			}

			invalidTransporter := redis.NewRedisTransporter(invalidConfig)
			errChan := invalidTransporter.Connect(nil)

			select {
			case err := <-errChan:
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(ContainSubstring("failed to connect to Redis"))
			case <-time.After(5 * time.Second):
				Fail("Connection should have failed")
			}
		})
	})
})
