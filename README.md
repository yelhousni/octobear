# Octobear

[![License](https://img.shields.io/badge/license-MIT-blue)](LICENSE) [![Go Report Card](https://goreportcard.com/badge/github.com/yelhousni/octobear)](https://goreportcard.com/report/github.com/yelhousni/octobear) [![Go Reference](https://pkg.go.dev/badge/github.com/yelhousni/octobear.svg)](https://pkg.go.dev/github.com/yelhousni/octobear)

<p align="center">
  <picture>
    <img src="octobear.png" width="200">
  </picture>
</p>



Reproducibility artifact for the paper **"Constraint-Friendly
Post-Quantum-Extendable Elliptic-Curve Multiset Hashing for zkVM Memory
Arguments."** It implements:

- the Octobear curve `E: Y^2 = X^3 - 3X + 17·w^5` over the octic extension
  `F_{p^8}` of the KoalaBear field (`p = 2^31 - 2^24 + 1`),
- three multiset-hash variants on Octobear (classical 1-point, vector linear,
  vector Poseidon2-sponge), both natively and in-circuit,
- the in-circuit KoalaBear Poseidon2 permutation (width 16, M4 = circ(2,3,1,1)).

The repo is a single self-contained Go module that depends on released
`gnark-crypto` v0.20.1 and `gnark` v0.15.0.

## Layout

```
curve/                          Native Octobear curve (octic ext. of KoalaBear)
  extensions/                   F_p^8 tower (E2/E4/E8 + torus-based Cbrt)
  fp/  fr/                      Base + scalar field
  internal/fptower/             Internal tower re-exports
  multiset-hash/                Classical + vector multiset-hash variants
circuits/                       In-circuit gadgets (gnark)
  poseidon2_koalabear/          Standalone in-circuit Poseidon2 over KoalaBear
  fields_octobear/              In-circuit F_p^8 tower
  maptocurve_octobear/          y-increment map + Linear/Poseidon2 vector maps
  sw_octobear/                  In-circuit curve + multiset-hash accumulators
internal/kbcbrt/                Free-function koalabear Cbrt + addchain helpers
internal/parallel/              Local copy of gnark-crypto's parallel package
internal/frontendtype/          Local copy of gnark's frontendtype helper
internal/widecommitter/         Local copy of gnark's widecommitter test harness
internal/field/asm/element_4w/  KoalaBear asm includes (referenced by curve/fr)
```

## Reproducing the paper's measurements

```sh
go test ./...                                # all unit tests
go test ./circuits/sw_octobear/ -run TestPrintConstraintCounts -v
```

The print test compiles each multiset-hash circuit (one Insert) on both the
SCS and R1CS frontends and prints the constraint counts. Expected output:

| Circuit                        |      SCS |     R1CS |
|--------------------------------|----------|----------|
| Classical 1-point              |      446 |      192 |
| Vector linear (N=23, T=128)    |     9580 |     4022 |
| Vector Poseidon2 (N=23)        |    14161 |     5904 |

## Other benchmarks

Native (out-of-circuit) map and hash benchmarks:

```sh
# one-point variant: Map(uint16) + Insert + Hash over 256 messages
go test ./curve/multiset-hash -bench 'BenchmarkMap$|BenchmarkAccumulatorInsert|BenchmarkHash256' -run '^$' -benchtime=3s

# vector linear (N=23, T=128): per-message map + 256-message hash
go test ./curve/multiset-hash -bench 'BenchmarkMapLinear|BenchmarkHashLinear256' -run '^$' -benchtime=3s

# vector Poseidon2 (N=23, T=256): per-message map + 256-message hash
go test ./curve/multiset-hash -bench 'BenchmarkMapPoseidon2|BenchmarkHashPoseidon2_256' -run '^$' -benchtime=3s
```

In-circuit witness-solve benchmarks (one Insert on each variant, SCS and R1CS):

```sh
go test ./circuits/sw_octobear -bench 'BenchmarkMultisetHashCircuitSolve|BenchmarkLinearMultisetHashCircuitSolve|BenchmarkPoseidon2MultisetHashCircuitSolve' -run '^$' -benchtime=1x -v
```

## License

Apache 2.0 — see `LICENSE`.
