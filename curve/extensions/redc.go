// Copyright 2020-2026 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package extensions

// montReduceLazy is the lazy Montgomery reduction used by the inlined
// extension-field multiplications below (E8.Mul, E8.Square). Result is in
// [0, 2q); a final reduceSmall canonicalizes.
//
// Vendored from gnark-crypto's field/koalabear/extensions/e6.go because
// e6.go is not part of this artifact.
func montReduceLazy(v uint64) uint64 {
	m := uint32(v) * qInvNeg
	return (v + uint64(m)*uint64(q)) >> 32
}
