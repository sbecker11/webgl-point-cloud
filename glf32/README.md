# glf32 Package

The `glf32` package provides fundamental types and functions for 3D mathematics, specifically tailored for use with WebGL in a Go (Golang) WebAssembly environment. It uses `float32` for all its calculations to align with WebGL's native data types.

## Features

### Core Data Types
- **`Vec3`**: Represents a 3D vector `[x, y, z]`.
- **`Vec4`**: Represents a 4D vector `[x, y, z, w]`.
- **`Mat4`**: Represents a 4x4 matrix, stored in column-major order as a flat `[]float32` slice of 16 elements. This layout is required for direct use with WebGL shader uniforms.

### Vector Operations
A suite of standard 3D vector math functions:
- `Subtract(a, b Vec3)`
- `Normalize(v Vec3)`
- `Cross(a, b Vec3)`
- `Dot(a, b Vec3)`

### Matrix Transformations
Functions to create common transformation matrices:
- `Identity()`
- `Translate(x, y, z)`
- `RotateX(angle)`, `RotateY(angle)`, `RotateZ(angle)`
- `MultiplyMatrices(a, b)`

### Camera and Projection
Essential matrices for setting up a 3D scene:
- **`LookAt(eye, center, up)`**: Creates a view matrix to position and orient the camera.
- **`Perspective(fov, aspect, near, far)`**: Creates a perspective projection matrix.

### WebGL Integration (WASM-only)
- **`UploadSliceToGL(...)`**: A utility function (available only when compiling for `js/wasm`) to efficiently upload numeric Go slices (`[]float32`, `[]uint16`, etc.) to a WebGL buffer on the GPU. This is separated by build tags to allow the core math library to be tested on the server side.

## Usage
To use this package, import it into your Go files:
```go
import "github.com/sbecker11/webgl-point-cloud/glf32"
```

The pure math portions can be used in any Go environment. The `UploadSliceToGL` function requires a `js/wasm` build target.

To run the associated tests:
```bash
go test
```

## MVP (Model-View-Projection) Example

Here is a complete example of how to generate all the necessary data to render a 1x1x1 cube with randomly colored faces. The resulting `mvpMatrix`, `vertexData`, and `colorData` can be passed directly to a WebGL program as uniforms and attribute buffers.

```go
package main

import (
	"fmt"
	"math"
	"math/rand"
	"time"
	"github.com/sbecker11/webgl-point-cloud/glf32"
)

// generateCubeData creates the vertex and color data for a 1x1x1 cube.
// To have distinct colors per face, we must define 36 vertices (6 faces * 2 triangles/face * 3 vertices/triangle).
func generateCubeData() (vertices, colors []float32) {
	// Define the 8 corners of a 1x1x1 cube centered at the origin.
	p := []glf32.Vec3{
		{-0.5, -0.5, 0.5}, // 0: Front, bottom, left
		{0.5, -0.5, 0.5},  // 1: Front, bottom, right
		{0.5, 0.5, 0.5},   // 2: Front, top, right
		{-0.5, 0.5, 0.5},  // 3: Front, top, left
		{-0.5, -0.5, -0.5}, // 4: Back, bottom, left
		{0.5, -0.5, -0.5},  // 5: Back, bottom, right
		{0.5, 0.5, -0.5},   // 6: Back, top, right
		{-0.5, 0.5, -0.5},  // 7: Back, top, left
	}
	
	// Define the 6 faces by referencing the corners. Two triangles per face.
	quads := [][]int{
		{0, 1, 2, 3}, // Front face
		{1, 5, 6, 2}, // Right face
		{5, 4, 7, 6}, // Back face
		{4, 0, 3, 7}, // Left face
		{3, 2, 6, 7}, // Top face
		{4, 5, 1, 0}, // Bottom face
	}

	// Generate a random color for each of the 6 faces.
	faceColors := make([]glf32.Vec3, 6)
	for i := 0; i < 6; i++ {
		faceColors[i] = glf32.Vec3{rand.Float32(), rand.Float32(), rand.Float32()}
	}
	
	// Assemble the vertex and color arrays.
	vertexData := []float32{}
	colorData := []float32{}

	for i, quad := range quads {
		// First triangle of the quad
		vertexData = append(vertexData, p[quad[0]]...)
		vertexData = append(vertexData, p[quad[1]]...)
		vertexData = append(vertexData, p[quad[2]]...)
		// Second triangle of the quad
		vertexData = append(vertexData, p[quad[0]]...)
		vertexData = append(vertexData, p[quad[2]]...)
		vertexData = append(vertexData, p[quad[3]]...)

		// Add the same color for all 6 vertices of this face.
		for v := 0; v < 6; v++ {
			colorData = append(colorData, faceColors[i]...)
		}
	}
	
	return vertexData, colorData
}


func main() {
	rand.Seed(time.Now().UnixNano())

	// 1. Generate Geometry and Colors for the Model
	vertexData, colorData := generateCubeData()

	// 2. Model Matrix: Defines the position, rotation, and scale of the object.
	// In a real application, this would be updated every frame.
	// Here, we'll set a rotation of 45 degrees around the Y axis.
	angle := float32(math.Pi / 4) // Example angle, would change each frame
	modelMatrix := glf32.RotateY(angle)

	// 3. View Matrix: Defines the camera's position and orientation.
	// Let's place the camera at (2, 2, 2) and have it look at the origin.
	eye := glf32.Vec3{2, 2, 2}
	center := glf32.Vec3{0, 0, 0}
	up := glf32.Vec3{0, 1, 0}
	viewMatrix := glf32.LookAt(eye, center, up)

	// 4. Projection Matrix: Defines the camera's viewing frustum.
	// This creates a perspective effect.
	fov := float32(math.Pi / 4) // 45-degree field of view
	aspect := float32(16.0 / 9.0) // Widescreen aspect ratio
	near := float32(0.1)
	far := float32(100.0)
	projectionMatrix := glf32.Perspective(fov, aspect, near, far)

	// 5. Combine Matrices: Create the final MVP matrix.
	// The correct order is Projection * View * Model.
	viewModelMatrix := glf32.MultiplyMatrices(viewMatrix, modelMatrix)
	mvpMatrix := glf32.MultiplyMatrices(projectionMatrix, viewModelMatrix)

	// The `mvpMatrix` (uniform), `vertexData` (attribute), and `colorData` (attribute)
	// are now ready to be sent to a WebGL program.
	fmt.Printf("Generated %d vertices and %d colors.\n", len(vertexData)/3, len(colorData)/3)
	fmt.Println("MVP Matrix ready for shader:")
	glf32.PrintMat4("mvp", mvpMatrix)
} 