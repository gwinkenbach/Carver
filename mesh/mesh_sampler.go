package mesh

import (
	"log"
	"math"

	"alvin.com/GoCarver/geom"
	"alvin.com/GoCarver/hmap"
)

type MeshSampler struct {
	mesh               *TriangleMesh
	cutterRadius       float64
	useBallPointCutter bool
}

var _ hmap.ScalarGridSampler = (*MeshSampler)(nil)

// Sample the given triangle with a ballcutter tool whose footprint is toolFootprint.
// Return success=true if a contact point is found and set z to the height of the tip
// of the ballpoint tool. Otherwise, return success = false and z = 0.
func sampleTriangleWithBallpointTool(
	toolFootprint Footprint, trg Triangle) (success bool, z float64) {

	// We treat the ballpoint tool as a sphere of radius r = cutter-radius centered
	// at toolLoc. Then we look for a contact point with the triangle at distance r
	// from toolLoc. Note that contact may happen within the face of the triangle or
	// along any of the three edges.

	// Position the tool at the center line, above z max.
	toolRadius := 0.5 * toolFootprint.GetWidth()
	toolXY := toolFootprint.GetCenterPoint()
	toolLoc := geom.NewPt3(toolXY.X, toolXY.Y, 100.0)

	// Look for a contact point with the triangle at distance r = cutter-radius.
	ok, toolZ := dropPointPToDistanceRFromTrianglePlane(toolLoc, toolRadius, trg)
	if ok {
		// We found a touch point, but is it within the triangle?
		toolLoc.Z = toolZ
		contactPt := toolLoc.SubV(trg.UnitNormal().Scale(toolRadius))
		if isPlanePointWithinTriangle(contactPt, trg) {
			// We found a valid contact point with the triangle. Note that the tool's z height
			// is cutter-radius lower due to the rounded tip.
			return true, toolZ - toolRadius
		}
	}

	// The tool may still make contact with one of the triangle's edges.
	for i := 0; i < 3; i++ {
		vi, vj := edge(trg, i)
		w := vj.Sub(vi)
		ok, toolZ = dropPointPToDistanceRFromLine(toolLoc, toolRadius, vi, w)
		if ok {
			// We got a hit, but is it within the edge's endpoints. To this end, set v1 to
			// (toolLoc - vi) and project it onto w = vi+1 - vi. That projection must lie
			// between vi and vi+1.
			toolLoc.Z = toolZ
			v1 := toolLoc.Sub(vi)
			l1 := v1.Dot(w)
			a := l1 * l1 / w.LenSq() // Use squared value to avoid a square-root. Check negative l1 below.
			if l1 >= 0 && 0.0 <= a && a <= 1.0 {
				// As above, the actual z-height is cutter-radius lower than the computed toolLoc z value.
				return true, toolZ - toolRadius
			}
		}
	}

	return false, 0.0
}

// Utility function to return triangle's edge Vi, Vi+1, modulo 3.
func edge(trg Triangle, i int) (geom.Pt3, geom.Pt3) {
	if i < 0 || i > 2 {
		log.Fatalln("Invalid edge index")
		return geom.Pt3{}, geom.Pt3{}
	}

	j := i + 1
	if j > 2 {
		j = 0
	}
	return trg.Vertex(i), trg.Vertex(j)
}

func NewMeshSamplerWithBallCutter(mesh *TriangleMesh, cutterDiameter float64) *MeshSampler {
	return &MeshSampler{
		mesh:               mesh,
		cutterRadius:       0.5 * cutterDiameter,
		useBallPointCutter: true,
	}
}

func NewMeshSamplerWithFlatCutter(mesh *TriangleMesh, cutterDiameter float64) *MeshSampler {
	return &MeshSampler{
		mesh:               mesh,
		cutterRadius:       0.5 * cutterDiameter,
		useBallPointCutter: false,
	}
}

func (ms *MeshSampler) GetNumSamplesFromX0ToX1(x0, x1 float64) int {
	// We use a zero-height footprint in the middle of the mesh to query the number
	// of triangles along the line between x0 and x1. We want one sample per cell in
	// the mesh, thus half the number of triangles.
	meshFootprint := ms.mesh.GetMeshFootprint()
	midY := 0.5 * (meshFootprint.PMin.Y + meshFootprint.PMax.Y)
	queryFootprint := NewFootprint(geom.NewPt2(x0, midY), geom.NewPt2(x1, midY))
	numTrg := ms.mesh.GetTrianglesUnderFootprint(queryFootprint)
	return numTrg.GetTriangleCount() / 2
}

func (ms *MeshSampler) GetNumSamplesFromY0ToY1(y0, y1 float64) int {
	// We use a zero-height footprint in the middle of the mesh to query the number
	// of triangles along the line between y0 and y1. We want one sample per cell in
	// the mesh, thus half the number of triangles.
	meshFootprint := ms.mesh.GetMeshFootprint()
	midX := 0.5 * (meshFootprint.PMin.X + meshFootprint.PMax.X)
	queryFootprint := NewFootprint(geom.NewPt2(midX, y0), geom.NewPt2(midX, y1))
	numTrg := ms.mesh.GetTrianglesUnderFootprint(queryFootprint)
	return numTrg.GetTriangleCount() / 2
}

func (ms *MeshSampler) EnableInvertImage(enable bool) {
	// No-op. Invert sampler that is used to construct the mesh instead.
}

func (ms *MeshSampler) At(p geom.Pt2) float64 {
	if ms.useBallPointCutter {
		return ms.sampleBallPointToolAt(p)
	}

	// TODO: flat-cutter implementation
	log.Fatal("Not implemented")
	return 1.0
}

// Sample location p with a ballpoint tool and return the z-coordinate for the
// tip of the tool.
func (ms *MeshSampler) sampleBallPointToolAt(p geom.Pt2) float64 {
	toolFootprint := NewFootprint(
		geom.NewPt2(p.X-ms.cutterRadius, p.Y-ms.cutterRadius),
		geom.NewPt2(p.X+ms.cutterRadius, p.Y+ms.cutterRadius))
	zMin, zMax := ms.mesh.GetZExtents()
	triangles := ms.mesh.GetTrianglesUnderFootprint(toolFootprint)

	foundContact := false
	toolZ := 0.0
	t := triangles.Next()
	for ; t != nil; t = triangles.Next() {
		if ok, z := sampleTriangleWithBallpointTool(toolFootprint, t); ok {
			foundContact = true
			toolZ = math.Max(z, toolZ)
		}
	}

	if foundContact {
		// Sampling actually return a value between 0 and 1.
		return (toolZ - zMin) / (zMax - zMin)
	}

	return 1.0
}
