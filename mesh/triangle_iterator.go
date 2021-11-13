package mesh

// TriangleIterator is used to retrieve and iterate over a set of triangles in a TriangleMesh.
type TriangleIterator interface {
	GetTriangleCount() int
	Next() Triangle
}
