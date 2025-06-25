// wasm/webgl_utils.go
package main

import (
	"fmt"
	"reflect"
	"syscall/js"
	"unsafe"
)

// sliceToJsFloat32Array converts a Go slice to a JavaScript Float32Array by
// copying the data. This is a safer approach than creating a view, as it
// prevents "detached ArrayBuffer" errors if the Go WASM memory is resized.
func sliceToJsFloat32Array(slice []float32) js.Value {
	// Create a new JavaScript ArrayBuffer of the required size.
	jsArray := js.Global().Get("Uint8Array").New(len(slice) * 4)

	// Create a Go byte slice that views the same memory as the float32 slice
	header := (*reflect.SliceHeader)(unsafe.Pointer(&slice))
	header.Len *= 4
	header.Cap *= 4
	byteSlice := *(*[]byte)(unsafe.Pointer(header))

	// Copy the data from Go to JavaScript.
	js.CopyBytesToJS(jsArray, byteSlice)

	// Restore the slice header to its original state to avoid memory corruption.
	header.Len /= 4
	header.Cap /= 4

	// Create a Float32Array view on the new buffer.
	return js.Global().Get("Float32Array").New(jsArray.Get("buffer"))
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