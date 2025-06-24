// wasm/wasm_main.go
// build: GOOS=js GOARCH=wasm go build -o wasm/main.wasm wasm/wasm_main.go
// usage: executed by wasm/index.html

package main


import (
	"math"
	"math/rand"
	"syscall/js"
	"time"
	"unsafe"
	"github.com/sbecker11/webgl-point-cloud/glf32"
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

// drawObject is a helper function that encapsulates the WebGL calls needed to draw a single object.
// It binds the position and color buffers, sets the attribute pointers, and issues a draw call.
func drawObject(gl, positionLoc, colorLoc, posBuf, colorBuf, drawMode js.Value, vertexCount int) {
	// Bind position buffer
	gl.Call("bindBuffer", gl.Get("ARRAY_BUFFER"), posBuf)
	gl.Call("vertexAttribPointer", positionLoc, 3, gl.Get("FLOAT"), false, 0, 0)

	// Bind color buffer
	gl.Call("bindBuffer", gl.Get("ARRAY_BUFFER"), colorBuf)
	gl.Call("vertexAttribPointer", colorLoc, 3, gl.Get("FLOAT"), false, 0, 0)

	// Draw the object
	gl.Call("drawArrays", drawMode, 0, vertexCount)
}

func main() {
	// Seed random number generator
	rand.Seed(time.Now().UnixNano())
	js.Global().Get("console").Call("log", "WASM module started")

	// --- Global Variables ---
	// Encapsulate WebGL state and matrices to be accessible by resize handler
	var gl js.Value
	var program js.Value
	var mvpLoc js.Value
	var viewMatrix, projMatrix glf32.Mat4
	var angle float32

	// Get canvas
	canvas := js.Global().Get("document").Call("getElementById", "canvas")
	if canvas.IsNull() {
		js.Global().Call("alert", "Canvas element not found")
		return
	}

	// --- WebGL and Data Setup ---
	// Most of the one-time setup code remains the same...
	gl = canvas.Call("getContext", "webgl")
	if gl.IsNull() {
		js.Global().Call("alert", "WebGL not supported")
		return
	}
	gl.Call("enable", gl.Get("DEPTH_TEST"))
	gl.Call("lineWidth", 8.0)
	gl.Call("clearColor", 0.0, 0.1, 0.25, 1.0)

	numElements := 1000
	pointSize := 3.0
	coordinates := make([]float32, numElements*3)
	pointsGenerated := 0
	for pointsGenerated < numElements {
		x := 2*rand.Float32() - 1
		y := 2*rand.Float32() - 1
		z := 2*rand.Float32() - 1
		if x*x+y*y+z*z <= 1.0 {
			coordinates[pointsGenerated*3] = x
			coordinates[pointsGenerated*3+1] = y
			coordinates[pointsGenerated*3+2] = z
			pointsGenerated++
		}
	}
	pointColors := make([]float32, numElements*3)
	for i := 0; i < numElements; i++ {
		pointColors[i*3], pointColors[i*3+1], pointColors[i*3+2] = 1.0, 1.0, 1.0
	}
	yAxisCylinder := generateCylinder(0.02, 2.0, 16)
	xAxisCylinder := applyMatrixToVec3s(yAxisCylinder, glf32.RotateZ(-math.Pi/2))
	zAxisCylinder := applyMatrixToVec3s(yAxisCylinder, glf32.RotateX(math.Pi/2))
	axisVertices := append(append(xAxisCylinder, yAxisCylinder...), zAxisCylinder...)
	numAxisVertices := len(axisVertices) / 3
	axisColors := make([]float32, numAxisVertices*3)
	numCylVerts := len(yAxisCylinder) / 3
	for i := 0; i < numCylVerts; i++ {
		axisColors[i*3], axisColors[i*3+1], axisColors[i*3+2] = 1.0, 0.0, 0.0
	}
	for i := 0; i < numCylVerts; i++ {
		offset := numCylVerts * 3
		axisColors[offset+i*3], axisColors[offset+i*3+1], axisColors[offset+i*3+2] = 0.0, 1.0, 0.0
	}
	for i := 0; i < numCylVerts; i++ {
		offset := (numCylVerts * 2) * 3
		axisColors[offset+i*3], axisColors[offset+i*3+1], axisColors[offset+i*3+2] = 0.0, 0.0, 1.0
	}
	baseCircle := generateCircle(1.0, 64)
	xCircleVertices := applyMatrixToVec3s(baseCircle, glf32.RotateY(math.Pi/2))
	yCircleVertices := applyMatrixToVec3s(baseCircle, glf32.RotateX(math.Pi/2))
	zCircleVertices := baseCircle

	// Combine all circle vertices
	circleVertices := append(append(xCircleVertices, yCircleVertices...), zCircleVertices...)
	numCircleVertices := len(circleVertices) / 3

	// --- Assign colors to circles to match their corresponding axis ---
	// The circle on the YZ plane (normal=X) is Red.
	// The circle on the XZ plane (normal=Y) is Green.
	// The circle on the XY plane (normal=Z) is Blue.
	numSegCircleVerts := len(baseCircle) / 3 // Vertices per individual circle

	xCircleColors := make([]float32, numSegCircleVerts*3)
	for i := 0; i < numSegCircleVerts; i++ { // Red for X-plane circle
		xCircleColors[i*3+0] = 1.0
		xCircleColors[i*3+1] = 0.0
		xCircleColors[i*3+2] = 0.0
	}
	yCircleColors := make([]float32, numSegCircleVerts*3)
	for i := 0; i < numSegCircleVerts; i++ { // Green for Y-plane circle
		yCircleColors[i*3+0] = 0.0
		yCircleColors[i*3+1] = 1.0
		yCircleColors[i*3+2] = 0.0
	}
	zCircleColors := make([]float32, numSegCircleVerts*3)
	for i := 0; i < numSegCircleVerts; i++ { // Blue for Z-plane circle
		zCircleColors[i*3+0] = 0.0
		zCircleColors[i*3+1] = 0.0
		zCircleColors[i*3+2] = 1.0
	}

	// Combine all circle colors in the same order as the vertices
	circleColors := append(append(xCircleColors, yCircleColors...), zCircleColors...)

	pointPositionBuffer := glf32.UploadSliceToGL(gl, coordinates, "ARRAY_BUFFER", gl.Get("STATIC_DRAW"))
	pointColorBuffer := glf32.UploadSliceToGL(gl, pointColors, "ARRAY_BUFFER", gl.Get("STATIC_DRAW"))
	axisPositionBuffer := glf32.UploadSliceToGL(gl, axisVertices, "ARRAY_BUFFER", gl.Get("STATIC_DRAW"))
	axisColorBuffer := glf32.UploadSliceToGL(gl, axisColors, "ARRAY_BUFFER", gl.Get("STATIC_DRAW"))
	circlePositionBuffer := glf32.UploadSliceToGL(gl, circleVertices, "ARRAY_BUFFER", gl.Get("STATIC_DRAW"))
	circleColorBuffer := glf32.UploadSliceToGL(gl, circleColors, "ARRAY_BUFFER", gl.Get("STATIC_DRAW"))

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
	fragmentShaderSource := `
		precision mediump float;
		varying vec4 vColor;
		void main() {
			gl_FragColor = vColor;
		}
	`

	vertexShader := gl.Call("createShader", gl.Get("VERTEX_SHADER"))
	gl.Call("shaderSource", vertexShader, vertexShaderSource)
	gl.Call("compileShader", vertexShader)

	fragmentShader := gl.Call("createShader", gl.Get("FRAGMENT_SHADER"))
	gl.Call("shaderSource", fragmentShader, fragmentShaderSource)
	gl.Call("compileShader", fragmentShader)

	program = gl.Call("createProgram")
	gl.Call("attachShader", program, vertexShader)
	gl.Call("attachShader", program, fragmentShader)
	gl.Call("linkProgram", program)
	gl.Call("useProgram", program)

	positionLoc := gl.Call("getAttribLocation", program, "position")
	colorLoc := gl.Call("getAttribLocation", program, "color")
	mvpLoc = gl.Call("getUniformLocation", program, "modelViewProjection")
	pointSizeLoc := gl.Call("getUniformLocation", program, "pointSize")

	gl.Call("enableVertexAttribArray", positionLoc)
	gl.Call("enableVertexAttribArray", colorLoc)
	gl.Call("uniform1f", pointSizeLoc, float32(pointSize))

	// --- Animation and Resize Loop ---
	var renderFrame js.Func
	var resizeHandler js.Func

	// update function recalculates matrices and viewport based on window size
	update := func() {
		// Get window inner dimensions
		winWidth := js.Global().Get("innerWidth").Float()
		winHeight := js.Global().Get("innerHeight").Float()

		// Determine the smaller dimension to create a square canvas
		size := math.Min(winWidth, winHeight)

		// Update canvas drawing buffer size
		canvas.Set("width", size)
		canvas.Set("height", size)

		// Update WebGL viewport
		gl.Call("viewport", 0, 0, size, size)

		// --- Update Camera and Projection ---
		camera_distance := float32(3.0)
		aspect := float32(1.0) // Always 1.0 for a square canvas
		fov := float32(math.Pi / 4)
		near, far := float32(0.1), float32(100.0)

		projMatrix = glf32.Perspective(fov, aspect, near, far)
		eyeVec := glf32.Vec3{camera_distance, camera_distance, camera_distance}
		centerVec := glf32.Vec3{0.0, 0.0, 0.0}
		upVec := glf32.Vec3{0.0, 1.0, 0.0}
		viewMatrix = glf32.LookAt(eyeVec, centerVec, upVec)
	}

	// renderFrame function performs the drawing for each animation frame
	renderFrame = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		angle += 0.01
		if angle > 2*math.Pi {
			angle -= 2 * math.Pi
		}

		modelMatrix := glf32.RotateY(angle)
		viewModelMatrix := glf32.MultiplyMatrices(viewMatrix, modelMatrix)
		finalMVPMatrix := glf32.MultiplyMatrices(projMatrix, viewModelMatrix)

		byteLen := len(finalMVPMatrix) * 4
		byteSlice := unsafe.Slice((*byte)(unsafe.Pointer(&finalMVPMatrix[0])), byteLen)
		jsBytes := js.Global().Get("Uint8Array").New(byteLen)
		js.CopyBytesToJS(jsBytes, byteSlice)
		jsMvp := js.Global().Get("Float32Array").New(jsBytes.Get("buffer"))
		gl.Call("uniformMatrix4fv", mvpLoc, false, jsMvp)

		gl.Call("clear", gl.Get("COLOR_BUFFER_BIT").Int()|gl.Get("DEPTH_BUFFER_BIT").Int())

		// Draw the scene objects using the helper function
		drawObject(gl, positionLoc, colorLoc, pointPositionBuffer, pointColorBuffer, gl.Get("POINTS"), numElements)
		drawObject(gl, positionLoc, colorLoc, axisPositionBuffer, axisColorBuffer, gl.Get("TRIANGLES"), numAxisVertices)
		drawObject(gl, positionLoc, colorLoc, circlePositionBuffer, circleColorBuffer, gl.Get("LINES"), numCircleVertices)

		js.Global().Call("requestAnimationFrame", renderFrame)
		return nil
	})

	// resizeHandler updates the canvas and matrices, then triggers a redraw
	resizeHandler = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		update()
		return nil
	})

	// --- Start Application ---
	// Set initial size
	update()
	// Add resize event listener
	js.Global().Call("addEventListener", "resize", resizeHandler)
	// Start animation loop
	js.Global().Call("requestAnimationFrame", renderFrame)
	js.Global().Get("console").Call("log", "Animation loop started")

	// Keep Go from exiting
	<-make(chan struct{})

	// Release JS functions when done (though this part is never reached in this app)
	defer renderFrame.Release()
	defer resizeHandler.Release()
}
