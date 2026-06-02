package sw_octobear

import (
	"fmt"
	"strings"
	"testing"

	"github.com/consensys/gnark-crypto/field/koalabear"
	"github.com/consensys/gnark/constraint"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/r1cs"
	"github.com/consensys/gnark/frontend/cs/scs"

	"github.com/yelhousni/octobear/internal/widecommitter"
)

func compileNbConstraintsForTable(t *testing.T, circuit frontend.Circuit, builder frontend.NewBuilderU32) int {
	t.Helper()
	ccs, err := frontend.CompileGeneric[constraint.U32](
		koalabear.Modulus(),
		widecommitter.From(builder),
		circuit,
		frontend.WithCompressThreshold(10),
	)
	if err != nil {
		t.Fatalf("compile: %v", err)
	}
	return ccs.GetNbConstraints()
}

// TestPrintConstraintCounts reproduces Table 2 of the paper by compiling each
// of the three multiset-hash circuits (classical 1-point, vector linear,
// vector Poseidon2) on both the SCS and R1CS backends and reporting the
// constraint counts.
//
// Paper-reported counts (Table 2):
//
//	classical:        446 SCS / 192 R1CS
//	vector linear:  9 580 SCS / 4 022 R1CS
//	vector Poseidon2: 14 094 SCS / 5 890 R1CS
//
// R1CS counts should match the paper exactly. SCS counts may be a few percent
// above the paper if the local gnark release lacks the PR-side SCS frontend
// optimizations (the paper's measurements were taken on the PR feat/kb8 branch
// of gnark; this artifact targets the latest released gnark v0.15.0).
func TestPrintConstraintCounts(t *testing.T) {
	rows := []struct {
		name    string
		circuit frontend.Circuit
	}{
		{"Classical 1-point", &multisetHashSingleInsertCircuit{}},
		{"Vector linear (N=23, T=128)", &linearSingleInsertCircuit{}},
		{"Vector Poseidon2 (N=23)", &poseidon2SingleInsertCircuit{}},
	}

	var b strings.Builder
	fmt.Fprintln(&b)
	fmt.Fprintf(&b, "| %-30s | %8s | %8s |\n", "Circuit", "SCS", "R1CS")
	fmt.Fprintf(&b, "|%s|%s|%s|\n", strings.Repeat("-", 32), strings.Repeat("-", 10), strings.Repeat("-", 10))
	for _, r := range rows {
		scsN := compileNbConstraintsForTable(t, r.circuit, scs.NewBuilder)
		r1csN := compileNbConstraintsForTable(t, r.circuit, r1cs.NewBuilder)
		fmt.Fprintf(&b, "| %-30s | %8d | %8d |\n", r.name, scsN, r1csN)
	}
	t.Log(b.String())
}
