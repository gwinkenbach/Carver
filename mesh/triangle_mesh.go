package mesh

import (
	"log"
	"math"

	"alvin.com/GoCarver/geom"
	"alvin.com/GoCarver/hmap"
)

// gridRow: Rows of vertices form a grid; each square in that grid form 2 triangles.
// row i+1: +---+---+---+-- -
//          |  /|  /|  /|
//          | / | / | / |
//          |/  |/  |/  |
// row i:   +---+---+---+-- -
// Note: The mesh is layed out over a regular grid. N+1 x-coordinates are stored within a single
//       array in the mesh since they are shared by all vertices. (See TriangleMesh below.)
type gridRow struct {
	y       float64     // y-coordinate for this entire row
	z       []float64   // Z-coordinates for N+1 vertices along the x-axis.
	normals []geom.Vec3 // 2*N precomputed unit normals for each triangle formed by row i and i+1
}

// TriangleMesh is a regular NxM triangle mesh representing the height map built from
// a scalar height map (interface codeSampler).
type TriangleMesh struct {
	xyBox      Footprint // The mesh footprint.
	zMin, zMax float64   // Z extents
	rows       []gridRow // Rows along the y-axis.
	x          []float64 // X-coordinate for N+1 vertices along the x-axis.
}

// triangleArray implements interface TriangleIterator
type triangleArray struct {
	triangles     []meshTriangle
	iteratorIndex int
}

var _ TriangleIterator = (*triangleArray)(nil)

// NewTriangleMesh creates and returns a new triangle mesh generated from the min/max corners
// of the grid and a sampler. The number of vertices along x and y in the grid is determined
// by the number of samples that the sampler generate in these directions. (See interface
// iSampler.) A grid point is generated for each sample.
func NewTriangleMesh(
	pMin geom.Pt2, pMax geom.Pt2,
	zBlack, zWhite float64,
	sampler hmap.ScalarGridSampler) *TriangleMesh {

	if pMin.X == pMax.X || pMin.Y == pMax.Y {
		return nil
	}

	mesh := &TriangleMesh{}
	mesh.xyBox = NewFootprint(pMin, pMax)
	mesh.buildMesh(zBlack, zWhite, sampler)

	mesh.zMin = math.Min(zBlack, zWhite)
	mesh.zMax = math.Max(zBlack, zWhite)

	return mesh
}

// GetZExtents returns the mesh z-extents.
func (t *TriangleMesh) GetZExtents() (zMin, zMax float64) {
	zMin, zMax = t.zMin, t.zMax
	return
}

// GetNumTriangles returns the number of triangles in the mesh as pair of numbers (nX, nY)
// where nX is the number of triangles along X and nY is the number of triangle rows along Y.
func (t *TriangleMesh) GetNumTriangles() (nX int, nY int) {
	nY = len(t.rows) - 1
	nX = 0
	if nY > 0 {
		nX = 2 * (len(t.x) - 1)
	}
	return
}

// GetMeshFootprint returns the footprint for the entire mesh.
func (t *TriangleMesh) GetMeshFootprint() Footprint {
	return t.xyBox
}

// GetTriangle returns the triangle at index (iX, iY) where 0 <= iX < nX and
// 0 <= iY < nY. Triangle counts nX and nY are returned by GetNumTriangles.
func (t *TriangleMesh) GetTriangle(iX, iY int) Triangle {
	nX, nY := t.GetNumTriangles()
	if iX < 0 || iX >= nX || iY < 0 || iY >= nY {
		log.Fatal("Triangle indices out of range")
		return &meshTriangle{}
	}

	trg := &meshTriangle{}
	trg.normal = t.rows[iY].normals[iX]

	// Vertices and triangles are layed out as follows in the grid:
	// nY+1 +---+
	//      |  /|
	//   T0 | / | T1
	//      |/  |
	//   nY +---+
	//     nV   nV+1
	iT := iX & 0x01
	iV := iX / 2
	xLeft := t.x[iV]
	xRight := t.x[iV+1]
	yBottom := t.rows[iY].y
	yTop := t.rows[iY+1].y
	if iT == 0 {
		trg.vertices[0] = geom.NewPt3(xLeft, yBottom, t.rows[iY].z[iV])
		trg.vertices[1] = geom.NewPt3(xLeft, yTop, t.rows[iY+1].z[iV])
		trg.vertices[2] = geom.NewPt3(xRight, yTop, t.rows[iY+1].z[iV+1])
	} else {
		trg.vertices[0] = geom.NewPt3(xLeft, yBottom, t.rows[iY].z[iV])
		trg.vertices[1] = geom.NewPt3(xRight, yTop, t.rows[iY+1].z[iV+1])
		trg.vertices[2] = geom.NewPt3(xRight, yBottom, t.rows[iY].z[iV+1])
	}

	return trg
}

// GetFootprintForTriangle returns the footprint for triangle at indices (iX, iY).
func (t *TriangleMesh) GetFootprintForTriangle(iX, iY int) Footprint {
	nX, nY := t.GetNumTriangles()
	if iX < 0 || iX >= nX || iY < 0 || iY >= nY {
		log.Fatal("Triangle indices out of range")
		return Footprint{}
	}

	// Vertices and triangles are layed out as follows in the grid:
	// nY+1 +---+
	//      |   |
	//      |   |
	//      |   |
	//   nY +---+
	//     nV   nV+1
	f := Footprint{}
	nV := iX / 2
	f.PMin.X = t.x[nV]
	f.PMin.Y = t.rows[iY].y
	f.PMax.X = t.x[nV+1]
	f.PMax.Y = t.rows[iY+1].y

	return f
}

// GetTrianglesUnderFootprint gathers all the mesh triangles that are covered by the given footprint
// into an iterator. The footprint is considered to be a closed set when looking for the triangles.
// That is, triangles whose boundaries abut with the footprint are considered to be covered by the
// footprint. This function may return an empty iterator.
func (t *TriangleMesh) GetTrianglesUnderFootprint(f Footprint) TriangleIterator {
	iMinRow, iMaxRow := t.findRowsForFootprint(f)
	if iMinRow < 0 {
		return &triangleArray{}
	}

	iMinCol, iMaxCol := t.findColumnsForFootprint(f)
	if iMinCol < 0 {
		return &triangleArray{}
	}

	// TODO: check footprints that lie entirely on one side of diagonal between triangles.

	numTriangles := 2 * (iMaxRow - iMinRow) * (iMaxCol - iMinCol)
	// fmt.Printf("Num triangles: %d\n", numTriangles)
	ta := &triangleArray{
		triangles:     make([]meshTriangle, numTriangles),
		iteratorIndex: 0,
	}

	n := 0
	for ic := iMinCol; ic < iMaxCol; ic++ {
		for ir := iMinRow; ir < iMaxRow; ir++ {
			row0 := &t.rows[ir]
			row1 := &t.rows[ir+1]

			ta.triangles[n].vertices[0] = geom.NewPt3(t.x[ic], row0.y, row0.z[ic])
			ta.triangles[n].vertices[1] = geom.NewPt3(t.x[ic], row1.y, row1.z[ic])
			ta.triangles[n].vertices[2] = geom.NewPt3(t.x[ic+1], row1.y, row1.z[ic+1])
			ta.triangles[n].normal = row0.normals[2*ic]

			n++

			ta.triangles[n].vertices[0] = geom.NewPt3(t.x[ic], row0.y, row0.z[ic])
			ta.triangles[n].vertices[1] = geom.NewPt3(t.x[ic+1], row1.y, row1.z[ic+1])
			ta.triangles[n].vertices[2] = geom.NewPt3(t.x[ic+1], row0.y, row0.z[ic+1])
			ta.triangles[n].normal = row0.normals[2*ic+1]
		}
	}

	return ta
}

// Find the rows that overlap with the given footprint. Returns indices iMinRow, iMaxRow
// such that the footprint fits entirely with the y-coordinates of each row. Returns
// iMinRow == iMaxRow == -1 if the footprint doesn't overlap the mesh at all.
func (t *TriangleMesh) findRowsForFootprint(f Footprint) (iMinRow, iMaxRow int) {
	// Check for empty mesh.
	if len(t.rows) == 0 {
		return -1, -1
	}

	// Check for footprint trivially above or below mesh's y-boundaries.
	if f.PMax.Y < t.xyBox.PMin.Y || f.PMin.Y > t.xyBox.PMax.Y {
		return -1, -1
	}

	// Index to top row of vertices.
	iTopRow := len(t.rows) - 1

	// Find the first row that is just below or level with the footprint.
	iMinRow = 0
	for {
		if iMinRow == iTopRow-1 {
			break // Reached top-most limit for iMinRow.
		}
		if f.PMin.Y <= t.rows[iMinRow+1].y {
			break
		}
		iMinRow++
	}

	// Now look for the first row that is strictly above the footprint.
	iMaxRow = iMinRow + 1
	for {
		if iMaxRow == iTopRow {
			break // Reached top-most row.
		}
		if t.rows[iMaxRow].y > f.PMax.Y {
			break
		}
		iMaxRow++
	}

	// fmt.Printf("Print min/max rows: %d - %d\n", iMinRow, iMaxRow)

	return
}

// Find the columns that overlap with the given footprint. Returns indices iMinCol, iMaxCol
// such that the footprint fits entirely with the x-coordinates of each column. Returns
// iMinCol == iMaxCol == -1 if the footprint doesn't overlap the mesh at all.
func (t *TriangleMesh) findColumnsForFootprint(f Footprint) (iMinCol, iMaxCol int) {
	// Check for empty mesh.
	if len(t.rows) == 0 {
		return -1, -1
	}

	// Check for footprint trivially to the left or right of mesh's y-boundaries.
	if f.PMax.X < t.xyBox.PMin.X || f.PMin.X > t.xyBox.PMax.X {
		return -1, -1
	}

	// Index to last column of x-coordinates.
	iLastCol := len(t.x) - 1

	// Find the first column that is just to the left or even with the footprint.
	iMinCol = 0
	for {
		if iMinCol == iLastCol-1 {
			break // Reached right-most limit for iMinCol.
		}
		if f.PMin.X <= t.x[iMinCol+1] {
			break
		}
		iMinCol++
	}

	// Now look for the first column that is strictly to the right of the footprint.
	iMaxCol = iMinCol + 1
	for {
		if iMaxCol == iLastCol {
			break // Reached right-most column.
		}
		if t.x[iMaxCol] > f.PMax.X {
			break
		}
		iMaxCol++
	}

	// fmt.Printf("Min.max column: %d, %d\n", iMinCol, iMaxCol)

	return
}

func (t *TriangleMesh) buildMesh(zBlack, zWhite float64, sampler hmap.ScalarGridSampler) {
	pMin := t.xyBox.PMin
	pMax := t.xyBox.PMax
	numGridRows := sampler.GetNumSamplesFromY0ToY1(pMin.Y, pMax.Y)
	numVerticesPerRow := sampler.GetNumSamplesFromX0ToX1(pMin.X, pMax.X)

	if numGridRows <= 1 {
		numGridRows = 2
	}
	if numVerticesPerRow <= 1 {
		numVerticesPerRow = 2
	}

	// Fill x-coordinates array.
	t.x = make([]float64, numVerticesPerRow)
	for i := range t.x {
		x := 0.0
		switch i {
		case 0:
			x = pMin.X
		case numVerticesPerRow - 1:
			x = pMax.X
		default:
			t := float64(i) / float64(numVerticesPerRow-1)
			x = (1.0-t)*pMin.X + t*pMax.X
		}

		t.x[i] = x
	}

	// fmt.Printf("x-coordinates: %v\n", t.x)

	// Fill rows.
	t.rows = make([]gridRow, numGridRows)
	for i := range t.rows {
		yRow := 0.0
		switch i {
		case 0:
			yRow = pMin.Y
		case numGridRows - 1:
			yRow = pMax.Y
		default:
			t := float64(i) / float64(numGridRows-1)
			yRow = (1-t)*pMin.Y + t*pMax.Y
		}

		t.rows[i].y = yRow
		t.populateVerticesForRow(i, zBlack, zWhite, sampler)

		if i > 0 {
			t.populateNormalsForRow(i - 1)
		}
	}
}

// Allocate and fill the array of z-coordinates for grid row with index rowIndex.
func (t *TriangleMesh) populateVerticesForRow(
	rowIndex int,
	zBlack, zWhite float64,
	sampler hmap.ScalarGridSampler) {

	numVertices := len(t.x)
	y := t.rows[rowIndex].y
	t.rows[rowIndex].z = make([]float64, numVertices)

	for i := range t.x {
		x := t.x[i]
		uv := geom.NewPt2(x, y)
		z := sampler.At(uv)
		z = (1.0-z)*zBlack + z*zWhite
		t.rows[rowIndex].z[i] = z
	}
}

// Allocate and fill the array of triangle normals for grid row with index rowIndex.
// Pre-condition: z-coordinates  for rows at index rowIndex and rowIndex+1 must be populated.
func (t *TriangleMesh) populateNormalsForRow(rowIndex int) {
	if rowIndex >= len(t.rows)-1 {
		log.Fatal("Index rowIndex out of range")
		return
	}

	//    xk  xl   x-coordinates xk and xl
	// yj +---+    row of vertices at y = yj
	//    |  /|
	//    | / |   grid cell with 2 triangles
	//    |/  |
	// yi +---+    row of vertices at y = yi
	yi := t.rows[rowIndex].y
	yj := t.rows[rowIndex+1].y
	zi := t.rows[rowIndex].z
	zj := t.rows[rowIndex+1].z

	// There are two normal vectors (two triangles) per grid cell. Each cell
	// consists of four vertices, two from each row.
	numVertices := len(t.x)
	numNormals := (numVertices - 1) * 2
	t.rows[rowIndex].normals = make([]geom.Vec3, numNormals)

	for k := 0; k < numVertices-1; k++ {
		xk := t.x[k]
		xl := t.x[k+1]

		p0 := geom.NewPt3(xk, yi, zi[k])
		p1 := geom.NewPt3(xk, yj, zj[k])
		p2 := geom.NewPt3(xl, yj, zj[k+1])

		// fmt.Printf("Normal: p0=%v, p1=%v, p2=%v", p0, p1, p2)

		w1 := p1.Sub(p0)
		w2 := p2.Sub(p0)
		n1 := w2.Cross(w1)
		t.rows[rowIndex].normals[2*k] = n1.Norm()

		// fmt.Printf(" - w1=%v, w2=%v", w1, w2)

		p2 = geom.NewPt3(xl, yi, zi[k+1])
		w1 = p2.Sub(p0)
		n2 := w1.Cross(w2)
		t.rows[rowIndex].normals[2*k+1] = n2.Norm()

		// fmt.Printf(" - n1=%v, n2=%v\n", n1, n2)
	}
}

// GetTriangleCount implements interface TriangleIterator.
func (t *triangleArray) GetTriangleCount() int {
	return len(t.triangles)
}

// Next implements interface TriangleIterator.
func (t *triangleArray) Next() Triangle {
	if t.iteratorIndex >= t.GetTriangleCount() {
		return nil
	}

	retVal := &t.triangles[t.iteratorIndex]
	t.iteratorIndex++
	return retVal
}
