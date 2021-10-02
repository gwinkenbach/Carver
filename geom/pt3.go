package geom

// Pt3 is a 3D point.
type Pt3 struct {
	X, Y, Z float64
}

// NewPt3 returns a new Pt3.
func NewPt3(x, y, z float64) Pt3 {
	return Pt3{X: x, Y: y, Z: z}
}

// NewPt3FromFloat32 returns a new Pt3 initialized with float32 values.
func NewPt3FromFloat32(x, y, z float32) Pt3 {
	return Pt3{X: float64(x), Y: float64(y), Z: float64(z)}
}

// Add returns p + v
func (p *Pt3) Add(v Vec3) Pt3 {
	return Pt3{X: p.X + v.X, Y: p.Y + v.Y, Z: p.Z + v.Z}
}

// SubV returns p - v
func (p *Pt3) SubV(v Vec3) Pt3 {
	return Pt3{X: p.X - v.X, Y: p.Y - v.Y, Z: p.Z - v.Z}
}

// Sub returns p - q
func (p *Pt3) Sub(q Pt3) Vec3 {
	return Vec3{X: p.X - q.X, Y: p.Y - q.Y, Z: p.Z - q.Z}
}

// Eq returns whether point p is equal to q.
func (p *Pt3) Eq(q Pt3) bool {
	return p.X == q.X && p.Y == q.Y && p.Z == q.Z
}

// EqXyz returns whether point p is equal to (x, y, z).
func (p *Pt3) EqXyz(x, y, z float64) bool {
	return p.X == x && p.Y == y && p.Z == z
}
