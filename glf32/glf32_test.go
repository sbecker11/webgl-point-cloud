// glf32/glf32_test.go
// usage: go test

package glf32

import (
	"math"
	"testing"
)

const float32EqualityThreshold = 1e-6

//
// Helper Functions
//

//
// Compare two float32 values to see if they are almost equal.
//
func almostEqual(a, b float32) bool {
	return math.Abs(float64(a-b)) <= float32EqualityThreshold
}

//
// Compare two Vec3 vectors to see if they are almost equal.
//
func vec3AlmostEqual(a, b Vec3) bool {
	if len(a) != 3 || len(b) != 3 {
		return false
	}
	return almostEqual(a[0], b[0]) && almostEqual(a[1], b[1]) && almostEqual(a[2], b[2])
}

//
// Compare two Mat4 matrices to see if they are almost equal.
//
func mat4AlmostEqual(a, b Mat4) bool {
	if len(a) != 16 || len(b) != 16 {
		return false
	}
	for i := range a {
		if !almostEqual(a[i], b[i]) {
			return false
		}
	}
	return true
}

//
// --- Vector Operation Tests ---
//

func TestSubtract(t *testing.T) {
	a := Vec3{1, 2, 3}
	b := Vec3{4, 5, 6}
	expected := Vec3{-3, -3, -3}
	result := Subtract(a, b)
	if !vec3AlmostEqual(result, expected) {
		t.Errorf("Subtract failed: expected %v, got %v", expected, result)
	}
}

func TestNormalize(t *testing.T) {
	v := Vec3{3, 4, 0}
	expected := Vec3{0.6, 0.8, 0}
	result := Normalize(v)
	if !vec3AlmostEqual(result, expected) {
		t.Errorf("Normalize failed: expected %v, got %v", expected, result)
	}

	// Test normalizing a zero vector
	zeroVec := Vec3{0, 0, 0}
	expectedZero := Vec3{0, 0, 0}
	resultZero := Normalize(zeroVec)
	if !vec3AlmostEqual(resultZero, expectedZero) {
		t.Errorf("Normalize zero vector failed: expected %v, got %v", expectedZero, resultZero)
	}
}

func TestCross(t *testing.T) {
	a := Vec3{1, 0, 0}
	b := Vec3{0, 1, 0}
	expected := Vec3{0, 0, 1}
	result := Cross(a, b)
	if !vec3AlmostEqual(result, expected) {
		t.Errorf("Cross product failed: expected %v, got %v", expected, result)
	}
}

func TestDot(t *testing.T) {
	a := Vec3{1, 2, 3}
	b := Vec3{4, -5, 6}
	expected := float32(12)
	result := Dot(a, b)
	if !almostEqual(result, expected) {
		t.Errorf("Dot product failed: expected %f, got %f", expected, result)
	}
}

//
// --- Matrix Generation Tests ---
//

func TestIdentity(t *testing.T) {
	expected := Mat4{
		1, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, 1, 0,
		0, 0, 0, 1,
	}
	result := Identity()
	if !mat4AlmostEqual(result, expected) {
		t.Errorf("Identity matrix failed: expected %v, got %v", expected, result)
	}
}

func TestTranslate(t *testing.T) {
	expected := Mat4{
		1, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, 1, 0,
		10, 20, 30, 1,
	}
	result := Translate(10, 20, 30)
	if !mat4AlmostEqual(result, expected) {
		t.Errorf("Translate matrix failed: expected %v, got %v", expected, result)
	}
}

func TestRotateX(t *testing.T) {
	angle := float32(math.Pi / 2)
	s, c := float32(math.Sin(float64(angle))), float32(math.Cos(float64(angle)))
	expected := Mat4{
		1, 0, 0, 0,
		0, c, s, 0,
		0, -s, c, 0,
		0, 0, 0, 1,
	}
	result := RotateX(angle)
	if !mat4AlmostEqual(result, expected) {
		t.Errorf("RotateX matrix failed: expected %v, got %v", expected, result)
	}
}

func TestRotateY(t *testing.T) {
	angle := float32(math.Pi / 2)
	s, c := float32(math.Sin(float64(angle))), float32(math.Cos(float64(angle)))
	expected := Mat4{
		c, 0, -s, 0,
		0, 1, 0, 0,
		s, 0, c, 0,
		0, 0, 0, 1,
	}
	result := RotateY(angle)
	if !mat4AlmostEqual(result, expected) {
		t.Errorf("RotateY matrix failed: expected %v, got %v", expected, result)
	}
}

func TestRotateZ(t *testing.T) {
	angle := float32(math.Pi / 2)
	s, c := float32(math.Sin(float64(angle))), float32(math.Cos(float64(angle)))
	expected := Mat4{
		c, s, 0, 0,
		-s, c, 0, 0,
		0, 0, 1, 0,
		0, 0, 0, 1,
	}
	result := RotateZ(angle)
	if !mat4AlmostEqual(result, expected) {
		t.Errorf("RotateZ matrix failed: expected %v, got %v", expected, result)
	}
}

//
// --- Matrix Operation Tests ---
//

func TestMultiplyMatrices(t *testing.T) {
	// Test multiplication with identity
	m := RotateX(0.5)
	ident := Identity()
	result := MultiplyMatrices(m, ident)
	if !mat4AlmostEqual(result, m) {
		t.Errorf("Matrix * Identity should be Matrix. Got %v", result)
	}

	// Test two rotation matrices
	rotX := RotateX(float32(math.Pi / 2))
	rotY := RotateY(float32(math.Pi / 2))
	// Expected result for rotY * rotX
	expected := Mat4{
		0, 0, -1, 0,
		1, 0, 0, 0,
		0, -1, 0, 0,
		0, 0, 0, 1,
	}
	// Our MultiplyMatrices(A, B) is A*B where A is on the left.
	// The order matters.
	resultXY := MultiplyMatrices(rotY, rotX)
	if !mat4AlmostEqual(resultXY, expected) {
		t.Errorf("RotateY * RotateX multiplication failed.\nExpected: %v\nGot:      %v", expected, resultXY)
	}
}

//
// --- Camera and Projection Tests ---
//

// transformPoint applies a Mat4 to a Vec3, treating it as a point (w=1).
func transformPoint(p Vec3, m Mat4) Vec3 {
	x, y, z := p[0], p[1], p[2]
	w := float32(1.0)
	
	resX := m[0]*x + m[4]*y + m[8]*z + m[12]*w
	resY := m[1]*x + m[5]*y + m[9]*z + m[13]*w
	resZ := m[2]*x + m[6]*y + m[10]*z + m[14]*w
	resW := m[3]*x + m[7]*y + m[11]*z + m[15]*w

	if resW != 0 {
		return Vec3{resX / resW, resY / resW, resZ / resW}
	}
	return Vec3{resX, resY, resZ}
}

func TestLookAt(t *testing.T) {
	eye := Vec3{0, 0, 5}
	center := Vec3{0, 0, 0}
	up := Vec3{0, 1, 0}
	
	// Expected result for a right-handed system
	// looking from +Z towards the origin.
	expected := Mat4{
		1,  0,  0,  0,
		0,  1,  0,  0,
		0,  0,  1,  0,
		0,  0, -5,  1,
	}

	viewMatrix := LookAt(eye, center, up)
	if !mat4AlmostEqual(viewMatrix, expected) {
		t.Errorf("LookAt failed.\nExpected: %v\nGot:      %v", expected, viewMatrix)
	}

	// Test that a point at the world center is transformed to the correct view space position.
	// Since the camera is at (0,0,5) looking at the origin, the origin should be transformed
	// to (0,0,-5) in view space (5 units in front of the camera along the -Z axis).
	worldPoint := Vec3{0, 0, 0}
	expectedViewPoint := Vec3{0, 0, -5}
	actualViewPoint := transformPoint(worldPoint, viewMatrix)
	if !vec3AlmostEqual(actualViewPoint, expectedViewPoint) {
		t.Errorf("LookAt transform failed. Expected %v, got %v", expectedViewPoint, actualViewPoint)
	}
}


func TestPerspective(t *testing.T) {
	fov := float32(math.Pi / 2) // 90 degrees
	aspect := float32(1.0)
	near := float32(1.0)
	far := float32(100.0)

	f := 1.0 / float32(math.Tan(float64(fov)/2))
	nf := 1.0 / (near - far)

	expected := Mat4{
		f / aspect, 0, 0, 0,
		0, f, 0, 0,
		0, 0, (far + near) * nf, -1,
		0, 0, (2 * far * near) * nf, 0,
	}

	result := Perspective(fov, aspect, near, far)

	if !mat4AlmostEqual(result, expected) {
		t.Errorf("Perspective matrix failed.\nExpected: %v\nGot:      %v", expected, result)
	}
}

//
// --- Vertex Transformation Test ---
//

func TestTransformVertices(t *testing.T) {
	vertices := []float32{
		0, 0, 0, // Point at origin
		1, 0, 0, // Point on X axis
	}
	
	// Translation matrix
	m := Translate(10, 20, 30)

	expected := []float32{
		10, 20, 30,
		11, 20, 30,
	}

	// The function modifies the slice in-place
	transformed := TransformVertices(vertices, m)

	if len(transformed) != len(expected) {
		t.Fatalf("TransformVertices returned slice with wrong length. Expected %d, got %d", len(expected), len(transformed))
	}

	for i := range expected {
		if !almostEqual(transformed[i], expected[i]) {
			t.Errorf("TransformVertices failed at index %d. Expected %f, got %f", i, expected[i], transformed[i])
		}
	}
}

//
// --- Documentation Example Test ---
//

// TestMVPExampleFromREADME ensures the MVP example in the README.md is correct.
// It calculates the full MVP matrix and performs a smoke test to ensure
// it is a valid, non-identity, non-zero transformation.
func TestMVPExampleFromREADME(t *testing.T) {
	// 1. Model Matrix
	angle := float32(math.Pi / 4)
	modelMatrix := RotateY(angle)

	// 2. View Matrix
	eye := Vec3{2, 2, 2}
	center := Vec3{0, 0, 0}
	up := Vec3{0, 1, 0}
	viewMatrix := LookAt(eye, center, up)

	// 3. Projection Matrix
	fov := float32(math.Pi / 4)
	aspect := float32(16.0 / 9.0)
	near, far := float32(0.1), float32(100.0)
	projectionMatrix := Perspective(fov, aspect, near, far)

	// 4. Combine Matrices
	viewModelMatrix := MultiplyMatrices(viewMatrix, modelMatrix)
	mvpMatrix := MultiplyMatrices(projectionMatrix, viewModelMatrix)

	// 5. Smoke Test
	// Check that the resulting matrix is not the identity matrix.
	if mat4AlmostEqual(mvpMatrix, Identity()) {
		t.Error("MVP matrix should not be the identity matrix")
	}
	// Check that the resulting matrix is not a zero matrix.
	if mat4AlmostEqual(mvpMatrix, make(Mat4, 16)) {
		t.Error("MVP matrix should not be a zero matrix")
	}
} 