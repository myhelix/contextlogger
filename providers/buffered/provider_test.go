package buffered

import (
	"bytes"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/myhelix/contextlogger/log"
	"github.com/myhelix/contextlogger/providers"
)

var _ = Describe("bufferedLogProvider", func() {

	var (
		buffer        *bytes.Buffer
		provider      providers.LogProvider
		contextLogger log.ContextLogger
	)

	BeforeEach(func() {
		buffer = new(bytes.Buffer)
		provider = LogProvider(buffer)
		contextLogger = log.BackgroundContext()
	})

	It("Should log level", func() {
		provider.Error(contextLogger, true, "Error message")
		Ω(strings.Contains(buffer.String(), "level=error")).Should(BeTrue())
	})

	It("Should log message", func() {
		provider.Error(contextLogger, true, "Error message")
		Ω(strings.Contains(buffer.String(), "msg=\"[Error message]\"")).Should(BeTrue())
	})

	It("Should log report status", func() {
		provider.Error(contextLogger, true, "Error message")
		Ω(strings.Contains(buffer.String(), "report=true")).Should(BeTrue())
	})

})
