package nats

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestNats(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "NATS Moleculer JS ↔ Go Compatibility Suite")
}
