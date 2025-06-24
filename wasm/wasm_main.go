package main

import (
	"fmt"
	"math"
	"math/rand"
	"syscall/js"
	"time"
	"github.com/sbecker11/webgl-point-cloud/glf32" // Import your new package. Make sure your go.mod references this path.
)

// applyMatrixToVec3s applies a 4x4 column-major matrix to a slice of 3D vertex coordinates (x,y,z).
// It treats each vertex as (x,y,z,1) and performs matrix * vector multiplication.
// It returns a NEW slice of transformed 3D vertices, without performing perspective divide.
// This is suitable for pre-transforming static model data (like axes or circles) before
// the main Model-View-Projection pipeline in the shader.
func applyMatrixToVec3s(coords []float32, m glf32.Mat4) []float32 {
	if len(m) != 16 {
		panic("applyMatrixToVec3s: matrix must be 4x4 (length 16)")
	}
	if len(coords)%3 != 0 {
		panic("applyMatrixToVec3s: coords slice length must be a multiple of 3")
	}

	transformed := make([]float32, len(coords))
	numVertices := len(coords) / 3

	for i := 0; i < numVertices; i++ {
		idx := i * 3
		x, y, z := coords[idx], coords[idx+1], coords[idx+2]
		w := float32(1.0) // Assume w=1 for homogeneous coordinates

		// Perform M * V (column-major multiplication)
		// newX = M[0][0]*x + M[0][1]*y + M[0][2]*z + M[0][3]*w
		// In column-major memory layout (m[col*4 + row]):
		transformed[idx] = m[0]*x + m[4]*y + m[8]*z + m[12]*w   // newX
		transformed[idx+1] = m[1]*x + m[5]*y + m[9]*z + m[13]*w // newY
		transformed[idx+2] = m[2]*x + m[6]*y + m[10]*z + m[14]*w // newZ
		// Note: The W component (m[3]*x + m[7]*y + m[11]*z + m[15]*w) is calculated
		// but not stored/divided here, as this is a model-space transformation.
		// The final perspective divide happens in the shader with gl_Position.
	}
	return transformed
}

// generateCylinder creates vertices for a cylinder.
// It generates triangles for the body and caps.
func generateCylinder(radius float32, height float32, segments int) []float32 {
	vertices := []float32{}
	h2 := height / 2.0 // Half-height for centered cylinder
	angleStep := float32(2.0 * math.Pi / float64(segments))

	for i := 0; i < segments; i++ {
		a1 := float32(i) * angleStep
		a2 := float32(i+1) * angleStep
		// Coordinates for the current and next segment on a circle
		x1, z1 := radius*float32(math.Cos(float64(a1))), radius*float32(math.Sin(float64(a1)))
		x2, z2 := radius*float32(math.Cos(float64(a2))), radius*float32(math.Sin(float64(a2)))

		// Define the four corners of the current segment rectangle
		v1 := glf32.Vec3{x1, -h2, z1} // Bottom-left
		v2 := glf32.Vec3{x2, -h2, z2} // Bottom-right
		v3 := glf32.Vec3{x1, h2, z1}  // Top-left
		v4 := glf32.Vec3{x2, h2, z2}  // Top-right

		// Body (two triangles forming a quad)
		// Triangle 1: v1, v2, v3
		vertices = append(vertices, v1...)
		vertices = append(vertices, v2...)
		vertices = append(vertices, v3...)
		// Triangle 2: v3, v2, v4
		vertices = append(vertices, v3...)
		vertices = append(vertices, v2...)
		vertices = append(vertices, v4...)

		// Top cap (triangle fan from center to segment vertices)
		ct := glf32.Vec3{0, h2, 0} // Center of top cap
		vertices = append(vertices, ct...)
		vertices = append(vertices, v3...)
		vertices = append(vertices, v4...)

		// Bottom cap (triangle fan from center to segment vertices, winding reversed for correct normal)
		cb := glf32.Vec3{0, -h2, 0} // Center of bottom cap
		vertices = append(vertices, cb...)
		vertices = append(vertices, v2...)
		vertices = append(vertices, v1...)
	}
	return vertices
}

// generateCircle creates vertices for a circle (as line segments).
// The circle is generated on the XY plane.
func generateCircle(radius float32, segments int) []float32 {
	vertices := []float32{}
	angleStep := float32(2.0 * math.Pi / float64(segments))
	for i := 0; i < segments; i++ {
		a1 := float32(i) * angleStep
		a2 := float32(i+1) * angleStep
		// Coordinates for the current and next point on the circle
		x1, y1 := radius*float32(math.Cos(float64(a1))), radius*float32(math.Sin(float64(a1)))
		x2, y2 := radius*float32(math.Cos(float64(a2))), radius*float32(math.Sin(float64(a2)))

		// Add two points per segment for GL_LINES
		vertices = append(vertices, x1, y1, 0) // Point 1
		vertices = append(vertices, x2, y2, 0) // Point 2
	}
	return vertices
}

func main() {
	// Seed random number generator
	rand.Seed(time.Now().UnixNano())

	// Debug log
	js.Global().Get("console").Call("log", "WASM module started")

	// Ensure the WASM module stays alive
	c := make(chan struct{}, 0)

	// Get canvas and WebGL context
	canvas := js.Global().Get("document").Call("getElementById", "canvas")
	if canvas.IsNull() {
		js.Global().Call("alert", "Canvas element not found")
		return
	}

	gl := canvas.Call("getContext", "webgl")
	if gl.IsNull() {
		js.Global().Call("alert", "WebGL not supported")
		return
	}

	// Enable depth testing
	gl.Call("enable", gl.Get("DEPTH_TEST"))
	gl.Call("lineWidth", 2.0)

	// Set viewport
	width := canvas.Get("width").Int()
	height := canvas.Get("height").Int()
	gl.Call("viewport", 0, 0, width, height)

	// Set up WebGL
	gl.Call("clearColor", 0.0, 0.1, 0.25, 1.0) // Dark blue clear color

	numElements := 1000
	pointSize := 4.0 // Slightly larger points for visibility

	// Create 3D points within a unit sphere using rejection sampling
	coordinates := make([]float32, numElements*3) // numElements * 3 coordinates (x, y, z)
	pointsGenerated := 0
	for pointsGenerated < numElements {
		x := 2*rand.Float32() - 1 // Generate between -1 and 1
		y := 2*rand.Float32() - 1
		z := 2*rand.Float32() - 1

		// Check if point is within unit sphere (distance squared <= 1.0 from origin)
		if x*x+y*y+z*z <= 1.0 {
			coordinates[pointsGenerated*3] = x
			coordinates[pointsGenerated*3+1] = y
			coordinates[pointsGenerated*3+2] = z
			pointsGenerated++ // Only increment if accepted
		}
	}
	js.Global().Get("console").Call("log", fmt.Sprintf("Generated %d points in a unit sphere", numElements))

	pointColors := make([]float32, numElements*3)
	for i := 0; i < numElements; i++ {
		pointColors[i*3] = 1.0 // White points
		pointColors[i*3+1] = 1.0
		pointColors[i*3+2] = 1.0
	}

	// NOTE: cylinderRadius is in world-space units.
	cylinderRadius := float32(0.02)
	cylinderHeight := float32(2.0)
	cylinderSegments := 16

	// Generate base cylinder (aligned with Y-axis)
	yAxisCylinder := generateCylinder(cylinderRadius, cylinderHeight, cylinderSegments)

	// Transform for X-axis (Rotate Y-aligned cylinder to be X-aligned)
	// Rotation around Z by -PI/2 (90 degrees clockwise from Y to X)
	rotZToX := glf32.RotateZ(float32(-math.Pi / 2))
	xAxisCylinder := applyMatrixToVec3s(yAxisCylinder, rotZToX)

	// Transform for Z-axis (Rotate Y-aligned cylinder to be Z-aligned)
	// Rotation around X by PI/2 (90 degrees counter-clockwise from Y to Z)
	rotXToZ := glf32.RotateX(float32(math.Pi / 2))
	zAxisCylinder := applyMatrixToVec3s(yAxisCylinder, rotXToZ)

	// Combine all axis vertices
	axisVertices := append(append(xAxisCylinder, yAxisCylinder...), zAxisCylinder...)
	numAxisVertices := len(axisVertices) / 3

	// Assign colors to axes (Red for X, Green for Y, Blue for Z)
	axisColors := make([]float32, numAxisVertices*3)
	numCylinderBodyVertices := len(yAxisCylinder) / 3 // Vertices per axis-cylinder
	for i := 0; i < numCylinderBodyVertices; i++ { // X-axis (Red)
		axisColors[i*3+0] = 1.0 // R
		axisColors[i*3+1] = 0.0 // G
		axisColors[i*3+2] = 0.0 // B
	}
	for i := 0; i < numCylinderBodyVertices; i++ { // Y-axis (Green)
		offset := numCylinderBodyVertices * 3
		axisColors[offset+i*3+0] = 0.0 // R
		axisColors[offset+i*3+1] = 1.0 // G
		axisColors[offset+i*3+2] = 0.0 // B
	}
	for i := 0; i < numCylinderBodyVertices; i++ { // Z-axis (Blue)
		offset := (numCylinderBodyVertices * 2) * 3
		axisColors[offset+i*3+0] = 0.0 // R
		axisColors[offset+i*3+1] = 0.0 // G
		axisColors[offset+i*3+2] = 1.0 // B
	}

	// Generate circles for axis planes
	circleSegments := 64
	circleRadius := float32(1.0)
	baseCircle := generateCircle(circleRadius, circleSegments) // Base circle on XY plane

	// Z-axis circle (blue, on XY plane - no rotation needed from base)
	zCircleVertices := baseCircle

	// X-axis circle (red, transform XY circle to YZ plane)
	// Rotate around Y by 90 degrees (to move X to Z, Z to -X)
	rot90y := glf32.RotateY(float32(math.Pi / 2))
	xCircleVertices := applyMatrixToVec3s(baseCircle, rot90y)

	// Y-axis circle (green, transform XY circle to XZ plane)
	// Rotate around X by 90 degrees (to move Y to Z, Z to -Y)
	rot90x := glf32.RotateX(float32(math.Pi / 2))
	yCircleVertices := applyMatrixToVec3s(baseCircle, rot90x)

	// Combine all circle vertices
	circleVertices := append(append(xCircleVertices, yCircleVertices...), zCircleVertices...)
	numCircleVertices := len(circleVertices) / 3
	numSegCircleVertices := len(baseCircle) / 3 // Vertices per circle

	// Assign colors to circles (Red for X-plane, Green for Y-plane, Blue for Z-plane)
	circleColors := make([]float32, numCircleVertices*3)
	for i := 0; i < numSegCircleVertices; i++ { // X-plane Circle (Red)
		circleColors[i*3+0] = 1.0 // R
		circleColors[i*3+1] = 0.0 // G
		circleColors[i*3+2] = 0.0 // B
	}
	for i := 0; i < numSegCircleVertices; i++ { // Y-plane Circle (Green)
		offset := numSegCircleVertices * 3
		circleColors[offset+i*3+0] = 0.0 // R
		circleColors[offset+i*3+1] = 1.0 // G
		circleColors[offset+i*3+2] = 0.0 // B
	}
	for i := 0; i < numSegCircleVertices; i++ { // Z-plane Circle (Blue)
		offset := (numSegCircleVertices * 2) * 3
		circleColors[offset+i*3+0] = 0.0 // R
		circleColors[offset+i*3+1] = 0.0 // G
		circleColors[offset+i*3+2] = 1.0 // B
	}

	js.Global().Get("console").Call("log", "Geometry generated")

	// --- WebGL Buffer Setup ---
	pointPositionBuffer := gl.Call("createBuffer")
	gl.Call("bindBuffer", gl.Get("ARRAY_BUFFER"), pointPositionBuffer)
	gl.Call("bufferData", gl.Get("ARRAY_BUFFER"), js.Global().Get("Float32Array").New(js.ValueOf(coordinates)), gl.Get("STATIC_DRAW"))

	pointColorBuffer := gl.Call("createBuffer")
	gl.Call("bindBuffer", gl.Get("ARRAY_BUFFER"), pointColorBuffer)
	gl.Call("bufferData", gl.Get("ARRAY_BUFFER"), js.Global().Get("Float32Array").New(js.ValueOf(pointColors)), gl.Get("STATIC_DRAW"))

	axisPositionBuffer := gl.Call("createBuffer")
	gl.Call("bindBuffer", gl.Get("ARRAY_BUFFER"), axisPositionBuffer)
	gl.Call("bufferData", gl.Get("ARRAY_BUFFER"), js.Global().Get("Float32Array").New(js.ValueOf(axisVertices)), gl.Get("STATIC_DRAW"))

	axisColorBuffer := gl.Call("createBuffer")
	gl.Call("bindBuffer", gl.Get("ARRAY_BUFFER"), axisColorBuffer)
	gl.Call("bufferData", gl.Get("ARRAY_BUFFER"), js.Global().Get("Float32Array").New(js.ValueOf(axisColors)), gl.Get("STATIC_DRAW"))

	circlePositionBuffer := gl.Call("createBuffer")
	gl.Call("bindBuffer", gl.Get("ARRAY_BUFFER"), circlePositionBuffer)
	gl.Call("bufferData", gl.Get("ARRAY_BUFFER"), js.Global().Get("Float32Array").New(js.ValueOf(circleVertices)), gl.Get("STATIC_DRAW"))

	circleColorBuffer := gl.Call("createBuffer")
	gl.Call("bindBuffer", gl.Get("ARRAY_BUFFER"), circleColorBuffer)
	gl.Call("bufferData", gl.Get("ARRAY_BUFFER"), js.Global().Get("Float32Array").New(js.ValueOf(circleColors)), gl.Get("STATIC_DRAW"))
	js.Global().Get("console").Call("log", "Buffer data set")

	// --- Shader Setup ---
	vertexShaderSource := `
        attribute vec3 position;
        attribute vec3 color;
        uniform mat4 modelViewProjection;
        uniform float pointSize;
        varying vec4 vColor;
        void main() {
            // Standard column-major multiplication: Matrix * Vector
            gl_Position = modelViewProjection * vec4(position, 1.0);
            gl_PointSize = pointSize;
            vColor = vec4(color, 1.0);
        }
    `
	vertexShader := gl.Call("createShader", gl.Get("VERTEX_SHADER"))
	gl.Call("shaderSource", vertexShader, vertexShaderSource)
	gl.Call("compileShader", vertexShader)
	if !gl.Call("getShaderParameter", vertexShader, gl.Get("COMPILE_STATUS")).Bool() {
		js.Global().Call("alert", "Vertex shader failed: "+gl.Call("getShaderInfoLog", vertexShader).String())
		return
	}

	fragmentShaderSource := `
        precision mediump float;
        varying vec4 vColor;
        void main() {
            gl_FragColor = vColor;
        }
    `
	fragmentShader := gl.Call("createShader", gl.Get("FRAGMENT_SHADER"))
	gl.Call("shaderSource", fragmentShader, fragmentShaderSource)
	gl.Call("compileShader", fragmentShader)
	if !gl.Call("getShaderParameter", fragmentShader, gl.Get("COMPILE_STATUS")).Bool() {
		js.Global().Call("alert", "Fragment shader failed: "+gl.Call("getShaderInfoLog", fragmentShader).String())
		return
	}

	program := gl.Call("createProgram")
	gl.Call("attachShader", program, vertexShader)
	gl.Call("attachShader", program, fragmentShader)
	gl.Call("linkProgram", program)
	if !gl.Call("getProgramParameter", program, gl.Get("LINK_STATUS")).Bool() {
		js.Global().Call("alert", "Program link failed: "+gl.Call("getProgramInfoLog", program).String())
		return
	}
	gl.Call("useProgram", program)

	positionLoc := gl.Call("getAttribLocation", program, "position")
	gl.Call("enableVertexAttribArray", positionLoc)

	colorLoc := gl.Call("getAttribLocation", program, "color")
	gl.Call("enableVertexAttribArray", colorLoc)

	mvpLoc := gl.Call("getUniformLocation", program, "modelViewProjection")
	if mvpLoc.IsNull() {
		js.Global().Call("alert", "Failed to get modelViewProjection uniform")
		return
	}

	pointSizeLoc := gl.Call("getUniformLocation", program, "pointSize")
	if pointSizeLoc.IsNull() {
		js.Global().Call("alert", "Failed to get pointSize uniform")
		return
	}
	gl.Call("uniform1f", pointSizeLoc, float32(pointSize))

	// --- Camera and Projection Setup ---
	camera_distance := float32(3.0) // Adjusted distance for better view of unit sphere (was 1000.0)
	aspect = float32(width) / float32(height) // Ensure aspect ratio matches canvas
	fov = float32(math.Pi / 4) // 45 degrees vertical FOV
	near, far = float32(0.1), float32(100.0)

	projMatrix := glf32.Perspective(fov, aspect, near, far) // Using glf32.Perspective
	
	// Define camera eye, center, and up as glf32.Vec3 types
	eyeVec := glf32.Normalize(glf32.Vec3{1, 1, 1}) // Normalize direction
	eyeVec = glf32.Vec3{eyeVec[0] * camera_distance, eyeVec[1] * camera_distance, eyeVec[2] * camera_distance} // Scale by distance
	centerVec := glf32.Vec3{0.0, 0.0, 0.0}
	upVec := glf32.Vec3{0.0, 1.0, 0.0}
	viewMatrix := glf32.LookAt(eyeVec, centerVec, upVec) // Using glf32.LookAt

	// --- Animation Loop ---
	var render js.Func
	var angle float32

	render = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		angle += 0.01 // Increment rotation angle for animation

		// Model Matrix: Rotation of the entire scene (the sphere and axes) around Y-axis
		// Use glf32.RotateY which returns a column-major matrix
		modelMatrix := glf32.RotateY(angle)

		// Calculate MVP Matrix (Projection * View * Model) for column-major
		// M_final = M_projection * M_view * M_model
		// Step 1: Combine View and Model
		viewModelMatrix := glf32.MultiplyMatrices(viewMatrix, modelMatrix)
		// Step 2: Combine Projection with (View * Model)
		finalMVPMatrix := glf32.MultiplyMatrices(projMatrix, viewModelMatrix)

		// Convert the Go Mat4 (slice of float32) to JavaScript Float32Array
		jsMvp := js.Global().Get("Float32Array").New(js.ValueOf(finalMVPMatrix))

		// Set the MVP uniform in the shader. 'false' indicates no transpose needed
		// because both our Go matrices and WebGL/GLSL are column-major.
		gl.Call("uniformMatrix4fv", mvpLoc, false, jsMvp)

		// Clear the canvas and depth buffer
		gl.Call("clear", gl.Get("COLOR_BUFFER_BIT").Int()|gl.Get("DEPTH_BUFFER_BIT").Int())

		// --- Draw the Points ---
		gl.Call("bindBuffer", gl.Get("ARRAY_BUFFER"), pointPositionBuffer)
		gl.Call("vertexAttribPointer", positionLoc, 3, gl.Get("FLOAT"), false, 0, 0)
		gl.Call("bindBuffer", gl.Get("ARRAY_BUFFER"), pointColorBuffer)
		gl.Call("vertexAttribPointer", colorLoc, 3, gl.Get("FLOAT"), false, 0, 0)
		gl.Call("drawArrays", gl.Get("POINTS"), 0, numElements)

		// --- Draw the Axes --- (Cylinders)
		gl.Call("bindBuffer", gl.Get("ARRAY_BUFFER"), axisPositionBuffer)
		gl.Call("vertexAttribPointer", positionLoc, 3, gl.Get("FLOAT"), false, 0, 0)
		gl.Call("bindBuffer", gl.Get("ARRAY_BUFFER"), axisColorBuffer)
		gl.Call("vertexAttribPointer", colorLoc, 3, gl.Get("FLOAT"), false, 0, 0)
		gl.Call("drawArrays", gl.Get("TRIANGLES"), 0, numAxisVertices) // Draw as triangles (solid cylinders)

		// --- Draw the Circles --- (Planes)
		gl.Call("bindBuffer", gl.Get("ARRAY_BUFFER"), circlePositionBuffer)
		gl.Call("vertexAttribPointer", positionLoc, 3, gl.Get("FLOAT"), false, 0, 0)
		gl.Call("bindBuffer", gl.Get("ARRAY_BUFFER"), circleColorBuffer)
		gl.Call("vertexAttribPointer", colorLoc, 3, gl.Get("FLOAT"), false, 0, 0)
		gl.Call("drawArrays", gl.Get("LINES"), 0, numCircleVertices) // Draw as lines (wireframe circles)


		// Request next animation frame
		js.Global().Call("requestAnimationFrame", render)
		return nil
	})
	defer render.Release()

	// Start animation loop
	js.Global().Call("requestAnimationFrame", render)
	js.Global().Get("console").Call("log", "Animation loop started")

	<-c // Keep the Go program running indefinitely
}
