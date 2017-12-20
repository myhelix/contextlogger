package buffered

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestBuffered(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "buffered Acceptance")
}
