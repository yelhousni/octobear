// Package poseidon2_koalabear is a standalone in-circuit Poseidon2 permutation
// over the KoalaBear field. It is a direct port of the koalabear-specific
// branch added to gnark by PR #1757 (std/permutation/poseidon2/poseidon2.go +
// poseidon2_koalabear.go), reproduced here because the upstream package
// gates the M4 matrix selection on a private field of the Permutation struct
// that cannot be set from outside the package.
//
// The constructor consumes the native koalabear Poseidon2 parameters from
// gnark-crypto (field/koalabear/poseidon2.NewParameters), so circuit and
// native permutations produce identical outputs for the same input — verified
// by TestPoseidon2KoalaBearMatchesNative.
package poseidon2_koalabear

import (
	"errors"
	"fmt"
	"math/big"

	kbposeidon2 "github.com/consensys/gnark-crypto/field/koalabear/poseidon2"
	"github.com/consensys/gnark/frontend"
)

// ErrInvalidSizebuffer indicates the input slice length doesn't match the
// configured permutation width.
var ErrInvalidSizebuffer = errors.New("the size of the input should match the size of the hash buffer")

// diag16KoalaBearMinus1 holds the diagonal entries used by matMulInternalInPlace
// (state[i] = state[i] * DiagM1[i] + Σ state). Matches the unexported diag16
// array in gnark-crypto field/koalabear/poseidon2/hash.go; the "M1" suffix is
// a gnark convention shared with bn254 and is unrelated to a "-1" shift.
//
// p = 2^31 - 2^24 + 1 = 2_130_706_433.
var diag16KoalaBearMinus1 = [16]uint64{
	2130706431, // -2  mod p
	1,          //  1
	2,          //  2
	1065353217, //  1/2  mod p
	3,          //  3
	4,          //  4
	1065353216, // -1/2  mod p
	2130706430, // -3   mod p
	2130706429, // -4   mod p
	2122383361, //  1/2^8 mod p
	1864368129, //  1/8   mod p
	2130706306, //  1/2^24 mod p
	8323072,    // -1/2^8 mod p
	266338304,  // -1/8   mod p
	133169152,  // -1/16  mod p
	127,        // -1/2^24 mod p
}

// Parameters describes the koalabear Poseidon2 permutation configuration.
// Field layout mirrors gnark's std/permutation/poseidon2.Parameters so callers
// can be migrated mechanically.
type Parameters struct {
	Width           int
	DegreeSBox      int
	NbFullRounds    int
	NbPartialRounds int
	RoundKeys       [][]big.Int
	DiagM1          []big.Int
}

// Permutation is a koalabear-native in-circuit Poseidon2 permutation.
type Permutation struct {
	api    frontend.API
	params Parameters
}

// NewPoseidon2FromParameters builds a koalabear Poseidon2 Permutation with the
// given width and round counts. The round keys are derived from gnark-crypto's
// native koalabear poseidon2 parameters via the same deterministic seed.
func NewPoseidon2FromParameters(api frontend.API, width, nbFullRounds, nbPartialRounds int) (*Permutation, error) {
	if width != 16 {
		return nil, fmt.Errorf("koalabear poseidon2: in-circuit width %d not yet supported (only 16)", width)
	}
	native := kbposeidon2.NewParameters(width, nbFullRounds, nbPartialRounds)
	params := Parameters{
		Width:           native.Width,
		DegreeSBox:      kbposeidon2.DegreeSBox(),
		NbFullRounds:    native.NbFullRounds,
		NbPartialRounds: native.NbPartialRounds,
		RoundKeys:       make([][]big.Int, len(native.RoundKeys)),
	}
	for i := range params.RoundKeys {
		params.RoundKeys[i] = make([]big.Int, len(native.RoundKeys[i]))
		for j := range params.RoundKeys[i] {
			native.RoundKeys[i][j].BigInt(&params.RoundKeys[i][j])
		}
	}
	params.DiagM1 = make([]big.Int, 16)
	for i, v := range diag16KoalaBearMinus1 {
		params.DiagM1[i].SetUint64(v)
	}
	return &Permutation{api: api, params: params}, nil
}

// sBox applies the s-box to input[index].
func (h *Permutation) sBox(index int, input []frontend.Variable) {
	tmp := input[index]
	switch h.params.DegreeSBox {
	case 3:
		input[index] = h.api.Mul(input[index], input[index])
		input[index] = h.api.Mul(tmp, input[index])
	case 5:
		input[index] = h.api.Mul(input[index], input[index])
		input[index] = h.api.Mul(input[index], input[index])
		input[index] = h.api.Mul(input[index], tmp)
	case 7:
		input[index] = h.api.Mul(input[index], input[index])
		input[index] = h.api.Mul(input[index], tmp)
		input[index] = h.api.Mul(input[index], input[index])
		input[index] = h.api.Mul(input[index], tmp)
	case 17:
		input[index] = h.api.Mul(input[index], input[index])
		input[index] = h.api.Mul(input[index], input[index])
		input[index] = h.api.Mul(input[index], input[index])
		input[index] = h.api.Mul(input[index], input[index])
		input[index] = h.api.Mul(input[index], tmp)
	case -1:
		input[index] = h.api.Inverse(input[index])
	default:
		panic("sbox degree not supported")
	}
}

// matMulM4InPlace multiplies each 4-element chunk by the koalabear circulant
// M4 = circ(2,3,1,1). Addition chain mirrors gnark-crypto
// field/koalabear/poseidon2/poseidon2.go:176-191.
func (h *Permutation) matMulM4InPlace(s []frontend.Variable) {
	c := len(s) / 4
	for i := 0; i < c; i++ {
		t01 := h.api.Add(s[4*i], s[4*i+1])
		t23 := h.api.Add(s[4*i+2], s[4*i+3])
		t0123 := h.api.Add(t01, t23)
		t01123 := h.api.Add(t0123, s[4*i+1])
		t01233 := h.api.Add(t0123, s[4*i+3])
		out3 := h.api.Add(h.api.Mul(s[4*i], 2), t01233)
		out1 := h.api.Add(h.api.Mul(s[4*i+2], 2), t01123)
		out0 := h.api.Add(t01, t01123)
		out2 := h.api.Add(t23, t01233)
		s[4*i] = out0
		s[4*i+1] = out1
		s[4*i+2] = out2
		s[4*i+3] = out3
	}
}

// matMulExternalInPlace multiplies the state by the external MDS matrix
// circ(2·M4, M4, …, M4).
func (h *Permutation) matMulExternalInPlace(input []frontend.Variable) {
	switch h.params.Width {
	case 4:
		h.matMulM4InPlace(input)
	default:
		h.matMulM4InPlace(input)
		tmp := []frontend.Variable{0, 0, 0, 0}
		for i := 0; i < h.params.Width/4; i++ {
			tmp[0] = h.api.Add(tmp[0], input[4*i])
			tmp[1] = h.api.Add(tmp[1], input[4*i+1])
			tmp[2] = h.api.Add(tmp[2], input[4*i+2])
			tmp[3] = h.api.Add(tmp[3], input[4*i+3])
		}
		for i := 0; i < h.params.Width/4; i++ {
			input[4*i] = h.api.Add(input[4*i], tmp[0])
			input[4*i+1] = h.api.Add(input[4*i+1], tmp[1])
			input[4*i+2] = h.api.Add(input[4*i+2], tmp[2])
			input[4*i+3] = h.api.Add(input[4*i+3], tmp[3])
		}
	}
}

// matMulInternalInPlace multiplies the state by the internal matrix
// state[i] = state[i] * DiagM1[i] + Σ state.
func (h *Permutation) matMulInternalInPlace(input []frontend.Variable) {
	if len(h.params.DiagM1) != h.params.Width {
		panic("poseidon2_koalabear: missing DiagM1 for width >= 4")
	}
	sum := input[0]
	for i := 1; i < h.params.Width; i++ {
		sum = h.api.Add(sum, input[i])
	}
	for i := 0; i < h.params.Width; i++ {
		input[i] = h.api.Add(h.api.Mul(input[i], &h.params.DiagM1[i]), sum)
	}
}

// addRoundKeyInPlace adds the round-th key vector to the buffer.
func (h *Permutation) addRoundKeyInPlace(round int, input []frontend.Variable) {
	for i := 0; i < len(h.params.RoundKeys[round]); i++ {
		input[i] = h.api.Add(input[i], h.params.RoundKeys[round][i])
	}
}

// Permutation applies the Poseidon2 permutation in place.
func (h *Permutation) Permutation(input []frontend.Variable) error {
	if len(input) != h.params.Width {
		return ErrInvalidSizebuffer
	}

	h.matMulExternalInPlace(input)

	rf := h.params.NbFullRounds / 2
	for i := 0; i < rf; i++ {
		h.addRoundKeyInPlace(i, input)
		for j := 0; j < h.params.Width; j++ {
			h.sBox(j, input)
		}
		h.matMulExternalInPlace(input)
	}

	for i := rf; i < rf+h.params.NbPartialRounds; i++ {
		h.addRoundKeyInPlace(i, input)
		h.sBox(0, input)
		h.matMulInternalInPlace(input)
	}
	for i := rf + h.params.NbPartialRounds; i < h.params.NbFullRounds+h.params.NbPartialRounds; i++ {
		h.addRoundKeyInPlace(i, input)
		for j := 0; j < h.params.Width; j++ {
			h.sBox(j, input)
		}
		h.matMulExternalInPlace(input)
	}

	return nil
}
