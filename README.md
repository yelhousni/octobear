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

R1CS counts match the paper exactly (192 / 4022) or within 0.2% (5904 vs 5890).
SCS counts match the paper exactly for the classical and vector-linear rows;
the vector-Poseidon2 row is 0.5% above the paper (14 161 vs 14 094), likely a
small residual dispatch-related variance in the in-circuit sponge.

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

## License

Apache 2.0 — see `LICENSE`.
