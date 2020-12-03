package bls12377

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"testing"
)

func (g *G1) one() *PointG1 {
	return g.New().Set(&g1One)
}

func (g *G1) rand() *PointG1 {
	p := &PointG1{}
	z, _ := new(fe).rand(rand.Reader)
	z6, bz6 := new(fe), new(fe)
	square(z6, z)
	square(z6, z6)
	mul(z6, z6, z)
	mul(z6, z6, z)
	mul(bz6, z6, b)
	for {
		x, _ := new(fe).rand(rand.Reader)
		y := new(fe)
		square(y, x)
		mul(y, y, x)
		add(y, y, bz6)
		if sqrt(y, y) {
			p.Set(&PointG1{*x, *y, *z})
			break
		}
	}
	if !g.IsOnCurve(p) {
		panic("rand point must be on curve")
	}
	if g.InCorrectSubgroup(p) {
		panic("rand point must be out of correct subgroup")
	}
	return p
}

func (g *G1) randCorrect() *PointG1 {
	return g.ClearCofactor(g.rand())
}

func (g *G1) randAffine() *PointG1 {
	return g.Affine(g.randCorrect())
}

func (g *G1) new() *PointG1 {
	return g.Zero()
}

func TestG1Serialization(t *testing.T) {
	var err error
	g1 := NewG1()
	zero := g1.Zero()
	b0 := g1.ToBytes(zero)
	p0, err := g1.FromBytes(b0)
	if err != nil {
		t.Fatal(err)
	}
	if !g1.IsZero(p0) {
		t.Fatal("infinity serialization failed")
	}
	for i := 0; i < fuz; i++ {
		a := g1.randAffine()
		_ = a
		uncompressed := g1.ToBytes(a)
		b, err := g1.FromBytes(uncompressed)
		if err != nil {
			t.Fatal(err)
		}
		if !g1.Equal(a, b) {
			t.Fatal("serialization failed")
		}
	}
}

func TestG1IsOnCurve(t *testing.T) {
	g := NewG1()
	zero := g.Zero()
	if !g.IsOnCurve(zero) {
		t.Fatal("zero must be on curve")
	}
	one := new(fe).one()
	p := &PointG1{*one, *one, *one}
	if g.IsOnCurve(p) {
		t.Fatal("(1, 1) is not on curve")
	}
}
func TestG1BatchAffine(t *testing.T) {
	n := 20
	g := NewG1()
	points0 := make([]*PointG1, n)
	points1 := make([]*PointG1, n)
	for i := 0; i < n; i++ {
		points0[i] = g.rand()
		points1[i] = g.New().Set(points0[i])
		if g.IsAffine(points0[i]) {
			t.Fatal("expect non affine point")
		}
	}
	g.AffineBatch(points0)
	for i := 0; i < n; i++ {
		if !g.Equal(points0[i], points1[i]) {
			t.Fatal("batch affine failed")
		}
	}
}

func TestG1AdditiveProperties(t *testing.T) {
	g := NewG1()
	t0, t1 := g.New(), g.New()
	zero := g.Zero()
	for i := 0; i < fuz; i++ {
		a, b := g.rand(), g.rand()
		g.Add(t0, a, zero)
		if !g.Equal(t0, a) {
			t.Fatal("a + 0 == a")
		}
		g.Add(t0, zero, zero)
		if !g.Equal(t0, zero) {
			t.Fatal("0 + 0 == 0")
		}
		g.Sub(t0, a, zero)
		if !g.Equal(t0, a) {
			t.Fatal("a - 0 == a")
		}
		g.Sub(t0, zero, zero)
		if !g.Equal(t0, zero) {
			t.Fatal("0 - 0 == 0")
		}
		g.Neg(t0, zero)
		if !g.Equal(t0, zero) {
			t.Fatal("- 0 == 0")
		}
		g.Sub(t0, zero, a)
		g.Neg(t0, t0)
		if !g.Equal(t0, a) {
			t.Fatal(" - (0 - a) == a")
		}
		g.Double(t0, zero)
		if !g.Equal(t0, zero) {
			t.Fatal("2 * 0 == 0")
		}
		g.Double(t0, a)
		g.Sub(t0, t0, a)
		if !g.Equal(t0, a) || !g.IsOnCurve(t0) {
			t.Fatal(" (2 * a) - a == a")
		}
		g.Add(t0, a, b)
		g.Add(t1, b, a)
		if !g.Equal(t0, t1) {
			t.Fatal("a + b == b + a")
		}
		g.Sub(t0, a, b)
		g.Sub(t1, b, a)
		g.Neg(t1, t1)
		if !g.Equal(t0, t1) {
			t.Fatal("a - b == - ( b - a )")
		}
		c := g.rand()
		g.Add(t0, a, b)
		g.Add(t0, t0, c)
		g.Add(t1, a, c)
		g.Add(t1, t1, b)
		if !g.Equal(t0, t1) {
			t.Fatal("(a + b) + c == (a + c ) + b")
		}
		g.Sub(t0, a, b)
		g.Sub(t0, t0, c)
		g.Sub(t1, a, c)
		g.Sub(t1, t1, b)
		if !g.Equal(t0, t1) {
			t.Fatal("(a - b) - c == (a - c) -b")
		}
	}
}

func TestG1MixedAdd(t *testing.T) {
	g := NewG1()

	t0, a := g.New(), g.rand()
	zero := g.Zero()

	g.AddMixed(t0, a, zero)
	if !g.Equal(t0, a) {
		t.Fatal("a + 0 == a")
	}
	g.AddMixed(a, t0, zero)
	if !g.Equal(t0, a) {
		t.Fatal("a + 0 == a")
	}
	g.Add(t0, zero, zero)
	if !g.Equal(t0, zero) {
		t.Fatal("0 + 0 == 0")
	}

	for i := 0; i < fuz; i++ {
		a, b := g.rand(), g.rand()
		if g.IsAffine(a) || g.IsAffine(b) {
			t.Fatal("expect non affine points")
		}
		bAffine := g.New().Set(b)
		g.Affine(bAffine)
		r0, r1 := g.New(), g.New()
		g.Add(r0, a, b)
		g.AddMixed(r1, a, bAffine)
		if !g.Equal(r0, r1) {
			t.Fatal("mixed addition failed")
		}
		aAffine := g.New().Set(a)
		g.Affine(aAffine)
		g.AddMixed(r0, a, aAffine)
		g.Double(r1, a)
		if !g.Equal(r0, r1) {
			t.Fatal("mixed addition must double where points are equal")
		}
	}
}

func TestG1MultiplicationCross(t *testing.T) {
	g := NewG1()
	for i := 0; i < fuz; i++ {

		a := g.randCorrect()
		s, _ := new(Fr).Rand(rand.Reader)
		sBig := s.ToBig()
		res0, res1, res2, res3, res4 := g.New(), g.New(), g.New(), g.New(), g.New()

		g.mulScalar(res0, a, s)
		g.glvMulFr(res1, a, s)
		g.glvMulBig(res2, a, sBig)
		g.wnafMulFr(res3, a, s)
		g.wnafMulBig(res4, a, sBig)

		if !g.Equal(res0, res1) {
			t.Fatal("cross multiplication failed (glv, fr)", i)
		}
		if !g.Equal(res0, res2) {
			t.Fatal("cross multiplication failed (glv, big)", i)
		}
		if !g.Equal(res0, res3) {
			t.Fatal("cross multiplication failed (wnaf, fr)", i)
		}
		if !g.Equal(res0, res4) {
			t.Fatal("cross multiplication failed (wnaf, big)", i)
		}
	}
}

func TestG1MultiplicativeProperties(t *testing.T) {
	g := NewG1()
	t0, t1 := g.New(), g.New()
	zero := g.Zero()
	for i := 0; i < fuz; i++ {
		a := g.randCorrect()
		s1, _ := new(Fr).Rand(rand.Reader)
		s2, _ := new(Fr).Rand(rand.Reader)
		s3, _ := new(Fr).Rand(rand.Reader)
		sone := &Fr{1}
		g.MulScalar(t0, zero, s1)
		if !g.Equal(t0, zero) {
			t.Fatal(" 0 ^ s == 0")
		}
		g.MulScalar(t0, a, sone)
		if !g.Equal(t0, a) {
			t.Fatal(" a ^ 1 == a")
		}
		g.MulScalar(t0, zero, s1)
		if !g.Equal(t0, zero) {
			t.Fatal(" 0 ^ s == a")
		}
		g.MulScalar(t0, a, s1)
		g.MulScalar(t0, t0, s2)
		s3.Mul(s1, s2)
		g.MulScalar(t1, a, s3)
		if !g.Equal(t0, t1) {
			t.Fatal(" (a ^ s1) ^ s2 == a ^ (s1 * s2)")
		}
		g.MulScalar(t0, a, s1)
		g.MulScalar(t1, a, s2)
		g.Add(t0, t0, t1)
		s3.Add(s1, s2)
		g.MulScalar(t1, a, s3)
		if !g.Equal(t0, t1) {
			t.Fatal(" (a ^ s1) + (a ^ s2) == a ^ (s1 + s2)")
		}
	}
}

func TestG1MultiExpExpected(t *testing.T) {
	g := NewG1()
	one := g.one()
	var scalars [2]*Fr
	var bases [2]*PointG1
	scalars[0] = &Fr{2}
	scalars[1] = &Fr{3}
	bases[0], bases[1] = new(PointG1).Set(one), new(PointG1).Set(one)
	expected, result := g.New(), g.New()
	g.mulScalar(expected, one, &Fr{5})
	_, _ = g.MultiExp(result, bases[:], scalars[:])
	if !g.Equal(expected, result) {
		t.Fatal("multi-exponentiation failed")
	}
}

func TestG1MultiExpBigExpected(t *testing.T) {
	g := NewG1()
	one := g.one()
	var scalars [2]*big.Int
	var bases [2]*PointG1
	scalars[0] = big.NewInt(2)
	scalars[1] = big.NewInt(3)
	bases[0], bases[1] = new(PointG1).Set(one), new(PointG1).Set(one)
	expected, result := g.New(), g.New()
	g.mulScalarBig(expected, one, big.NewInt(5))
	_, _ = g.MultiExpBig(result, bases[:], scalars[:])
	if !g.Equal(expected, result) {
		t.Fatal("multi-exponentiation failed")
	}
}

func TestG1ClearCofactor(t *testing.T) {
	g := NewG1()
	for i := 0; i < fuz; i++ {
		p0 := g.rand()
		if g.InCorrectSubgroup(p0) {
			t.Fatal("rand point should be out of correct subgroup")
		}
		g.ClearCofactor(p0)
		if !g.InCorrectSubgroup(p0) {
			t.Fatal("cofactor clearing is failed")
		}
	}
}

func BenchmarkG1Add(t *testing.B) {
	g1 := NewG1()
	a, b, c := g1.rand(), g1.rand(), PointG1{}
	t.ResetTimer()
	for i := 0; i < t.N; i++ {
		g1.Add(&c, a, b)
	}
}

func BenchmarkG1MulWNAF(t *testing.B) {
	g := NewG1()
	p := new(PointG1).Set(&g1One)
	s, _ := new(Fr).Rand(rand.Reader)
	sBig := s.ToBig()
	res := new(PointG1)
	t.Run("Naive", func(t *testing.B) {
		t.ResetTimer()
		for i := 0; i < t.N; i++ {
			g.mulScalar(res, p, s)
		}
	})
	for i := 1; i < 8; i++ {
		wnafMulWindowG1 = uint(i)
		t.Run(fmt.Sprintf("Fr, window: %d", i), func(t *testing.B) {
			t.ResetTimer()
			for i := 0; i < t.N; i++ {
				g.wnafMulFr(res, p, s)
			}
		})
		t.Run(fmt.Sprintf("Big, window: %d", i), func(t *testing.B) {
			t.ResetTimer()
			for i := 0; i < t.N; i++ {
				g.wnafMulBig(res, p, sBig)
			}
		})
	}
}

func BenchmarkG1MulGLV(t *testing.B) {

	g := NewG1()
	p := new(PointG1).Set(&g1One)
	s, _ := new(Fr).Rand(rand.Reader)
	sBig := s.ToBig()
	res := new(PointG1)
	t.Run("Naive", func(t *testing.B) {
		t.ResetTimer()
		for i := 0; i < t.N; i++ {
			g.mulScalar(res, p, s)
		}
	})
	for i := 1; i < 8; i++ {
		glvMulWindowG1 = uint(i)
		t.Run(fmt.Sprintf("Fr, window: %d", i), func(t *testing.B) {
			t.ResetTimer()
			for i := 0; i < t.N; i++ {
				g.glvMulFr(res, p, s)
			}
		})
		t.Run(fmt.Sprintf("Big, window: %d", i), func(t *testing.B) {
			t.ResetTimer()
			for i := 0; i < t.N; i++ {
				g.glvMulBig(res, p, sBig)
			}
		})
	}
}

func BenchmarkG1MultiExp(t *testing.B) {
	g := NewG1()
	v := func(n int) ([]*PointG1, []*Fr) {
		bases := make([]*PointG1, n)
		scalars := make([]*Fr, n)
		var err error
		for i := 0; i < n; i++ {
			scalars[i], err = new(Fr).Rand(rand.Reader)
			if err != nil {
				t.Fatal(err)
			}
			bases[i] = g.randAffine()
		}
		return bases, scalars
	}
	for _, i := range []int{2, 10, 100, 1000} {
		t.Run(fmt.Sprint(i), func(t *testing.B) {
			bases, scalars := v(i)
			result := g.New()
			t.ResetTimer()
			for i := 0; i < t.N; i++ {
				_, _ = g.MultiExp(result, bases, scalars)
			}
		})
	}
}
