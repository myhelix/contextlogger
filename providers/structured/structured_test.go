// © 2017 Helix OpCo LLC. All rights reserved.
// Initial Author: jpecknerhelix

package structured

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestStructured(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "structured Acceptance")
}
