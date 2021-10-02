package geom

import "testing"

func TestPt2(t *testing.T) {
	p1 := NewPt2(3, -5)
	p2 := NewPt2(8, -4)
	v1 := NewVec2(7, 21)

	p := p1.Add(v1)
	if p.X != 10 || p.Y != 16 {
		t.Errorf("Pt2 add: got %v, want (10, 16)\n", p)
	}

	p = p1.SubV(v1)
	if p.X != -4 || p.Y != -26 {
		t.Errorf("Pt2 SubV: got %v, want (-4, -26)\n", p)
	}

	v := p1.Sub(p2)
	if v.X != -5 || v.Y != -1 {
		t.Errorf("Pt2 Sub: got %v, want (-5, -1)\n", v)
	}

	m := NewTranslateMatrix33(20, 30)
	p = p1.Xform(&m)
	if p.X != 23 || p.Y != 25 {
		t.Errorf("Pt2 Xform: got %v, want (23, 25)\n", p)
	}
}
