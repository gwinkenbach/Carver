package mesh

import (
	"alvin.com/GoCarver/geom"
)

// Triangle is a generic interface for a triangle in the mesh.
type Triangle interface {
	Vertex(i int) geom.Pt3 // Access vertex V0, v1 or V2.
	UnitNormal() geom.Vec3 // Access unit normal vector.
}

// Concrete implementation of the Triangle interface.
type meshTriangle struct {
	normal   geom.Vec3
	vertices [3]geom.Pt3
}

var _ Triangle = (*meshTriangle)(nil)

func (m *meshTriangle) Vertex(i int) geom.Pt3 {
	switch {
	case i <= 0:
		return m.vertices[0]
	case i >= 2:
		return m.vertices[2]
	default:
		return m.vertices[1]
	}
}

// Return the triangle's unit normal vector.
func (m *meshTriangle) UnitNormal() geom.Vec3 {
	return m.normal
}
