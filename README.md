# Octobear

Reproducibility artifact for the paper **"OctoBear: Constraint-Friendly
Post-Quantum-Extendable Elliptic-Curve Multiset Hashing for zkVM Memory
Arguments."** It implements:

- the Octobear curve `E: Y^2 = X^3 - 3X + 17·w^5` over the octic extension
  `F_{p^8}` of the KoalaBear field (`p = 2^31 - 2^24 + 1`),
- three multiset-hash variants on Octobear (classical 1-point, vector linear,
  vector Poseidon2-sponge), both natively and in-circuit,
- the in-circuit KoalaBear Poseidon2 permutation (width 16, M4 = circ(2,3,1,1)).

The repo is a single self-contained Go module that depends on released
`gnark-crypto` v0.20.1 and `gnark` v0.15.0 — no `replace` directives.

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
| Classical 1-point              |      492 |      192 |
| Vector linear (N=23, T=128)    |    10638 |     4022 |
| Vector Poseidon2 (N=23)        |    15219 |     5904 |

R1CS counts match the paper's Table 2 (192 / 4022 / 5890) exactly except for
the Poseidon2 R1CS row, which is within 0.2 % (5904 vs 5890). SCS counts are
≈ 5–11 % higher than the paper because the released gnark v0.15.0 SCS
frontend lacks the SCS optimizations on the PR's feat/kb8 branch that
produced the paper's measurements.

## Native benchmarks

```sh
go test ./curve/multiset-hash -bench . -run '^$' -benchtime=3s
```

## Circuit benchmarks

```sh
go test ./circuits/sw_octobear -bench 'BenchmarkMultisetHashCircuitSolve|BenchmarkLinearMultisetHashCircuitSolve|BenchmarkPoseidon2MultisetHashCircuitSolve' -run '^$' -benchtime=1x -v
```

Expected order-of-magnitude (Apple M5, darwin/arm64):

- `BenchmarkMap-10`: ~25 µs/op (cubic solve + accumulator)
- `BenchmarkHash256-10`: ~7 ms/op (256 inserts, one-point variant)
- `BenchmarkHashLinear256-10`: ~160 ms/op (256 inserts, 23-coord linear)
- `BenchmarkHashPoseidon2_256-10`: ~160 ms/op (256 inserts, 23-coord Poseidon2)

## Provenance

Code is ported from two upstream PRs on the `feat/kb8` branches of the
respective repositories:

- gnark-crypto [#832](https://github.com/Consensys/gnark-crypto/pull/832) (head `f6b0b47`)
- gnark [#1757](https://github.com/Consensys/gnark/pull/1757) (head `0a7b510`)

The few upstream-file edits in those PRs that cannot be applied from an
external module — the koalabear `Cbrt`/`ExpByCbrt*` helpers, the
`useKoalaBearM4` flag on `gnark/std/permutation/poseidon2.Permutation`, and
the `ecc.OCTOBEAR` enum entry — are reimplemented locally
(`internal/kbcbrt/`, `circuit/poseidon2_koalabear/`,
`curve/octobear/octobear.go`).

## License

Apache 2.0 — see `LICENSE`.
