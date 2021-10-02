package geom

// Matrix33 is a 3x3 matrix of float64.
type Matrix33 struct {
	a [3][3]float64
}

// NewIndentityMatrix33 returns a new Identity matrix.
func NewIndentityMatrix33() Matrix33 {
	return Matrix33{
		a: [3][3]float64{{1, 0, 0}, {0, 1, 0}, {0, 0, 1}},
	}
}

// NewTranslateMatrix33 returns a new translation matrix.
func NewTranslateMatrix33(dx, dy float64) Matrix33 {
	return Matrix33{
		a: [3][3]float64{{1, 0, 0}, {0, 1, 0}, {dx, dy, 1}},
	}
}

// NewScaleMatrix33 returns a new scale matrix.
func NewScaleMatrix33(sx, sy float64) Matrix33 {
	return Matrix33{
		a: [3][3]float64{{sx, 0, 0}, {0, sy, 0}, {0, 0, 1}},
	}
}

// NewMatrix33 returns a new generic matrix with 2x2 upper-left and
// translation components specified.
func NewMatrix33(a00, a01, a10, a11, dx, dy float64) Matrix33 {
	return Matrix33{
		a: [3][3]float64{{a00, a01, 0}, {a10, a11, 0}, {dx, dy, 1}},
	}
}

// Copy returns a copy of matrix m.
func (m *Matrix33) Copy() Matrix33 {
	return Matrix33{a: m.a}
}

// Mul evaluates m = m * n in place and returns m.
func (m *Matrix33) Mul(n *Matrix33) *Matrix33 {
	m.a[0][0], m.a[0][1], m.a[0][2] =
		m.a[0][0]*n.a[0][0]+m.a[0][1]*n.a[1][0]+m.a[0][2]*n.a[2][0],
		m.a[0][0]*n.a[0][1]+m.a[0][1]*n.a[1][1]+m.a[0][2]*n.a[2][1],
		m.a[0][0]*n.a[0][2]+m.a[0][1]*n.a[1][2]+m.a[0][2]*n.a[2][2]

	m.a[1][0], m.a[1][1], m.a[1][2] =
		m.a[1][0]*n.a[0][0]+m.a[1][1]*n.a[1][0]+m.a[1][2]*n.a[2][0],
		m.a[1][0]*n.a[0][1]+m.a[1][1]*n.a[1][1]+m.a[1][2]*n.a[2][1],
		m.a[1][0]*n.a[0][2]+m.a[1][1]*n.a[1][2]+m.a[1][2]*n.a[2][2]

	m.a[2][0], m.a[2][1], m.a[2][2] =
		m.a[2][0]*n.a[0][0]+m.a[2][1]*n.a[1][0]+m.a[2][2]*n.a[2][0],
		m.a[2][0]*n.a[0][1]+m.a[2][1]*n.a[1][1]+m.a[2][2]*n.a[2][1],
		m.a[2][0]*n.a[0][2]+m.a[2][1]*n.a[1][2]+m.a[2][2]*n.a[2][2]

	return m
}

// Get returns component a_ij.
func (m *Matrix33) Get(i, j int) float64 {
	return m.a[i][j]
}

// Set sets component a_ij = v.
func (m *Matrix33) Set(i, j int, v float64) {
	m.a[i][j] = v
}
