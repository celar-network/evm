package types

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
)

// OpCodeHooks extends the geth OpCodeHooks interface to add custom hooks for EVM operations.
// The hooks run before the respective opcode execution every time they are called.
type OpCodeHooks interface {
	vm.OpCodeHooks
	AddCallHooks(hooks ...CallHook)
	AddCreateHooks(hooks ...CreateHook)
}

// DefaultOpCodesHooks is the default implementation of OpCodeHooks for ATOM chain
// The hooks are used to enforce access control policies on EVM operations.
// They are ran BEFORE the respective opcode execution every time they are called.
type DefaultOpCodesHooks struct {
	callHooks        []CallHook
	createHooks      []CreateHook
	createdCodeHooks []CreatedCodeHook
}

// Make sure we comply with geth's OpCodeHooks interface
var _ OpCodeHooks = (*DefaultOpCodesHooks)(nil)

type (
	CreateHook func(ev *vm.EVM, caller common.Address) error
	CallHook   func(ev *vm.EVM, caller common.Address, recipient common.Address) error
)

// CreatedCodeHook runs on the runtime code a contract creation produced,
// after the constructor has run and before that code is written to state. A
// non-nil error fails the creation.
//
// WHY THIS EXISTS SEPARATELY FROM CreateHook. CreateHook answers "may this
// account deploy" and runs before the opcode, so it cannot answer "is this
// code permitted" - the code does not exist yet. A chain that admits
// contracts by what their code IS has no other point at which to decide.
type CreatedCodeHook func(ev *vm.EVM, address common.Address, code []byte) error

// CreatedCodeHooksAdder is an OPTIONAL extension to OpCodeHooks, detected by
// type assertion rather than added to that interface.
//
// Adding a method to OpCodeHooks would break every downstream implementation
// of it. Mirrors how geth exposes vm.CreatedCodeHooks, and for the same
// reason: an extension nobody is forced to implement.
type CreatedCodeHooksAdder interface {
	AddCreatedCodeHooks(hooks ...CreatedCodeHook)
}

// NewDefaultOpCodesHooks creates a new DefaultOpCodesHooks instance
func NewDefaultOpCodesHooks() OpCodeHooks {
	return &DefaultOpCodesHooks{}
}

// AddCallHooks adds one or more hooks to the queue to be executed before the CALL opcode.
// Hooks will be executed in the order they are added.
func (h *DefaultOpCodesHooks) AddCallHooks(hooks ...CallHook) {
	h.callHooks = append(h.callHooks, hooks...)
}

// AddCreateHooks adds one or more hooks to the queue to be executed before the CREATE opcode.
// Hooks will be executed in the order they are added.
func (h *DefaultOpCodesHooks) AddCreateHooks(hooks ...CreateHook) {
	h.createHooks = append(h.createHooks, hooks...)
}

// CreateHook checks if the caller has permission to deploy contracts
func (h *DefaultOpCodesHooks) CreateHook(evm *vm.EVM, caller common.Address) error {
	for _, hook := range h.createHooks {
		if err := hook(evm, caller); err != nil {
			return err
		}
	}
	return nil
}

// AddCreatedCodeHooks adds one or more hooks to be executed on the code a
// creation produced, before it is written to state. Executed in order.
func (h *DefaultOpCodesHooks) AddCreatedCodeHooks(hooks ...CreatedCodeHook) {
	h.createdCodeHooks = append(h.createdCodeHooks, hooks...)
}

// CreatedCodeHook satisfies geth's optional vm.CreatedCodeHooks interface.
func (h *DefaultOpCodesHooks) CreatedCodeHook(evm *vm.EVM, address common.Address, code []byte) error {
	for _, hook := range h.createdCodeHooks {
		if err := hook(evm, address, code); err != nil {
			return err
		}
	}
	return nil
}

// CallHook checks if the caller has permission to perform a call
func (h *DefaultOpCodesHooks) CallHook(evm *vm.EVM, caller common.Address, recipient common.Address) error {
	for _, hook := range h.callHooks {
		if err := hook(evm, caller, recipient); err != nil {
			return err
		}
	}
	return nil
}
