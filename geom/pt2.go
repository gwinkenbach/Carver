package geom

// Pt2 is a 2D point.
type Pt2 struct {
	X, Y float64
}

// NewPt2 returns a new Pt2.
func NewPt2(x, y float64) Pt2 {
	return Pt2{X: x, Y: y}
}

// NewPt2FromFloat32 returns a new Pt2 initialized with float32 values.
func NewPt2FromFloat32(x, y float32) Pt2 {
	return Pt2{X: float64(x), Y: float64(y)}
}

// Add returns p + v
func (p Pt2) Add(v Vec2) Pt2 {
	return Pt2{X: p.X + v.X, Y: p.Y + v.Y}
}

// SubV returns p - v
func (p Pt2) SubV(v Vec2) Pt2 {
	return Pt2{X: p.X - v.X, Y: p.Y - v.Y}
}

// Sub returns p - q
func (p Pt2) Sub(q Pt2) Vec2 {
	return Vec2{X: p.X - q.X, Y: p.Y - q.Y}
}

// Xform returns p * m, or p transformed by matrix m.
func (p Pt2) Xform(m *Matrix33) Pt2 {
	return Pt2{
		X: p.X*m.a[0][0] + p.Y*m.a[1][0] + m.a[2][0],
		Y: p.X*m.a[0][1] + p.Y*m.a[1][1] + m.a[2][1],
	}
}

// Eq returns whether point p is equal to (x, y).
func (p Pt2) Eq(x, y float64) bool {
	return p.X == x && p.Y == y
}
