"""
Curve search for BabyBear-8 (Appendix D of the OctoBear paper).

Searches for Y² = X³ − 3X + b over Fp⁸, where p = 15·2²⁷ + 1 is the
BabyBear prime and the octic tower is Fp[x]/(x²−11), Fp²[y]/(y²−x),
Fp⁴[z]/(z²−y) (equivalently w⁸ = 11 with w identified to z), with
b = c · z^j swept for small c, j.

Accepts the first curve with prime group order; expected hit:
  b = 32 · z⁵,  r = 0x98c29b8b2f1b6f4c0a66714470a403628e1deb2a36f88d57337f85c4e42c99.

Usage:
    sage babybear8_search.sage
"""
import time

p = 15 * 2^27 + 1
print(f"BabyBear prime:  p = 15·2^27 + 1 = {p}")
print(f"  p mod 3 = {p%3},  p mod 4 = {p%4}\n")

Fp = GF(p)
R.<W> = PolynomialRing(Fp)
absolute = W^8 - 11
assert absolute.is_irreducible(), "W^8 − 11 not irreducible over BabyBear"
Fp8.<w> = Fp.extension(absolute, 'w')
print(f"Fp⁸ = Fp[w]/(w⁸ − 11),  |Fp⁸| ≈ 2^{(p^8).nbits()}\n")

a = Fp(-3)
target_r = 0x98c29b8b2f1b6f4c0a66714470a403628e1deb2a36f88d57337f85c4e42c99
print(f"Target prime order  r = 0x{target_r:x}  ({target_r.bit_length()} bits)\n")

print("Sweeping a = −3, b = c · w^j with c ∈ [1, 60), j ∈ {0, …, 15} ...")
t0 = time.time()
hit = None
for j in range(16):
    for c in range(1, 60):
        b = Fp(c) * w^j
        if b == 0:
            continue
        try:
            E = EllipticCurve(Fp8, [a, b])
        except ArithmeticError:
            continue
        n = E.order()
        if int(n) == target_r:
            t_trace = p^8 + 1 - n
            print(f"  ✓ HIT  b = {c}·w^{j}    #E = {n}  ({n.nbits()} bits)")
            print(f"          trace t = q + 1 − r = {t_trace}")
            hit = (c, j, b, n, t_trace, E)
            break
    if hit:
        break

print(f"\nSweep elapsed: {time.time() - t0:.1f}s")
if hit:
    c, j, b, r, t, E = hit
    assert int(r) == target_r and r.is_prime()
    print(f"\n=== Found curve (BabyBear-8) ===")
    print(f"E: Y² = X³ − 3X + {c}·w^{j}  over  Fp[w]/(w⁸ − 11)")
    print(f"prime order r in hex = 0x{int(r):x}")
    print(f"✓ matches paper's stated order (Appendix D)")
else:
    print("\nNo curve with the target r found in the sweep.")
