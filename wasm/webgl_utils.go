// wasm/webgl_utils.go
package main

import (
	"fmt"
	"syscall/js"
	"unsafe"
)


// sliceToJsFloat32Array converts a Go slice to a JavaScript Float32Array
func sliceToJsFloat32Array(slice []float32) js.Value {
	if len(slice) == 0 {
		return js.Null()
	}
	// Get the WebAssembly memory from the global `go` instance.
	goInstance := js.Global().Get("go")
	mem := goInstance.Get("_inst").Get("exports").Get("mem")
	buffer := mem.Get("buffer")
	// Create a new Float32Array view over the specified section of the buffer.
	return js.Global().Get("Float32Array").New(buffer, uintptr(unsafe.Pointer(&slice[0])), len(slice))
}

// drawObject is a helper function that encapsulates the WebGL calls needed to draw a single object.
func drawObject(gl, positionLoc, colorLoc, posBuf, colorBuf, drawMode js.Value, vertexCount int) {
	// Bind position buffer
	gl.Call("bindBuffer", gl.Get("ARRAY_BUFFER"), posBuf)
	gl.Call("vertexAttribPointer", positionLoc, 3, gl.Get("FLOAT"), false, 0, 0)

	// Bind color buffer
	gl.Call("bindBuffer", gl.Get("ARRAY_BUFFER"), colorBuf)
	gl.Call("vertexAttribPointer", colorLoc, 4, gl.Get("FLOAT"), false, 0, 0) // 4 components for RGBA

	// Draw the object
	gl.Call("drawArrays", drawMode, 0, vertexCount)
}

// createVBO is a helper function to create a Vertex Buffer Object
func createVBO(gl js.Value, data []float32) js.Value {
	buffer := gl.Call("createBuffer")
	gl.Call("bindBuffer", gl.Get("ARRAY_BUFFER"), buffer)
	jsArray := sliceToJsFloat32Array(data)
	gl.Call("bufferData", gl.Get("ARRAY_BUFFER"), jsArray, gl.Get("STATIC_DRAW"))
	return buffer
}

// createShaderProgram compiles and links the vertex and fragment shaders.
func createShaderProgram(gl js.Value, vertSrc, fragSrc string) (js.Value, error) {
	vertShader := gl.Call("createShader", gl.Get("VERTEX_SHADER"))
	gl.Call("shaderSource", vertShader, vertSrc)
	gl.Call("compileShader", vertShader)
	if !gl.Call("getShaderParameter", vertShader, gl.Get("COMPILE_STATUS")).Bool() {
		log := gl.Call("getShaderInfoLog", vertShader).String()
		return js.Null(), fmt.Errorf("vertex shader compile error: %s", log)
	}

	fragShader := gl.Call("createShader", gl.Get("FRAGMENT_SHADER"))
	gl.Call("shaderSource", fragShader, fragSrc)
	gl.Call("compileShader", fragShader)
	if !gl.Call("getShaderParameter", fragShader, gl.Get("COMPILE_STATUS")).Bool() {
		log := gl.Call("getShaderInfoLog", fragShader).String()
		return js.Null(), fmt.Errorf("fragment shader compile error: %s", log)
	}

	p := gl.Call("createProgram")
	gl.Call("attachShader", p, vertShader)
	gl.Call("attachShader", p, fragShader)
	gl.Call("linkProgram", p)
	if !gl.Call("getProgramParameter", p, gl.Get("LINK_STATUS")).Bool() {
		log := gl.Call("getProgramInfoLog", p).String()
		return js.Null(), fmt.Errorf("shader link error: %s", log)
	}
	return p, nil
} 