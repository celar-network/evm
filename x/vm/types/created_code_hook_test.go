package types

import (
	"errors"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
)

// The adder is an OPTIONAL extension, which is the whole reason it is not a
// method on OpCodeHooks: adding it there would break every downstream
// implementation. If someone later moves it onto that interface, this stops
// describing the design and the comment above it becomes false.
func TestDefaultHooksAcceptCreatedCodeHooks(t *testing.T) {
	var h OpCodeHooks = NewDefaultOpCodesHooks()
	if _, ok := h.(CreatedCodeHooksAdder); !ok {
		t.Fatal("DefaultOpCodesHooks does not accept created-code hooks, so an " +
			"app supplying one would be silently ignored")
	}
}

// Order and short-circuit, because a queue whose later hooks run after an
// earlier refusal would let a rejected creation be re-approved.
func TestCreatedCodeHooksRunInOrderAndStopOnError(t *testing.T) {
	var ran []string
	boom := errors.New("refused")

	h := NewDefaultOpCodesHooks().(*DefaultOpCodesHooks)
	h.AddCreatedCodeHooks(
		func(*vm.EVM, common.Address, []byte) error { ran = append(ran, "first"); return nil },
		func(*vm.EVM, common.Address, []byte) error { ran = append(ran, "second"); return boom },
		func(*vm.EVM, common.Address, []byte) error { ran = append(ran, "third"); return nil },
	)

	err := h.CreatedCodeHook(nil, common.Address{}, []byte{0x00})
	if !errors.Is(err, boom) {
		t.Fatalf("expected the refusal to propagate, got %v", err)
	}
	if len(ran) != 2 || ran[0] != "first" || ran[1] != "second" {
		t.Fatalf("hooks did not run in order and stop at the refusal: %v", ran)
	}
}
