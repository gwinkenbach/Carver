package carving

import (
	"testing"
)

func TestRunXForward(t *testing.T) {
	r := xCarvingRun{}
	sampler := newConstantDepthSampler(0)
	gen := unitTestGenerator{}

	r.configure(&sampler, &gen, 100, 10, 0, 0, -0.1, 0.2)
	if r.isDone() {
		t.Errorf("xCarvingRun: should not be Done right after configure.\n")
	}

	r.doOnePass(1)

	if !gen.pathCompleted {
		t.Errorf("xCarvingRun: path left open after one pass\n")
	}

	q := gen.firstPoint
	if q.X != 10 || q.Y != 0 {
		t.Errorf("xCarvingRun: 1st path point should be (10, 0) got %v\n", q)
	}

	q = gen.lastPoint
	if q.X != 110 || q.Y != 0 {
		t.Errorf("xCarvingRun: last path point should be (110, 0) got %v\n", q)
	}

	if gen.lastDepth != -0.1 {
		t.Errorf("xCarvingRun: expected depth should be -0.1 got %f\n", gen.lastDepth)
	}

	if gen.numPoints != 99 {
		t.Errorf("xCarvingRun: expected 99 points along path, got %d\n", gen.numPoints)
	}

	if !r.isDone() {
		t.Errorf("xCarvingRun: should be Done after one pass.\n")
	}
}

func TestRunXBackward(t *testing.T) {
	r := xCarvingRun{}
	sampler := newConstantDepthSampler(0)
	gen := unitTestGenerator{}

	r.configure(&sampler, &gen, 100, 10, 0, 0, -0.1, 0.2)
	if r.isDone() {
		t.Errorf("xCarvingRun: should not be Done right after configure.\n")
	}

	r.doOnePass(-1)

	if !gen.pathCompleted {
		t.Errorf("xCarvingRun: path left open after one pass\n")
	}

	q := gen.firstPoint
	if q.X != 110 || q.Y != 0 {
		t.Errorf("xCarvingRun: 1st path point should be (110, 0) got %v\n", q)
	}

	q = gen.lastPoint
	if q.X != 10 || q.Y != 0 {
		t.Errorf("xCarvingRun: last path point should be (10, 0) got %v\n", q)
	}

	if gen.lastDepth != -0.1 {
		t.Errorf("xCarvingRun: expected depth should be -0.1 got %f\n", gen.lastDepth)
	}

	if gen.numPoints != 99 {
		t.Errorf("xCarvingRun: expected 99 points along path, got %d\n", gen.numPoints)
	}

	if !r.isDone() {
		t.Errorf("xCarvingRun: should be Done after one pass.\n")
	}
}

func TestRunXMultipass(t *testing.T) {
	r := xCarvingRun{}
	sampler := newConstantDepthSampler(0)
	gen := unitTestGenerator{}

	r.configure(&sampler, &gen, 100, 10, 0, 0, -0.4, 0.25)
	if r.isDone() {
		t.Errorf("xCarvingRun: should not be Done right after configure.\n")
	}

	r.doOnePass(1)

	q := gen.firstPoint
	if q.X != 10 || q.Y != 0 {
		t.Errorf("xCarvingRun: 1st path point should be (10, 0) got %v\n", q)
	}

	if gen.lastDepth != -0.25 {
		t.Errorf("xCarvingRun: expected depth should be -0.25 got %f\n", gen.lastDepth)
	}

	if r.isDone() {
		t.Errorf("xCarvingRun: another pass should be needed.\n")
	}

	r.doOnePass(-1)

	q = gen.lastPoint
	if q.X != 10 || q.Y != 0 {
		t.Errorf("xCarvingRun: last path point should be (10, 0) got %v\n", q)
	}

	if gen.lastDepth != -0.4 {
		t.Errorf("xCarvingRun: expected depth should be -0.4 got %f\n", gen.lastDepth)
	}

	if !r.isDone() {
		t.Errorf("xCarvingRun: only two passes should be needed.\n")
	}
}

func TestRunYForward(t *testing.T) {
	r := yCarvingRun{}
	sampler := newConstantDepthSampler(0)
	gen := unitTestGenerator{}

	r.configure(&sampler, &gen, 100, 10, 0, 0, -0.1, 0.2)
	if r.isDone() {
		t.Errorf("yCarvingRun: should not be Done right after configure.\n")
	}

	r.doOnePass(1)

	if !gen.pathCompleted {
		t.Errorf("yCarvingRun: path left open after one pass\n")
	}

	q := gen.firstPoint
	if q.X != 0 || q.Y != 10 {
		t.Errorf("yCarvingRun: 1st path point should be (0, 10) got %v\n", q)
	}

	q = gen.lastPoint
	if q.X != 0 || q.Y != 110 {
		t.Errorf("yCarvingRun: last path point should be (0, 110) got %v\n", q)
	}

	if gen.lastDepth != -0.1 {
		t.Errorf("yCarvingRun: expected depth should be -0.1 got %f\n", gen.lastDepth)
	}

	if gen.numPoints != 99 {
		t.Errorf("yCarvingRun: expected 99 points along path, got %d\n", gen.numPoints)
	}

	if !r.isDone() {
		t.Errorf("yCarvingRun: should be Done after one pass.\n")
	}
}
