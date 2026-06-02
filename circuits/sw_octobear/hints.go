package sw_octobear

import (
	"github.com/consensys/gnark/constraint/solver"
	"github.com/yelhousni/octobear/circuits/fields_octobear"
)

func GetHints() []solver.Hint {
	return fields_octobear.GetHints()
}
