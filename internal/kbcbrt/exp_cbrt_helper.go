// Package kbcbrt provides the koalabear-field cube-root and the two addition-
// chain exponentiations used by the extension-field cube-root algorithms.
// Released gnark-crypto v0.20.1's koalabear package exposes none of these, so
// we ship them here as free functions over koalabear.Element. The bodies are
// bit-identical to the upstream methods added in PR #832
// (field/koalabear/element_exp.go and field/koalabear/element.go).
package kbcbrt

import fr "github.com/consensys/gnark-crypto/field/koalabear"

// Cbrt sets z to ∛x (mod q) and returns z. Since q ≡ 2 (mod 3), cubing is a
// bijection on Fq, so every input has a unique cube root.
func Cbrt(z *fr.Element, x *fr.Element) *fr.Element {
	ExpByCbrt2q1o3(z, *x)
	return z
}

// ExpByCbrt2q1o3 is equivalent to z.Exp(x, 54aaaaab).
// It raises x to the (2q-1)/3 power using an addition chain.
func ExpByCbrt2q1o3(z *fr.Element, x fr.Element) *fr.Element {
	var t0, t1 fr.Element

	z.Square(&x)

	t1.Square(z)
	for s := 1; s < 2; s++ {
		t1.Square(&t1)
	}

	t0.Mul(&x, &t1)

	for range 2 {
		t1.Square(&t1)
	}

	t0.Mul(&t0, &t1)

	z.Mul(z, &t0)

	t1.Mul(&t0, z)

	t0.Mul(&x, &t1)

	t1.Mul(&t1, &t0)

	for range 8 {
		t1.Square(&t1)
	}

	t1.Mul(&t0, &t1)

	for range 8 {
		t1.Square(&t1)
	}

	t0.Mul(&t0, &t1)

	for range 7 {
		t0.Square(&t0)
	}

	z.Mul(z, &t0)

	return z
}

// ExpByCbrtHelperQMinus2Div9 is equivalent to z.Exp(x, e1c71c7).
// It raises x to the (q-2)/9 power using an addition chain.
//
// Used by the extension-field cube-root computations.
func ExpByCbrtHelperQMinus2Div9(z *fr.Element, x fr.Element) *fr.Element {
	// addition chain:
	//
	//	_10    = 2*1
	//	_11    = 1 + _10
	//	_110   = 2*_11
	//	_111   = 1 + _110
	//	i10    = _111 << 6
	//	i11    = _111 + i10
	//	return   ((i10 + i11) << 6 + _111) << 12 + i11
	//
	// Operations: 26 squares 6 multiplies
	var t0, t1 fr.Element

	z.Square(&x)
	z.Mul(&x, z)
	z.Square(z)
	t0.Mul(&x, z)

	t1.Square(&t0)
	for s := 1; s < 6; s++ {
		t1.Square(&t1)
	}

	z.Mul(&t0, &t1)
	t1.Mul(&t1, z)

	for range 6 {
		t1.Square(&t1)
	}

	t0.Mul(&t0, &t1)

	for range 12 {
		t0.Square(&t0)
	}

	z.Mul(z, &t0)

	return z
}
