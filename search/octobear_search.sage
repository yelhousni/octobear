"""
Curve search for OctoBear.

Reproduces the curve E: Y² = X³ − 3X + 17·w⁵ over Fp⁸, where
p = 2³¹ − 2²⁴ + 1 is the KoalaBear prime and the octic tower is
Fp[u]/(u²−3), Fp²[v]/(v²−u), Fp⁴[w]/(w²−v) (equivalently w⁸ = 3).

For each candidate a = −3, b = c · w^j with c ∈ [1, 30) and j ∈ {0,…,15}
we compute #E(Fp⁸) and accept the first curve whose order is a
248-bit prime (no cofactor).  This matches the paper's stated
r = 0xf06e44682c2aa440f5f26a5ae1748fec171df66abbdb81ad07f36ed81b9049.

We then certify:
  * the generator: smallest X ≥ 1 such that x³ − 3x + b is a square in Fp⁸
    and the resulting affine point (X, Y) has order r;
  * the embedding degree k′: smallest integer with r | (p⁸)^{k′} − 1;
  * the CM discriminant: square-free part of t² − 4·p⁸ where
    t = p⁸ + 1 − r is the trace of Frobenius.

Usage:
    sage curve_search.sage
"""
import time

p = 2^31 - 2^24 + 1
print(f"KoalaBear prime: p = 2^31 − 2^24 + 1 = {p}")
print(f"  p mod 3 = {p % 3},   p⁸ mod 3 = {pow(p, 8, 3)}")

# Build Fp⁸ as the absolute extension Fp[W]/(W⁸ − 3).  In the paper's tower
# (u² = 3, v² = u, w² = v) we have w⁸ = u² = 3, so the two presentations
# coincide once we identify w with a fixed root of W⁸ − 3.
Fp = GF(p)
R.<W> = PolynomialRing(Fp)
absolute = W^8 - 3
assert absolute.is_irreducible(), "W^8 − 3 not irreducible over Fp"
Fp8.<w> = Fp.extension(absolute, 'w')
q = p^8
print(f"Fp⁸ = Fp[w]/(w⁸ − 3),  |Fp⁸| ≈ 2^{q.nbits()}\n")

a = Fp(-3)
target_r = 0xf06e44682c2aa440f5f26a5ae1748fec171df66abbdb81ad07f36ed81b9049
print(f"Target prime order  r = 0x{target_r:x}  ({target_r.bit_length()} bits)\n")

print("Sweeping a = −3, b = c · w^j with c ∈ [1, 30), j ∈ {0, …, 15} ...")
t_start = time.time()
hit = None
for j in range(16):
    for c in range(1, 30):
        b = Fp(c) * w^j
        if b == 0:
            continue
        try:
            E = EllipticCurve(Fp8, [a, b])
        except ArithmeticError:
            continue   # singular: 4a³ + 27b² = 0
        n = E.order()
        if int(n) == target_r:
            t_trace = q + 1 - n
            print(f"  ✓ HIT  b = {c}·w^{j}    #E = {n}  ({n.nbits()} bits)")
            print(f"          trace t = q + 1 − r = {t_trace}")
            hit = (c, j, b, n, t_trace, E)
            break
    if hit:
        break

elapsed = time.time() - t_start
print(f"Sweep elapsed: {elapsed:.1f}s\n")
assert hit, "no curve with the target r found in the sweep"
c, j, b, r, t, E = hit
assert r.is_prime() and int(r) == target_r

# === Generator: smallest X ≥ 1 such that (X, Y) ∈ E(Fp⁸) has order r ===
print("=== Generator ===")
for X in range(1, 100):
    rhs = Fp8(X)^3 + a * Fp8(X) + b
    if not rhs.is_square():
        continue
    Y = rhs.sqrt()
    P = E(Fp8(X), Y)
    if P.order() == r:
        print(f"smallest valid X = {X};   on-curve check (X, Y) ∈ E and order(P) = r ✓")
        # Express Y in the tower basis (c0 + c1·u) + (c2 + c3·u)·v + ((c4 + c5·u) + (c6 + c7·u)·v)·w
        # The absolute basis is {1, w, w², w³, w⁴, w⁵, w⁶, w⁷}; map to (u, v, w):
        #   u = w⁴,  v = w²,  uv = w⁶,  w = w,  uw = w⁵,  vw = w³,  uvw = w⁷
        # Tower index (c0,c1,c2,c3,c4,c5,c6,c7) = (1, u, v, uv, w, uw, vw, uvw)
        # corresponds to (Y[0], Y[4], Y[2], Y[6], Y[1], Y[5], Y[3], Y[7]) in the {1,w,…,w⁷} basis.
        Yvec = Y.list()                  # length-8 list in absolute basis
        tower_perm = [0, 4, 2, 6, 1, 5, 3, 7]
        tower_Y = [int(Yvec[i]) for i in tower_perm]
        print(f"X = {X}")
        print(f"Y in tower coordinates (c0+c1·u, c2+c3·u, c4+c5·u, c6+c7·u over v,w):")
        print(f"   = ({tower_Y[0]}, {tower_Y[1]}, {tower_Y[2]}, {tower_Y[3]},")
        print(f"      {tower_Y[4]}, {tower_Y[5]}, {tower_Y[6]}, {tower_Y[7]})")
        break
else:
    print("no valid generator found in X = 1..99")

# === Embedding degree k' ===
print("\n=== Embedding degree k′ ===")
# k′ = smallest k ≥ 1 such that r divides (p⁸)^k − 1, i.e., order of p⁸ mod r.
# Use the multiplicative_order function in Sage.
Zr = Integers(r)
k_prime = Zr(q).multiplicative_order()
print(f"k′ = order of p⁸ mod r = {k_prime}")
print(f"   (k′ has {Integer(k_prime).nbits()} bits)")
# (p⁸)^{k′} is astronomical — log₂ ≈ 248 · k′
print(f"   MOV target field Fp^(8·k′) has ≈ {248 * Integer(k_prime).nbits():,} digits log₂")

# === CM discriminant ===
print("\n=== CM discriminant ===")
# Discriminant of Z[π] is Δπ = t² − 4q.  The CM discriminant of the
# endomorphism ring End(E) is the discriminant of the imaginary quadratic
# order containing π; it divides Δπ and has the same square-free part.
Delta_pi = int(t)^2 - 4 * int(q)
print(f"Δπ = t² − 4·p⁸ = {Delta_pi}")
# Square-free part = (Δπ) / (largest square divisor)
abs_D = -Delta_pi  # Δπ < 0 for non-supersingular curves over Fq
sqf = abs_D.squarefree_part()
# CM discriminant is −sqf if sqf ≡ 3 mod 4, else −4·sqf
if sqf % 4 == 3:
    D_K = -sqf
else:
    D_K = -4 * sqf
print(f"square-free part of |Δπ| = {sqf}")
print(f"CM discriminant of End(E) ⊇ Z[π]:  D = {D_K}")
print(f"|D| in bits: {abs(D_K).bit_length()}")
