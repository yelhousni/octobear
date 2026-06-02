// Package frontendtype detects whether the active gnark compiler is the SCS
// (Plonk) or R1CS frontend, so gadgets can pick the constraint-optimal
// algorithm for each backend.
//
// gnark exposes the same logic via its own internal/frontendtype package, but
// that package is `internal/` and therefore unimportable from outside the
// gnark module. A locally vendored copy with the same FrontendTyper interface
// does not work either: the gnark builder's FrontendType() method returns
// gnark's `Type`, not the local one, so the type assertion silently fails and
// every gadget falls back to its R1CS-optimal path even when compiling for
// SCS. That misdispatch was the source of the ≈10% SCS over-count in the
// paper-Table-2 reproduction.
//
// The fix here is reflection: we read the concrete compiler type's package
// path and decide from "/cs/scs" or "/cs/r1cs". The widecommitter wrapper
// preserves the embedded builder's PkgPath via its *embedded* field, which
// reflect.TypeOf does not see directly — so when wrapped, we fall back to
// invoking FrontendType() through reflect.Value.MethodByName.
package frontendtype

import (
	"reflect"
	"strings"
)

type Type int

const (
	R1CS Type = iota
	SCS
)

// DetectFromCompiler returns the active frontend type of a gnark compiler
// value, or false if it cannot be determined (e.g. test engine).
//
// It first inspects the concrete compiler's package path, which is the
// reliable path for unwrapped builders. If the compiler is wrapped (e.g. by
// widecommitter) it then invokes the FrontendType method by name via
// reflection; the wrapper forwards the method to the embedded builder, and
// reflection sidesteps the named-type mismatch on the return value.
func DetectFromCompiler(compiler any) (Type, bool) {
	if compiler == nil {
		return 0, false
	}
	t := reflect.TypeOf(compiler)
	for t != nil && t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t != nil {
		switch {
		case strings.HasSuffix(t.PkgPath(), "/cs/scs"):
			return SCS, true
		case strings.HasSuffix(t.PkgPath(), "/cs/r1cs"):
			return R1CS, true
		}
	}
	// Wrapped builder fallback: call FrontendType() reflectively. The result
	// is an int-kinded named type whose values match our local R1CS/SCS
	// constants by convention (both gnark's internal package and ours
	// declare `R1CS = iota` first then `SCS`).
	v := reflect.ValueOf(compiler)
	m := v.MethodByName("FrontendType")
	if !m.IsValid() || m.Type().NumIn() != 0 || m.Type().NumOut() != 1 {
		return 0, false
	}
	out := m.Call(nil)
	if len(out) != 1 || out[0].Kind() != reflect.Int {
		return 0, false
	}
	switch out[0].Int() {
	case int64(R1CS):
		return R1CS, true
	case int64(SCS):
		return SCS, true
	}
	return 0, false
}

// FrontendTyper is preserved for source-compatibility with the upstream PR
// code that asserts against it. The local gadgets should use
// DetectFromCompiler instead — see the package doc for why the interface
// assertion is unreliable across module boundaries.
type FrontendTyper interface {
	FrontendType() Type
}
