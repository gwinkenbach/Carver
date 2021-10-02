package geom

import "testing"

func TestXformMc2NicModeFill(t *testing.T) {
	xfc := NewXformCache(128, 256, 64, 128, 32, 64, 256, 512, ImageModeFill)
	xf := xfc.GetMc2NicXform()

	p1 := NewPt2(32, 64)
	q := p1.Xform(xf)
	if !q.Eq(0, 0) {
		t.Errorf("Xform GetMc2NicXform: got %v, expected (0, 0)\n", q)
	}

	p1 = NewPt2(96, 192)
	q = p1.Xform(xf)
	if !q.Eq(1, 1) {
		t.Errorf("Xform GetMc2NicXform: got %v, expected (1, 1)\n", q)
	}

	p1 = NewPt2(64, 128)
	q = p1.Xform(xf)
	if !q.Eq(0.5, 0.5) {
		t.Errorf("Xform GetMc2NicXform: got %v, expected (0.5, 0.5)\n", q)
	}
}

func TestXformMc2NicModeFit(t *testing.T) {
	xfc := NewXformCache(128, 256, 64, 128, 32, 64, 256, 256, ImageModeFit)
	xf := xfc.GetMc2NicXform()

	p1 := NewPt2(32, 96)
	q := p1.Xform(xf)
	if !q.Eq(0, 0) {
		t.Errorf("Xform GetMc2NicXform: got %v, expected (0, 0)\n", q)
	}

	p1 = NewPt2(96, 160)
	q = p1.Xform(xf)
	if !q.Eq(1, 1) {
		t.Errorf("Xform GetMc2NicXform: got %v, expected (1, 1)\n", q)
	}

	p1 = NewPt2(64, 128)
	q = p1.Xform(xf)
	if !q.Eq(0.5, 0.5) {
		t.Errorf("Xform GetMc2NicXform: got %v, expected (0.5, 0.5)\n", q)
	}
}

func TestXformMc2NicModeCrop(t *testing.T) {
	xfc := NewXformCache(128, 256, 64, 128, 32, 64, 256, 256, ImageModeCrop)
	xf := xfc.GetMc2NicXform()

	p1 := NewPt2(0, 64)
	q := p1.Xform(xf)
	if !q.Eq(0, 0) {
		t.Errorf("Xform GetMc2NicXform: got %v, expected (0, 0)\n", q)
	}

	p1 = NewPt2(128, 192)
	q = p1.Xform(xf)
	if !q.Eq(1, 1) {
		t.Errorf("Xform GetMc2NicXform: got %v, expected (1, 1)\n", q)
	}

	p1 = NewPt2(64, 128)
	q = p1.Xform(xf)
	if !q.Eq(0.5, 0.5) {
		t.Errorf("Xform GetMc2NicXform: got %v, expected (0.5, 0.5)\n", q)
	}
}
