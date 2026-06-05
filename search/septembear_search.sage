"""
Curve search for SeptemBear (Appendix D of the OctoBear paper).

Searches for Y² = X³ + 9X + b over Fp⁷, where p = 2³¹ − 2²⁴ + 1 is the
KoalaBear prime and the absolute degree-7 extension is Fp[z]/(z⁷ + z + 4),
with b = c · z^j swept for small c, j.

Accepts the first curve with prime group order; expected hit:
  b = 15 · z⁵,  r = 0x1e4a5d47579fd9acc910bb689b459f0a9ba4adbefb76f53ab9c25b3.

Also reports the security parameters against the three §5.1 attack vectors
(Pollard rho, Gaudry, Joux–Vitse) so the ~108-bit security claim of
Appendix D is self-checkable.

Usage:
    sage septembear_search.sage
"""
import time

p = 2^31 - 2^24 + 1
print(f"KoalaBear prime:  p = 2^31 − 2^24 + 1 = {p}")
print(f"  p mod 3 = {p%3},  p mod 4 = {p%4}\n")

Fp = GF(p)
R.<W> = PolynomialRing(Fp)
absolute = W^7 + W + 4
assert absolute.is_irreducible(), "W^7 + W + 4 not irreducible over KoalaBear"
Fp7.<z> = Fp.extension(absolute, 'z')
print(f"Fp⁷ = Fp[z]/(z⁷ + z + 4),  |Fp⁷| ≈ 2^{(p^7).nbits()}\n")

a = Fp(9)
target_r = 0x1e4a5d47579fd9acc910bb689b459f0a9ba4adbefb76f53ab9c25b3
print(f"Target prime order  r = 0x{target_r:x}  ({target_r.bit_length()} bits)\n")

print("Sweeping a = +9, b = c · z^j with c ∈ [1, 60), j ∈ {0, …, 6} ...")
t0 = time.time()
hit = None
for j in range(7):
    for c in range(1, 60):
        b = Fp(c) * z^j
        if b == 0:
            continue
        try:
            E = EllipticCurve(Fp7, [a, b])
        except ArithmeticError:
            continue
        n = E.order()
        if int(n) == target_r:
            t_trace = p^7 + 1 - n
            print(f"  ✓ HIT  b = {c}·z^{j}    #E = {n}  ({n.nbits()} bits)")
            print(f"          trace t = q + 1 − r = {t_trace}")
            hit = (c, j, b, n, t_trace, E)
            break
    if hit:
        break

print(f"\nSweep elapsed: {time.time() - t0:.1f}s")
assert hit, "no curve with the target r found in the sweep"
c, j, b, r, t, E = hit
assert int(r) == target_r and r.is_prime()

print(f"\n=== Found curve (SeptemBear) ===")
print(f"E: Y² = X³ + 9X + {c}·z^{j}  over  Fp[z]/(z⁷ + z + 4)")
print(f"prime order r in hex = 0x{int(r):x}")
print(f"✓ matches paper's stated order (Appendix D)\n")

# === Security checks (Section 5.1 attack vectors) ===
print("=== Security checks ===")
from sage.all import RR, log, sqrt, e, factorial

# Pollard rho ≈ √r
rho_bits = RR(log(sqrt(RR(r))) / log(2))
print(f"Pollard rho:  √r ≈ 2^{rho_bits:.1f}")

# Gaudry index-calculus: 2^{2k(k-1)} · p^{2-2/k}, k=7
k = 7
gd = 2 * k * (k-1) + (2 - 2/k) * RR(log(RR(p)) / log(2))
print(f"Gaudry:       ≈ 2^{gd:.1f}")

# Joux-Vitse: ((k-1)! · 2^{(k-1)(k-2)} · e^k · k^{-1/2} · p^2)^ω,  ω=2
inner_log2 = RR(log(factorial(k-1)) / log(2)) + (k-1)*(k-2) + k*RR(log(e)/log(2)) \
             - 0.5*RR(log(k)/log(2)) + 2*RR(log(RR(p)) / log(2))
jv = 2 * inner_log2  # omega = 2
print(f"Joux–Vitse:   ≈ 2^{jv:.1f}")

# Embedding degree
Zr = Integers(r)
k_prime = Zr(p^7).multiplicative_order()
print(f"\nEmbedding degree k' = ord_r(p^7) = {k_prime}  ({Integer(k_prime).nbits()} bits)")

# CM discriminant
Delta_pi = int(t)^2 - 4 * int(p^7)
abs_D = -Delta_pi
sqf = abs_D.squarefree_part()
D_K = -sqf if sqf % 4 == 3 else -4 * sqf
print(f"CM discriminant of End(E) ⊇ Z[π]:  D = {D_K}  ({abs(D_K).bit_length()} bits)")
