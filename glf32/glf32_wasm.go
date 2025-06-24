// glf32/glf32_wasm.go
//go:build js && wasm

package glf32

import (
	"fmt"
	"reflect"
	"syscall/js"
	"unsafe"
)

// UploadSliceToGL uploads a numeric Go slice to a WebGL buffer.
// Accepts []float32, []uint16, or []uint32.
// `target` is either "ARRAY_BUFFER" or "ELEMENT_ARRAY_BUFFER".
// `usage` is usually gl.Get("STATIC_DRAW").
func UploadSliceToGL(gl js.Value, data interface{}, target string, usage js.Value) js.Value {
	v := reflect.ValueOf(data)
	if v.Kind() != reflect.Slice || v.Len() == 0 {
		panic("UploadSliceToGL: data must be a non-empty slice")
	}

	var byteSlice []byte
	var byteLen int
	var jsTypedArrayType string

	switch v.Type().Elem().Kind() {
	case reflect.Float32:
		byteLen = v.Len() * 4
		byteSlice = unsafe.Slice((*byte)(unsafe.Pointer((*float32)(unsafe.Pointer(v.Index(0).UnsafeAddr())))), byteLen)
		jsTypedArrayType = "Float32Array"
	case reflect.Uint16:
		byteLen = v.Len() * 2
		byteSlice = unsafe.Slice((*byte)(unsafe.Pointer((*uint16)(unsafe.Pointer(v.Index(0).UnsafeAddr())))), byteLen)
		jsTypedArrayType = "Uint16Array"
	case reflect.Uint32:
		byteLen = v.Len() * 4
		byteSlice = unsafe.Slice((*byte)(unsafe.Pointer((*uint32)(unsafe.Pointer(v.Index(0).UnsafeAddr())))), byteLen)
		jsTypedArrayType = "Uint32Array"
	default:
		panic(fmt.Sprintf("UploadSliceToGL: unsupported slice type %s", v.Type().Elem()))
	}

	// Create buffer and bind
	buffer := gl.Call("createBuffer")
	gl.Call("bindBuffer", gl.Get(target), buffer)

	// Transfer bytes to JS
	jsBytes := js.Global().Get("Uint8Array").New(len(byteSlice))
	js.CopyBytesToJS(jsBytes, byteSlice)

	// Create correct JS typed array from the same buffer
	jsTypedArray := js.Global().Get(jsTypedArrayType).New(jsBytes.Get("buffer"))

	// Upload to GPU
	gl.Call("bufferData", gl.Get(target), jsTypedArray, usage)

	return buffer
} 