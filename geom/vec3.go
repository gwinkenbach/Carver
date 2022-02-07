package geom

import "math"

// Vec3 is a 3D vector.
type Vec3 struct {
	X, Y, Z float64
}

// NewVec3 returns a new Vec3.
func NewVec3(x, y, z float64) Vec3 {
	return Vec3{X: x, Y: y, Z: z}
}

// NewVec3FromFloat32 returns a new Vec3 initialized with float32 values.
func NewVec3FromFloat32(x, y, z float32) Vec3 {
	return Vec3{X: float64(x), Y: float64(y), Z: float64(z)}
}

// Add returns v + w
func (v Vec3) Add(w Vec3) Vec3 {
	return Vec3{X: v.X + w.X, Y: v.Y + w.Y, Z: v.Z + w.Z}
}

// Sub returns v - w
func (v Vec3) Sub(w Vec3) Vec3 {
	return Vec3{X: v.X - w.X, Y: v.Y - w.Y, Z: v.Z - w.Z}
}

// Scale returns v * s
func (v Vec3) Scale(s float64) Vec3 {
	return Vec3{X: v.X * s, Y: v.Y * s, Z: v.Z * s}
}

// Dot returns the dot product v * w
func (v Vec3) Dot(w Vec3) float64 {
	return v.X*w.X + v.Y*w.Y + v.Z*w.Z
}

// Cross returns the cross product v x w
func (v Vec3) Cross(w Vec3) Vec3 {
	return NewVec3(v.Y*w.Z-v.Z*w.Y, v.Z*w.X-v.X*w.Z, v.X*w.Y-v.Y*w.X)
}

// Norm normalizes v in place. Fatal error if v has zero length.
// Returns v.
func (v Vec3) Norm() Vec3 {
	l := math.Sqrt(v.Dot(v))
	v = v.Scale(1.0 / l)
	return v
}

// LenSq returns the length of vector v squared.
func (v Vec3) LenSq() float64 {
	return v.X*v.X + v.Y*v.Y + v.Z*v.Z
}

// Len returns the length of vector v.
func (v Vec3) Len() float64 {
	return math.Sqrt(v.X*v.X + v.Y*v.Y + v.Z*v.Z)
}

// Returns whether vector v is strictly equal to (x, y, z)
func (v Vec3) EqXyz(x, y, z float64) bool {
	return v.X == x && v.Y == y && v.Z == z
}
