package geom

import "testing"

func TestVec2(t *testing.T) {
	v1 := NewVec2(0, 0)
	v2 := NewVec2(10, 10)

	v := v1.Add(v2)
	if v.X != 10 || v.Y != 10 {
		t.Errorf("Vec2 add: got %v, want (10, 10)\n", v)
	}

	v = v1.Sub(v2)
	if v.X != -10 || v.Y != -10 {
		t.Errorf("Vec2 sub: got %v, want (-10, -10)\n", v)
	}

	v = v2.Scale(10)
	if v.X != 100 || v.Y != 100 {
		t.Errorf("Vec2 scale: got %v, want (100, 100)\n", v)
	}

	l := v2.Dot(v2)
	if l != 200 {
		t.Errorf("Vec2 dot: got %v, want 200\n", l)
	}

	v3 := NewVec2(100, 0)
	v = v3.Norm()
	if v.X != 1 || v.Y != 0 {
		t.Errorf("Vec2 Norm: got %v, want (1, 0)\n", v)
	}

	m := NewScaleMatrix33(2, -4)
	v = v2.Xform(&m)
	if v.X != 20 || v.Y != -40 {
		t.Errorf("Vec2 Xform: got %v, want (20, -40)\n", v)
	}
}
