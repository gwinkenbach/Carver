package hmap

import (
	"math"
	"testing"

	_ "image/jpeg"
	_ "image/png"

	"alvin.com/GoCarver/geom"
	"alvin.com/GoCarver/util"
)

func TestPixelDepthSampler(t *testing.T) {

	img := util.LoadGray8Image("../images/various_grays.png")
	if img == nil {
		t.Fatalf("Could not load test image\n")
	}

	xform := geom.NewXformCache(128, 128, 128, 128, 0, 0, 128, 128, geom.ImageModeFill)
	sampler := NewPixelDepthSampler(
		xform.GetMc2NicXform(), geom.NewPt2(0, 0), geom.NewSize2(128, 128), img)

	q := geom.NewPt2(0, 128)
	d := sampler.At(&q)
	if math.Abs(d-1.0) > 0.0001 {
		t.Errorf("Expected at%v == 1.0, got %f\n", q, d)
	}

	q = geom.NewPt2(128, 128)
	d = sampler.At(&q)
	if math.Abs(d-0.501961) > 0.0001 {
		t.Errorf("Expected at%v == 0.501961, got %f\n", q, d)
	}

	q = geom.NewPt2(0, 0)
	d = sampler.At(&q)
	if math.Abs(d-0.752941) > 0.0001 {
		t.Errorf("Expected at%v == 0.752941, got %f\n", q, d)
	}

	q = geom.NewPt2(128, 0)
	d = sampler.At(&q)
	if math.Abs(d-0.247059) > 0.0001 {
		t.Errorf("Expected at%v == 0.247059, got %f\n", q, d)
	}

	q = geom.NewPt2(64, 64)
	d = sampler.At(&q)
	if d != 0 {
		t.Errorf("Expected at%v == 0, got %f\n", q, d)
	}
}
