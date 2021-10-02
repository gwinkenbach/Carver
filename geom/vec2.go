package geom

import "math"

// Vec2 is a 2D vector.
type Vec2 struct {
	X, Y float64
}

// NewVec2 returns a new Vec2.
func NewVec2(x, y float64) Vec2 {
	return Vec2{X: x, Y: y}
}

// NewVec2FromFloat32 returns a new Vec2 initialized with float32 values.
func NewVec2FromFloat32(x, y float32) Vec2 {
	return Vec2{X: float64(x), Y: float64(y)}
}

// Add returns v + w
func (v *Vec2) Add(w Vec2) Vec2 {
	return Vec2{X: v.X + w.X, Y: v.Y + w.Y}
}

// Sub returns v - w
func (v *Vec2) Sub(w Vec2) Vec2 {
	return Vec2{X: v.X - w.X, Y: v.Y - w.Y}
}

// Scale returns v * s
func (v *Vec2) Scale(s float64) Vec2 {
	return Vec2{X: v.X * s, Y: v.Y * s}
}

// Dot returns the dot product v * w
func (v *Vec2) Dot(w Vec2) float64 {
	return v.X*w.X + v.Y*w.Y
}

// Norm normalizes v in place. Fatal error if v has zero length.
// Returns v.
func (v *Vec2) Norm() Vec2 {
	l := math.Sqrt(v.Dot(*v))
	*v = v.Scale(1.0 / l)
	return *v
}

// LenSq returns the length of vector v squared.
func (v *Vec2) LenSq() float64 {
	return v.X*v.X + v.Y*v.Y
}

// Len returns the length of vector v.
func (v *Vec2) Len() float64 {
	return math.Sqrt(v.X*v.X + v.Y*v.Y)
}

// Xform returns v * m, or v transformed by matrix m.
func (v *Vec2) Xform(m *Matrix33) Vec2 {
	return Vec2{
		X: v.X*m.a[0][0] + v.Y*m.a[1][0],
		Y: v.X*m.a[0][1] + v.Y*m.a[1][1],
	}
}
