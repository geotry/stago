package compute

import (
	"log"
	"math"
	"slices"
)

type Simplex struct {
	Points    []SupportPoint
	Direction Vector3
	Size      int
}

type Polytope struct {
	Points []SupportPoint
	Faces  []Vector3i
}

type SupportPoint struct {
	Point          Vector3
	IndexA, IndexB int
}

type CollisionInfo struct {
	Contact Vector3
	Normal  Vector3
	Depth   float64
}

const MaxGJKIterations = 16
const MaxEPAIterations = 32
const Epsilon = 0.0001

// Run GJK algorithm on two bounding volumes and return intersection
// https://winter.dev/articles/gjk-algorithm
// Read: https://medium.com/@mbayburt/walkthrough-of-the-gjk-collision-detection-algorithm-80823ef5c774
// Also: https://www.medien.ifi.lmu.de/lehre/ss10/ps/Ausarbeitung_Beispiel.pdf
func GJK(a, b []Vector3) (bool, CollisionInfo) {
	simplex := NewSimplex(4)

	// First support point
	support := calculateSupport(a, b, simplex.Direction)
	simplex.PushFront(support)
	simplex.Direction = support.Point.Opposite().Normalize()

	// Loop to find next support points
	iterations := 0
	for {
		support = calculateSupport(a, b, simplex.Direction)
		if support.Point.Dot(simplex.Direction) <= 0 {
			return false, CollisionInfo{}
		}
		simplex.PushFront(support)

		if simplex.Next() {
			// Collision found
			// Use simplex to find penetration depth and normal
			return true, EPA(simplex, a, b)
		}

		iterations++
		if iterations >= MaxGJKIterations {
			log.Printf("Did not find simplex after %d iterations, simplex: %v", MaxGJKIterations, simplex)
			return false, CollisionInfo{}
		}
	}
}

// https://winter.dev/articles/epa-algorithm
func EPA(simplex *Simplex, a, b []Vector3) CollisionInfo {
	polytope := NewPolytope(simplex.Points)
	polytope.Faces = []Vector3i{
		{X: 0, Y: 1, Z: 2},
		{X: 0, Y: 3, Z: 1},
		{X: 0, Y: 2, Z: 3},
		{X: 1, Y: 3, Z: 2},
	}

	normals, minFace := polytope.getFaceNormals(polytope.Faces)
	minNormal := normals[minFace].ToVector3()
	minDistance := math.Inf(1)

	iterations := 0

	for minDistance == math.Inf(1) && iterations < MaxEPAIterations {
		minNormal = normals[minFace].ToVector3()
		minDistance = normals[minFace].W

		support := calculateSupport(a, b, minNormal)
		sDistance := minNormal.Dot(support.Point)

		if math.Abs(sDistance-minDistance) > Epsilon {
			minDistance = math.Inf(1)
			uniqueEdges := []Vector2i{}

			for i := 0; i < len(normals); i++ {
				normal := normals[i]
				if (normal.ToVector3().Dot(support.Point) - normal.W) > 0 {
					face := polytope.Faces[i]
					uniqueEdges = addIfUniqueEdge(uniqueEdges, Vector2i{X: face.X, Y: face.Y})
					uniqueEdges = addIfUniqueEdge(uniqueEdges, Vector2i{X: face.Y, Y: face.Z})
					uniqueEdges = addIfUniqueEdge(uniqueEdges, Vector2i{X: face.Z, Y: face.X})

					polytope.Faces[i] = polytope.Faces[len(polytope.Faces)-1]
					polytope.Faces = polytope.Faces[:len(polytope.Faces)-1]

					normals[i] = normals[len(normals)-1]
					normals = normals[:len(normals)-1]

					i = i - 1
				}
			}

			newFaces := []Vector3i{}
			for _, edge := range uniqueEdges {
				newFaces = append(newFaces, Vector3i{X: edge.X, Y: edge.Y, Z: len(polytope.Points)})
			}

			polytope.Points = append(polytope.Points, support)

			newNormals, newMinFace := polytope.getFaceNormals(newFaces)

			oldMinDistance := math.Inf(1)

			for i, normal := range normals {
				if normal.W < oldMinDistance {
					oldMinDistance = normal.W
					minFace = i
				}
			}

			if newNormals[newMinFace].W < oldMinDistance {
				minFace = newMinFace + len(normals)
			}

			polytope.Faces = append(polytope.Faces, newFaces...)
			normals = append(normals, newNormals...)
		}
		iterations = iterations + 1
	}

	// https://github.com/exatb/GJKEPA/blob/main/gjkepa.cs#L502
	// Todo: Find contact point

	sa := polytope.Points[polytope.Faces[minFace].X]
	sb := polytope.Points[polytope.Faces[minFace].Y]
	sc := polytope.Points[polytope.Faces[minFace].Z]

	// Find the projected point on surface of triangle
	distance := sa.Point.Dot(minNormal)
	projectedPoint := minNormal.Mult(-distance)

	// Getting the barycentric coordinates of this projection within the triangle belonging to the simplex
	u, v, w := GetBarycentricCoordinates(projectedPoint, sa.Point, sb.Point, sc.Point)

	// Taking the corresponding triangle from the first polyhedron
	a1 := a[sa.IndexA]
	b1 := a[sb.IndexA]
	c1 := a[sc.IndexA]

	// Contact point on the first polyhedron
	contactPoint1 := a1.Mult(u).Add(b1.Mult(v)).Add(c1.Mult(w))

	// Taking the corresponding triangle from the second polyhedron
	a2 := b[sa.IndexB]
	b2 := b[sb.IndexB]
	c2 := b[sc.IndexB]

	// Contact point on the second polyhedron
	contactPoint2 := a2.Mult(u).Add(b2.Mult(v)).Add(c2.Mult(w))

	contactPoint := contactPoint1.Add(contactPoint2).Div(2) // Returning the midpoint

	return CollisionInfo{
		Contact: contactPoint,
		Normal:  minNormal,
		Depth:   minDistance + 0.001,
	}
}

func GetBarycentricCoordinates(p Vector3, a, b, c Vector3) (float64, float64, float64) {
	// Vectors from vertex A to vertices B and C
	v0 := b.Sub(a)
	v1 := c.Sub(a)
	v2 := p.Sub(a)

	// Compute dot products
	d00 := v0.Dot(v0) // Same as length squared V0
	d01 := v0.Dot(v1)
	d11 := v1.Dot(v1) // Same as length squared V1
	d20 := v2.Dot(v0)
	d21 := v2.Dot(v1)
	denom := d00*d11 - d01*d01

	u := 0.0
	v := 0.0
	w := 0.0

	// Check for a zero denominator before division = check for degenerate triangle (area is zero)
	if math.Abs(denom) <= Epsilon {
		// The triangle is degenerate

		// Check if all vertices coincide (triangle collapses to a point)
		if d00 <= Epsilon && d11 <= Epsilon {
			// All edges are degenerate (vertices coincide at a point)
			// Return barycentric coordinates corresponding to vertex a
			u = 1
			v = 0
			w = 0
		} else {
			// Seems triangle collapses to a line (vertices are colinear)
			// We can check it:
			// Vector3 cross = Vector3.Cross(v0, v1);
			// if (Vector3.Dot(cross, cross) <= Epsilon)....
			// But if the triangle area is close to zero and the triangle has not colapsed to a point then it has colapsed to a line
			// Use edge AB if it's not degenerate
			if d00 > Epsilon {
				// Compute parameter t for projection of point p onto line AB
				t := v2.Dot(v0) / d00
				// if |t|>1 then p lies in AC but we can use u,v,w calculated to AB with a small error
				// Barycentric coordinates for edge AB
				u = 1.0 - t // weight for vertex a
				v = t       // weight for vertex b
				w = 0.0     // vertex c does not contribute
			} else if d11 > Epsilon {
				// Else, use edge AC
				// Compute parameter t for projection of point p onto line AC
				t := v2.Dot(v1) / d11
				// Barycentric coordinates for edge AC
				u = 1.0 - t // weight for vertex a
				v = 0.0     // vertex b does not contribute
				w = t       // weight for vertex c
			} else {
				// The triangle is degenerate in an unexpected way
				// Return barycentric coordinates corresponding to vertex a
				u = 1
				v = 0
				w = 0
			}
		}
	} else {
		// Compute barycentric coordinates
		v = (d11*d20 - d01*d21) / denom
		w = (d00*d21 - d01*d20) / denom
		u = 1.0 - v - w
	}
	return u, v, w
}

func NewPolytope(points []SupportPoint) *Polytope {
	polytope := &Polytope{
		Points: make([]SupportPoint, len(points)),
		Faces: []Vector3i{
			{X: 0, Y: 1, Z: 2},
			{X: 0, Y: 3, Z: 1},
			{X: 0, Y: 3, Z: 2},
			{X: 1, Y: 2, Z: 3},
		},
	}
	copy(polytope.Points, points)

	return polytope
}

func (p *Polytope) getFaceNormals(faces []Vector3i) ([]Vector4, int) {
	minTriangle := 0
	minDistance := math.Inf(1)
	normals := make([]Vector4, 0)

	for i, face := range faces {
		a := p.Points[face.X].Point
		b := p.Points[face.Y].Point
		c := p.Points[face.Z].Point

		normal := b.Sub(a).Cross(c.Sub(a))
		l := normal.Length()
		var distance float64

		if l < Epsilon {
			normal = Vector3{}
			distance = math.Inf(1)
		} else {
			normal = normal.Normalize()
			distance = normal.Dot(a)
		}

		if distance < 0 {
			normal = normal.Mult(-1)
			distance *= -1
		}

		normals = append(normals, Vector4{X: normal.X, Y: normal.Y, Z: normal.Z, W: distance})

		if distance < minDistance {
			minTriangle = i
			minDistance = distance
		}
	}

	return normals, minTriangle
}

func addIfUniqueEdge(edges []Vector2i, faceEdge Vector2i) []Vector2i {
	reverseIndex := -1
	for i, edge := range edges {
		if edge.X == faceEdge.Y && edge.Y == faceEdge.X {
			reverseIndex = i
			break
		}
	}

	if reverseIndex != -1 {
		edges = slices.Delete(edges, reverseIndex, reverseIndex+1)
	} else {
		edges = append(edges, faceEdge)
	}

	return edges
}

// To be define in primitive
// Find the most extreme point in points in direction
func getSupportPoint(vertices []Vector3, direction Vector3) (Vector3, int) {
	var maxPoint Vector3
	maxDistance := math.Inf(-1)
	maxIndex := -1

	for i, vertex := range vertices {
		distance := vertex.Dot(direction)
		if distance > maxDistance {
			maxDistance = distance
			maxPoint = vertex
			maxIndex = i
		}
	}

	return maxPoint, maxIndex
}

func calculateSupport(a, b []Vector3, d Vector3) SupportPoint {
	sa, ia := getSupportPoint(a, d)
	sb, ib := getSupportPoint(b, d.Opposite())
	return SupportPoint{Point: sa.Sub(sb), IndexA: ia, IndexB: ib}
}

func sameDirection(direction Vector3, ao Vector3) bool {
	return direction.Dot(ao) > 0
}

func NewSimplex(size int) *Simplex {
	return &Simplex{
		Points:    make([]SupportPoint, size),
		Size:      0,
		Direction: Vector3{Z: 1},
	}
}

func (s *Simplex) PushFront(p SupportPoint) {
	// s.Size must be < len(p.Points)
	// Shift points starting from last one
	for i := range s.Size {
		s.Points[s.Size-i] = s.Points[s.Size-i-1]
	}
	s.Points[0] = p
	s.Size++
}

func (s *Simplex) Set(p SupportPoint, index int) {
	s.Points[index] = p
	s.Size = index + 1
}

func (s *Simplex) Next() bool {
	switch s.Size {
	case 2:
		return s.line()
	case 3:
		return s.triangle()
	case 4:
		return s.tetrahedron()
	}
	return false
}

func (s *Simplex) line() bool {
	a := s.Points[0].Point
	b := s.Points[1].Point

	ab := b.Sub(a)
	ao := a.Opposite()

	if sameDirection(ab, ao) {
		s.Direction = ab.Cross(ao).Cross(ab).Normalize()
	} else {
		s.Set(s.Points[0], 0)
		s.Direction = ao.Normalize()
	}

	return false
}

func (s *Simplex) triangle() bool {
	a := s.Points[0].Point
	b := s.Points[1].Point
	c := s.Points[2].Point

	ab := b.Sub(a)
	ac := c.Sub(a)
	ao := a.Opposite()

	abc := ab.Cross(ac)

	if sameDirection(abc.Cross(ac), ao) {
		if sameDirection(ac, ao) {
			s.Set(s.Points[0], 0)
			s.Set(s.Points[2], 1)
			s.Direction = ac.Cross(ao).Cross(ac).Normalize()
		} else {
			s.Set(s.Points[0], 0)
			s.Set(s.Points[1], 1)
			return s.line()
		}
	} else {
		if sameDirection(ab.Cross(abc), ao) {
			s.Set(s.Points[0], 0)
			s.Set(s.Points[1], 1)
			return s.line()
		} else {
			if sameDirection(abc, ao) {
				s.Direction = abc.Normalize()
			} else {
				s.Set(s.Points[0], 0)
				s.Set(s.Points[2], 1)
				s.Set(s.Points[1], 2)
				s.Direction = abc.Opposite().Normalize()
			}
		}
	}

	return false
}

func (s *Simplex) tetrahedron() bool {
	a := s.Points[0].Point
	b := s.Points[1].Point
	c := s.Points[2].Point
	d := s.Points[3].Point

	ab := b.Sub(a)
	ac := c.Sub(a)
	ad := d.Sub(a)
	ao := a.Opposite()

	abc := ab.Cross(ac)
	if sameDirection(abc, ao) {
		s.Set(s.Points[0], 0)
		s.Set(s.Points[1], 1)
		s.Set(s.Points[2], 2)
		return s.triangle()
	}

	acd := ac.Cross(ad)
	if sameDirection(acd, ao) {
		s.Set(s.Points[0], 0)
		s.Set(s.Points[2], 1)
		s.Set(s.Points[3], 2)
		return s.triangle()
	}

	adb := ad.Cross(ab)
	if sameDirection(adb, ao) {
		s.Set(s.Points[0], 0)
		s.Set(s.Points[3], 1)
		s.Set(s.Points[1], 2)
		return s.triangle()
	}

	return true
}
