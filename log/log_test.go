package log

import (
	"testing"

	"github.com/calm/contextlogger/v2/sanitizer"
	"github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestSuiteLog(t *testing.T) {
	RegisterFailHandler(ginkgo.Fail)
	RunSpecs(t, "Log Tests")
}

var _ = Describe("log package tests", func() {

	Context("Field sanitization", func() {

		It("the WithField method should sanitize its data", func() {
			ctx := BackgroundContext().WithField("password", "secret")
			fields, ok := ctx.Value(contextLogFieldsKey{}).(Fields)
			Expect(ok).To(BeTrue())
			Expect(fields).To(Equal(Fields{"password": sanitizer.DefaultSanitizePlaceholder}))
		})

		It("the WithFields method should sanitize its data", func() {

			type S struct {
				Token string
				Addr  string
			}

			ctx := BackgroundContext().WithFields(Fields{
				"password": "secret",
				"greeting": "hello",
				"s":        S{Token: "foo", Addr: "foo@bar.com"},
			})
			fields, ok := ctx.Value(contextLogFieldsKey{}).(Fields)
			Expect(ok).To(BeTrue())
			Expect(fields).To(Equal(Fields{
				"password": sanitizer.DefaultSanitizePlaceholder,
				"greeting": "hello",
				"s":        S{Token: sanitizer.DefaultSanitizePlaceholder, Addr: sanitizer.DefaultSanitizePlaceholder},
			}))
		})

	})

})
