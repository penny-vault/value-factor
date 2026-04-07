package value_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestValueFactor(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Value Factor Suite")
}
