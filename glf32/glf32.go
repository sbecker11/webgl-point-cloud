// glf32/glf32.go
package glf32

import (
	"math"
	"fmt"
)
 
// Mat4 represents a 4x4 column-major matrix as a slice of 16 float32 values.
// The elements are stored in column-major order:
// m0, m1, m2, m3 (column 0)
// m4, m5, m6, m7 (column 1)
// m8, m9, m10, m11 (column 2)
// m12, m13, m14, m15 (column 3)
type Mat4 []float32

// Vec3 represents a 3D vector as a slice of 3 float32 values.
type Vec3 []float32

// Vec4 represents a 4D vector as a slice of 4 float32 values.
type Vec4 []float32

// PrintMat4 prints a 4x4 matrix in column major order with an optional label.
func PrintMat4(label string, m Mat4) {
	if label != "" {
		fmt.Printf("%s:\n", label)
	}
	//                            col0, col1, col2, col3
	fmt.Printf("[%f %f %f %f]\n", m[0], m[4], m[8], m[12])  // row0
	fmt.Printf("[%f %f %f %f]\n", m[1], m[5], m[9], m[13])  // row1
	fmt.Printf("[%f %f %f %f]\n", m[2], m[6], m[10], m[14]) // ros2
	fmt.Printf("[%f %f %f %f]\n", m[3], m[7], m[11], m[15]) // row3
}

// PrintVec3 prints a 3D vector with optional label.
func PrintVec3(label string, v Vec3) {
	if label != "" {
		fmt.Printf("%s: ", label)
	}
	fmt.Printf("[%f %f %f]\n", v[0], v[1], v[2])
}

// PrintVec4 prints a 4D vector with optional label.
func PrintVec4(label string, v Vec4) {
	if label != "" {
		fmt.Printf("%s: ", label)
	}
	fmt.Printf("[%f %f %f %f]\n", v[0], v[1], v[2], v[3])
}

// PrintVec3ListElement prints the i'th 3D vector element in a packed list of 3D vectors.
func PrintVec3ListElement(i int, m Vec3) {
	fmt.Printf("%d: [%f %f %f]\n", i, m[i*3], m[i*3+1], m[i*3+2])
}

// PrintVec4ListElement prints the i'th 4D vector element in a packed list of 4D vectors.
func PrintVec4ListElement(i int, m Vec4) {
	fmt.Printf("%d: [%f %f %f %f]\n", i, m[i*4], m[i*4+1], m[i*4+2], m[i*4+3])
}

// Identity returns a new 4x4 identity matrix (column-major).
func Identity() Mat4 {
	return Mat4{
		1, 0, 0, 0, // Column 0
		0, 1, 0, 0, // Column 1
		0, 0, 1, 0, // Column 2
		0, 0, 0, 1, // Column 3
	}
}

// Subtract performs component-wise subtraction of two 3D vectors.
// It returns a new Vec3.
// Panics if input vectors are not of length 3.
func Subtract(a, b Vec3) Vec3 {
	if len(a) != 3 || len(b) != 3 {
		panic("Subtract: input vectors must be Vec3 (length 3)")
	}
	return Vec3{a[0] - b[0], a[1] - b[1], a[2] - b[2]}
}

// Normalize returns a new 3D vector with the same direction but a magnitude of 1.
// If the input vector is a zero vector, it returns the zero vector [0,0,0].
// Panics if input vector is not of length 3.
func Normalize(v Vec3) Vec3 {
	if len(v) != 3 {
		panic("Normalize: input vector must be Vec3 (length 3)")
	}
	l := float32(math.Sqrt(float64(v[0]*v[0] + v[1]*v[1] + v[2]*v[2])))
	if l > 0 {
		return Vec3{v[0] / l, v[1] / l, v[2] / l}
	}
	return Vec3{0, 0, 0}
}

// Cross calculates the cross product of two 3D vectors.
// It returns a new Vec3.
// Panics if input vectors are not of length 3.
func Cross(a, b Vec3) Vec3 {
	if len(a) != 3 || len(b) != 3 {
		panic("Cross: input vectors must be Vec3 (length 3)")
	}
	return Vec3{
		a[1]*b[2] - a[2]*b[1],
		a[2]*b[0] - a[0]*b[2],
		a[0]*b[1] - a[1]*b[0],
	}
}

// Dot calculates the dot product of two 3D vectors.
// Panics if input vectors are not of length 3.
func Dot(a, b Vec3) float32 {
	if len(a) != 3 || len(b) != 3 {
		panic("Dot: input vectors must be Vec3 (length 3)")
	}
	return a[0]*b[0] + a[1]*b[1] + a[2]*b[2]
}

// Translate creates a 4x4 column-major translation matrix.
//
// Parameters:
//   x, y, z: The translation amounts along the respective axes.
//
// Returns a column major Mat4 representing the translation matrix.
func Translate(x, y, z float32) Mat4 {
	return Mat4{
		1, 0, 0, 0, // Column 0
		0, 1, 0, 0, // Column 1
		0, 0, 1, 0, // Column 2
		x, y, z, 1, // Column 3
	}
}

// RotateX creates a 4x4 column-major matrix for rotation around the X-axis.
//
// Parameters:
//   angle: The rotation angle in radians.
//
// Returns a column-majorMat4 representing the rotation matrix.
func RotateX(angle float32) Mat4 {
	s, c := float32(math.Sin(float64(angle))), float32(math.Cos(float64(angle)))
	return Mat4{
		1, 0, 0, 0, // Column 0
		0, c, s, 0, // Column 1
		0, -s, c, 0, // Column 2
		0, 0, 0, 1, // Column 3
	}
}

// RotateY creates a 4x4 column-major matrix for rotation around the Y-axis.
//
// Parameters:
//   angle: The rotation angle in radians.
//
// Returns a column-major Mat4 representing the rotation matrix.
func RotateY(angle float32) Mat4 {
	s, c := float32(math.Sin(float64(angle))), float32(math.Cos(float64(angle)))
	return Mat4{
		c, 0, -s, 0, // Column 0
		0, 1, 0, 0, // Column 1
		s, 0, c, 0, // Column 2
		0, 0, 0, 1, // Column 3
	}
}

// RotateZ creates a 4x4 column-major matrix for rotation around the Z-axis.
//
// Parameters:
//   angle: The rotation angle in radians.
//
// Returns a column-major Mat4 representing the rotation matrix.
func RotateZ(angle float32) Mat4 {
	s, c := float32(math.Sin(float64(angle))), float32(math.Cos(float64(angle)))
	return Mat4{
		c, s, 0, 0, // Column 0
		-s, c, 0, 0, // Column 1
		0, 0, 1, 0, // Column 2
		0, 0, 0, 1, // Column 3
	}
}

// LookAt creates a 4x4 column-major view matrix that transforms world
// coordinates into camera (view) coordinates. This is used to position and
// orient the camera in the scene.
//
// Parameters:
//   eye: The position of the camera in world space (e.g., Vec3{x, y, z}).
//   center: The point in world space that the camera is looking at.
//   up: The world's "up" direction (typically Vec3{0, 1, 0}).
//
// Returns a Mat4 representing the 4x4 column-major view matrix.
// Panics if input vectors are not of length 3.
func LookAt(eye, center, up Vec3) Mat4 {
	if len(eye) != 3 || len(center) != 3 || len(up) != 3 {
		panic("LookAt: input vectors must be Vec3 (length 3)")
	}

	f := Normalize(Subtract(center, eye))
	s := Normalize(Cross(f, up))
	u := Cross(s, f)

	tx := -Dot(s, eye)
	ty := -Dot(u, eye)
	tz := Dot(f, eye) // This is equivalent to -Dot(-f, eye)

	// The view matrix is the inverse of the camera's transformation matrix.
	// For column-major order, this is the correctly transposed layout.
	return Mat4{
		// Column 0
		s[0], u[0], -f[0], 0,
		// Column 1
		s[1], u[1], -f[1], 0,
		// Column 2
		s[2], u[2], -f[2], 0,
		// Column 3
		tx, ty, tz, 1,
	}
}

// Perspective creates a 4x4 column-major perspective projection matrix.
// This matrix transforms 3D camera-space coordinates into 2D clip-space coordinates,
// accounting for perspective (objects further away appear smaller).
// It maps the Z coordinate to the range [-1, 1] in clip space (common for OpenGL/WebGL).
//
// Parameters:
//   fov: The vertical field of view in radians.
//   aspect: The aspect ratio of the viewport (width / height).
//   near: The distance to the near clipping plane. Must be positive.
//   far: The distance to the far clipping plane. Must be positive and greater than near.
//
// Returns a Mat4 representing the 4x4 column-major perspective matrix.
func Perspective(fov, aspect, near, far float32) Mat4 {
	// f is the reciprocal of the tangent of half the field of view.
	// This scales the X and Y coordinates to fit the frustum.
	f := 1.0 / float32(math.Tan(float64(fov)/2))

	// nf is a pre-calculated factor used for the Z transformation, to avoid division.
	// It's 1 / (near - far).
	nf := 1.0 / (near - far)

	return Mat4{
		// Column 0
		f / aspect, 0, 0, 0,
		// Column 1
		0, f, 0, 0,
		// Column 2
		0, 0, (far + near) * nf, -1, // Z transformation and projection term (for W)
		// Column 3
		0, 0, (2 * far * near) * nf, 0, // Z translation for W component
	}
}

// MultiplyMatrices performs the multiplication of two 4x4 column-major matrices (A * B).
// The result is also a 4x4 column-major matrix.
//
// Parameters:
//   a: The first 4x4 column-major matrix (left operand).
//   b: The second 4x4 column-major matrix (right operand).
//
// Returns a new Mat4 representing the product matrix (A * B).
// Panics if input matrices are not of length 16.
func MultiplyMatrices(a, b Mat4) Mat4 {
	if len(a) != 16 || len(b) != 16 {
		panic("MultiplyMatrices: input matrices must be Mat4 (length 16)")
	}

	c := make(Mat4, 16) // Result matrix

	// C[row][col] = sum( A[row][k_inner] * B[k_inner][col] )
	// For column-major indices:
	// A[row][k_inner] is a[k_inner*4 + row]
	// B[k_inner][col] is b[col*4 + k_inner]
	// C[row][col] is c[col*4 + row]

	for i := 0; i < 4; i++ { // 'i' iterates over rows of the result (0 to 3)
		for j := 0; j < 4; j++ { // 'j' iterates over columns of the result (0 to 3)
			c[j*4+i] = 0 // Initialize the element at c[i][j]
			for k := 0; k < 4; k++ { // 'k' iterates over the inner sum (0 to 3)
				c[j*4+i] += a[k*4+i] * b[j*4+k]
			}
		}
	}
	return c
}

// TransformVertices applies a 4x4 column-major matrix to a slice of 3D vertex coordinates.
// Each vertex (x, y, z) is treated as a 4D homogeneous vector (x, y, z, 1) for transformation.
// The transformed x, y, z components are then stored back into the original slice,
// after performing the perspective divide (dividing by w).
//
// Parameters:
//   coords: A []float32 slice containing 3D vertex coordinates (e.g., [x0, y0, z0, x1, y1, z1, ...]).
//   m: The 4x4 column-major transformation matrix to apply.
//
// Returns:
//   The modified 'coords' slice with transformed vertices.
//   Panics if the matrix 'm' is not of length 16 or if 'coords' length is not a multiple of 3.
func TransformVertices(coords []float32, m Mat4) []float32 {
	if len(m) != 16 {
		panic("TransformVertices: transformation matrix must be Mat4 (length 16)")
	}
	if len(coords)%3 != 0 {
		panic("TransformVertices: coords slice length must be a multiple of 3")
	}

	numVertices := len(coords) / 3

	for i := 0; i < numVertices; i++ {
		idx := i * 3 // Starting index for the current vertex

		x := coords[idx]
		y := coords[idx+1]
		z := coords[idx+2]
		w := float32(1.0) // Homogeneous coordinate

		// Perform M * V for column-major matrix multiplication:
		// Result components are the dot product of matrix rows with the vector components.
		// newX = M[0][0]*x + M[0][1]*y + M[0][2]*z + M[0][3]*w
		// In column-major memory layout (m[col*4 + row]):
		// newX = m[0]*x + m[4]*y + m[8]*z + m[12]*w
		transformedX := m[0]*x + m[4]*y + m[8]*z + m[12]*w
		// newY = M[1][0]*x + M[1][1]*y + M[1][2]*z + M[1][3]*w
		transformedY := m[1]*x + m[5]*y + m[9]*z + m[13]*w
		// newZ = M[2][0]*x + M[2][1]*y + M[2][2]*z + M[2][3]*w
		transformedZ := m[2]*x + m[6]*y + m[10]*z + m[14]*w
		// newW = M[3][0]*x + M[3][1]*y + M[3][2]*z + M[3][3]*w
		transformedW := m[3]*x + m[7]*y + m[11]*z + m[15]*w

		// Perspective Divide: Divide by w if it's not 0, to convert back to 3D Cartesian coordinates.
		// This is crucial after projection, where W stores depth information.
		if transformedW != 0 {
			coords[idx] = transformedX / transformedW
			coords[idx+1] = transformedY / transformedW
			coords[idx+2] = transformedZ / transformedW
		} else {
			// Handle case where W is 0 (e.g., point at infinity or invalid transformation).
			// For practical purposes in graphics, this often means the point is clipped or invalid.
			// Set to 0 or some other indicator as appropriate for your rendering pipeline.
			coords[idx] = 0
			coords[idx+1] = 0
			coords[idx+2] = 0
		}
	}
	return coords
}
