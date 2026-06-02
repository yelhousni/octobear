package poseidon2_koalabear

import (
	"testing"

	"github.com/consensys/gnark-crypto/field/koalabear"
	kbposeidon2 "github.com/consensys/gnark-crypto/field/koalabear/poseidon2"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/test"
)

const koalaBearWidth = 16

type poseidon2KoalaBearCircuit struct {
	Input  [koalaBearWidth]frontend.Variable
	Output [koalaBearWidth]frontend.Variable `gnark:",public"`
}

func (c *poseidon2KoalaBearCircuit) Define(api frontend.API) error {
	h, err := NewPoseidon2FromParameters(api, koalaBearWidth, 6, 21)
	if err != nil {
		return err
	}
	state := make([]frontend.Variable, koalaBearWidth)
	copy(state, c.Input[:])
	if err := h.Permutation(state); err != nil {
		return err
	}
	for i := range c.Output {
		api.AssertIsEqual(c.Output[i], state[i])
	}
	return nil
}

func TestPoseidon2KoalaBearMatchesNative(t *testing.T) {
	assert := test.NewAssert(t)

	var in [koalaBearWidth]koalabear.Element
	for i := range in {
		in[i].SetUint64(uint64(i)*1234567 + 7)
	}

	native := kbposeidon2.NewPermutation(koalaBearWidth, 6, 21)
	state := in
	if err := native.Permutation(state[:]); err != nil {
		t.Fatal(err)
	}

	var witness poseidon2KoalaBearCircuit
	for i := range in {
		witness.Input[i] = in[i].String()
		witness.Output[i] = state[i].String()
	}
	assert.CheckCircuit(&poseidon2KoalaBearCircuit{},
		test.WithValidAssignment(&witness),
		test.WithoutCurveChecks(),
		test.WithSmallfieldCheck())
}
