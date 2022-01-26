package mesh

import (
	"alvin.com/GoCarver/geom"
	"alvin.com/GoCarver/hmap"
)

type MeshSampler struct {
	mesh           *TriangleMesh
	cutterDiameter float64
	useBallCutter  bool
}

func NewMeshSamplerWithBallCutter(mesh *TriangleMesh, cutterDiameter float64) *MeshSampler {
	return &MeshSampler{
		mesh:           mesh,
		cutterDiameter: cutterDiameter,
		useBallCutter:  true,
	}
}

func NewMeshSamplerWithFlatCutter(mesh *TriangleMesh, cutterDiameter float64) *MeshSampler {
	return &MeshSampler{
		mesh:           mesh,
		cutterDiameter: cutterDiameter,
		useBallCutter:  false,
	}
}

var _ hmap.ScalarGridSampler = (*MeshSampler)(nil)

func (b *MeshSampler) GetNumSamplesFromX0ToX1(x0, x1 float64) int {
	// We use a zero-height footprint in the middle of the mesh to query the number
	// of triangles along the line between x0 and x1. We want one sample per cell in
	// the mesh, thus half the number of triangles.
	meshFootprint := b.mesh.GetMeshFootprint()
	midY := 0.5 * (meshFootprint.PMin.Y + meshFootprint.PMax.Y)
	queryFootprint := NewFootprint(geom.NewPt2(x0, midY), geom.NewPt2(x1, midY))
	numTrg := b.mesh.GetTrianglesUnderFootprint(queryFootprint)
	return numTrg.GetTriangleCount() / 2
}

func (b *MeshSampler) GetNumSamplesFromY0ToY1(y0, y1 float64) int {
	// We use a zero-height footprint in the middle of the mesh to query the number
	// of triangles along the line between y0 and y1. We want one sample per cell in
	// the mesh, thus half the number of triangles.
	meshFootprint := b.mesh.GetMeshFootprint()
	midX := 0.5 * (meshFootprint.PMin.X + meshFootprint.PMax.X)
	queryFootprint := NewFootprint(geom.NewPt2(midX, y0), geom.NewPt2(midX, y1))
	numTrg := b.mesh.GetTrianglesUnderFootprint(queryFootprint)
	return numTrg.GetTriangleCount() / 2
}

func (b *MeshSampler) At(p *geom.Pt2) float64 {
	return 0
}
