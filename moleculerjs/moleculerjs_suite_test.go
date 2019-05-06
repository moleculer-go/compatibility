package moleculerjs

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestMoleculerjs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Moleculerjs Suite")
}
