package geom

// Size2 is a 2-dimensional size object with width and height.
type Size2 struct {
	W float64
	H float64
}

// ToVec2 returns a Vec2 object with same w,h as size objwct.
func (s *Size2) ToVec2() Vec2 {
	return Vec2{X: s.W, Y: s.H}
}

// NewSize2 creates and returns a new Size2 object.
func NewSize2(w, h float64) Size2 {
	return Size2{W: w, H: h}
}

// NewSize2FromFloat32 creates and returns a new Size2 object from float32 values.
func NewSize2FromFloat32(w, h float32) Size2 {
	return Size2{W: float64(w), H: float64(h)}
}

// NewSize2FromVec2 creates and returns a new Size2 object from a Vec2.
func NewSize2FromVec2(v Vec2) Size2 {
	return Size2{W: v.X, H: v.Y}
}
