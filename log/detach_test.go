package log_test

import (
	"context"

	"github.com/myhelix/contextlogger/log"
	"github.com/myhelix/contextlogger/providers/dummy"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// guardedContext allows Value() during Detach's own synchronous extraction,
// then panics on any read after being armed, proving Detach never delegates
// Value() back to it once it's returned.
type guardedContext struct {
	context.Context
	armed *bool
}

func (g guardedContext) Value(key interface{}) interface{} {
	if *g.armed {
		panic("Value() should never be called on the original context after Detach")
	}
	return g.Context.Value(key)
}

var _ = Describe("Detach", func() {
	It("TestDetach_PreservesFields", func() {
		ctx := log.ContextWithFields(context.Background(), log.Fields{"foo": "bar"})

		detached := log.Detach(ctx)

		Expect(log.FieldsFromContext(detached)).To(HaveKeyWithValue("foo", "bar"))
	})

	It("TestDetach_PreservesProvider", func() {
		provider := dummy.LogProvider(nil)
		ctx := log.FromContextAndProvider(context.Background(), provider)

		detached := log.FromContext(log.Detach(ctx))

		Expect(detached.LogProvider()).To(BeIdenticalTo(provider))
	})

	It("TestDetach_ForwardsCancellation", func() {
		ctx, cancel := context.WithCancel(context.Background())
		detached := log.Detach(ctx)

		Expect(detached.Err()).NotTo(HaveOccurred())

		cancel()

		Eventually(detached.Done()).Should(BeClosed())
		Expect(detached.Err()).To(MatchError(context.Canceled))
	})

	It("TestDetach_ForwardsDeadline", func() {
		wantDeadline, cancel := context.WithTimeout(context.Background(), 0)
		defer cancel()
		wantTime, wantOK := wantDeadline.Deadline()

		detached := log.Detach(wantDeadline)
		gotTime, gotOK := detached.Deadline()

		Expect(gotOK).To(Equal(wantOK))
		Expect(gotTime).To(Equal(wantTime))
	})

	It("TestDetach_NeverReadsFromOriginalContextAfterConstruction", func() {
		base := log.ContextWithFields(context.Background(), log.Fields{"foo": "bar"})
		armed := false
		guarded := guardedContext{base, &armed}

		detached := log.Detach(guarded)
		armed = true

		Expect(func() {
			log.FieldsFromContext(detached)
			detached.Value("some-unrelated-key")
		}).NotTo(Panic())
	})
})
