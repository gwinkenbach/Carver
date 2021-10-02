package geom

import "testing"

func TestMatrix(t *testing.T) {
	p := NewPt2(1, 1)

	m1 := NewTranslateMatrix33(5, -5)
	q := p.Xform(&m1)
	if !q.Eq(6, -4) {
		t.Errorf("NewTranslateMatrix33: got %v, expected (6, -4)\n", q)
	}

	m2 := NewScaleMatrix33(2, -3)
	q = p.Xform(&m2)
	if !q.Eq(2, -3) {
		t.Errorf("NewScaleMatrix33: got %v, expected (2, -3)\n", q)
	}

	i := NewIndentityMatrix33()
	q = p.Xform(&i)
	if !q.Eq(1, 1) {
		t.Errorf("NewIndentityMatrix33: got %v, expected (1, 1)\n", q)
	}

	m := m1.Mul(&m2)
	q = p.Xform(m)
	if !q.Eq(12, 12) {
		t.Errorf("Matrix33 Mul: got %v, expected (12, 12)\n", q)
	}

	c := m.Copy()
	q = p.Xform(&c)
	if !q.Eq(12, 12) {
		t.Errorf("Matrix33 Copy: got %v, expected (12, 12)\n", q)
	}

	h := NewMatrix33(0, -1, 1, 0, 1, 1)
	q = p.Xform(&h)
	if !q.Eq(2, 0) {
		t.Errorf("NewMatrix33: got %v, expected (2, 0)\n", q)
	}
}
