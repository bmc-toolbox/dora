package gingonic_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestGingonic(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Gingonic Suite")
}
