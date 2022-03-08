package mesh

import (
	"log"
	"math"

	"alvin.com/GoCarver/geom"
)

// Move point p vertically until it's at distance r from the plane that embeds triangle trg.
// Returns the pair (success, pz) where success indicates whether the operations was successful or
// not. If it is successful, pz contains the new z-coordinate for point p. Note that the function
// fails when the plane is (nearly) vertical.
func dropPointPToDistanceRFromTrianglePlane(
	p geom.Pt3, r float64, trg Triangle) (success bool, pz float64) {

	n := trg.UnitNormal()
	if math.Abs(n.Z) < 1e-6 || r < 0 {
		return false, 0
	}

	q := trg.Vertex(0)
	pxy := geom.NewPt2(p.X, p.Y)
	qxy := geom.NewPt2(q.X, q.Y)
	nxy := geom.NewVec2(n.X, n.Y)

	d := math.Abs(pxy.Sub(qxy).Dot(nxy))
	pz = (r-d)/n.Z + q.Z
	success = true
	return
}

// Project the point p onto the plane embedding triangle trg and return the resulting point.
func projectPointToTrianglePlane(p geom.Pt3, trg Triangle) geom.Pt3 {
	n := trg.UnitNormal()
	q1 := trg.Vertex(0)
	v := p.Sub(q1)
	d := v.Dot(n)
	return p.SubV(n.Scale(d))
}

// Plane p is known to be on the plane embedding triangle trg. Returns whether it is within the
// bounds of the triangle.
func isPlanePointWithinTriangle(p geom.Pt3, trg Triangle) bool {
	// We take advantage of how the mesh triangles are built to optimize this test.
	// First, we can ignore the z-height and work exclusively in XY plane.
	p1 := geom.NewPt2(p.X, p.Y)
	q0 := geom.NewPt2(trg.Vertex(0).X, trg.Vertex(0).Y)
	q1 := geom.NewPt2(trg.Vertex(1).X, trg.Vertex(1).Y)
	q2 := geom.NewPt2(trg.Vertex(2).X, trg.Vertex(2).Y)

	// Second we know that each triangle has a horizontal and a vertical edge. Let's determine which
	// of the two triangles within the mesh cell we're working with.
	if q0.X == q1.X && q1.Y == q2.Y {
		// The top-left triangle: q1 +-+ q2
		//                           |/
		//                        q0 +
		if p1.Y > q1.Y || p1.X < q0.X {
			return false // Above or to the left of the triangle.
		}

		// Check what side of diagonal edge p1 is on.
		w := geom.NewVec2(q0.Y-q2.Y, q2.X-q0.X) // w is perpendicular to diagonal pointing inside trg.
		v := p1.Sub(q2)
		// l := v.Dot(w)
		// fmt.Printf("Check side w=%v, v=%v, l=%f\n", w, v, l)
		return v.Dot(w) >= 0
	}

	if q1.X == q2.X && q2.Y == q0.Y {
		// The bottom-right triangle.   + q1
		//                             /|
		//                         q0 +-+ q2
		if p1.Y < q0.Y || p1.X > q2.X {
			return false // Bellow or to the right of the triangle.
		}

		// Check what side of diagonal edge p1 is on.
		w := geom.NewVec2(q1.Y-q0.Y, q0.X-q1.X) // w is perpendicular to diagonal pointing inside trg.
		v := p1.Sub(q1)
		return v.Dot(w) >= 0
	}

	// If we get down here it means the triangle configuration has changed in
	// the mesh. The code most likely needs to be updated accordingly.
	log.Fatal("Unexpected triangle configuration.")
	return false
}

// Project point p onto the line passing through q1 and q2 and returns the result. In case of a
// degenerate line (q1 == q2) returns q1.
func projectPointToLine(p geom.Pt3, q1, q2 geom.Pt3) geom.Pt3 {
	w := q2.Sub(q1)
	d := w.LenSq()
	if d < 1e-6 {
		// Coincident points q1 and q2, i.e. degenerate line.
		return q1
	}

	v := p.Sub(q1)
	s := v.Dot(w) / d
	return q1.Add(w.Scale(s))
}

// Return whether point p, assumed to lie on the line through q1 and q2, lies on or between
// the two points.
func isLinePointOnSegment(p geom.Pt3, q1, q2 geom.Pt3) bool {
	v := p.Sub(q1)
	w := q2.Sub(q1)

	// Vector v can't be longer than w.
	if v.LenSq() > w.LenSq() {
		return false
	}

	// Vectors v and w must have the same orientation.
	return v.Dot(w) >= 0
}

// Let L be the line defined by point q and direction vector w. Move point p along
// the z-direction until it is at distance R from L if possible. If successful,
// return success = true and set pz to the new Z-coordinate for point p.
func dropPointPToDistanceRFromLine(
	p geom.Pt3, r float64, q geom.Pt3, w geom.Vec3) (success bool, pz float64) {

	// Let p1 = p + lambda * n, where n = (0 0 1). Let q1 be the point on the line defined by q
	// and w that is closest to p1. We want to find lambda such that (p1 - q1) * (p1 - q1) = r^2,
	// if possible. There is a quadratic close-form solution to this problem that leads
	// to the following code.
	v := p.Sub(q)
	l := w.LenSq()

	// fmt.Printf("\nv = %v, w = %v, l = %f\n", v, w, l)

	if math.Abs(l) < 1e-6 {
		log.Fatalln("Can't define a line with a 0-length vector")
		return false, 0
	}

	m := v.Dot(w) / l
	s := w.Z / l

	k_x := v.X - w.X*m // v.X * (1 - w.X*w.X/l)
	k_y := v.Y - w.Y*m
	k_z := v.Z - w.Z*m

	// fmt.Printf("kx, ky, kz = %4f, %4f, %4f\n", k_x, k_y, k_z)

	t_x := -s * w.X
	t_y := -s * w.Y
	t_z := 1.0 - s*w.Z

	// fmt.Printf("tz = %4f\n", t_z)

	a := t_x*t_x + t_y*t_y + t_z*t_z
	b := 2.0 * (t_x*k_x + t_y*k_y + t_z*k_z)
	c := k_x*k_x + k_y*k_y + k_z*k_z - r*r

	// fmt.Printf("a = %4f, b = %4f, c = %4f\n", a, b, c)

	D := b*b - 4*a*c
	if D < 0 {
		// fmt.Printf("D = %4f\n", D)
		return false, 0
	}
	if math.Abs(a) < 1e-6 {
		// fmt.Printf("a == 0\n")
		return false, 0
	}

	den := 1.0 / (2.0 * a)
	D = math.Sqrt(D)
	lambda1 := (-b + D) * den
	lambda2 := (-b - D) * den

	// fmt.Printf("lambda1,2 = %4f, %4f\n", lambda1, lambda2)

	z1 := p.Z + lambda1 // a.k.a z-coord. for p + lambda n
	z2 := p.Z + lambda2

	// We want the highest z.
	return true, math.Max(z1, z2)
}
