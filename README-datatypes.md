Here's a single function that:

Accepts a []float32 slice (e.g., vertex data),

Accepts the WebGL buffer target (either "ARRAY_BUFFER" or "ELEMENT_ARRAY_BUFFER"),

Accepts the usage parameter (e.g., gl.Get("STATIC_DRAW")),

Handles all binding and uploading,

Returns the WebGL buffer js.Value.

```go
import (
    "syscall/js"
    "unsafe"
)

// UploadFloat32ToGL uploads a []float32 to WebGL and returns the created buffer.
// `target` should be "ARRAY_BUFFER" or "ELEMENT_ARRAY_BUFFER".
// `usage` is typically gl.Get("STATIC_DRAW") or gl.Get("DYNAMIC_DRAW").
func UploadFloat32ToGL(gl js.Value, data []float32, target string, usage js.Value) js.Value {
    if len(data) == 0 {
        panic("UploadFloat32ToGL: data slice is empty")
    }

    // Convert to []byte
    byteSlice := unsafe.Slice((*byte)(unsafe.Pointer(&data[0])), len(data)*4)

    // Create and bind buffer
    buffer := gl.Call("createBuffer")
    gl.Call("bindBuffer", gl.Get(target), buffer)

    // Transfer data to JS as Float32Array
    jsBytes := js.Global().Get("Uint8Array").New(len(byteSlice))
    js.CopyBytesToJS(jsBytes, byteSlice)
    jsFloat32Array := js.Global().Get("Float32Array").New(jsBytes.Get("buffer"))

    // Upload to WebGL
    gl.Call("bufferData", gl.Get(target), jsFloat32Array, usage)

    return buffer
}
```

🧪 Example Usage
``` go
Vertex Buffer (ARRAY_BUFFER):
positionBuffer := UploadFloat32ToGL(gl, positions, "ARRAY_BUFFER", gl.Get("STATIC_DRAW"))

Index Buffer (ELEMENT_ARRAY_BUFFER):
indexBuffer := UploadFloat32ToGL(gl, indices, "ELEMENT_ARRAY_BUFFER", gl.Get("STATIC_DRAW"))
```

Here's a generalized and reusable function that:

✅ Works with:

[]float32 → Float32Array

[]uint16 → Uint16Array

[]uint32 → Uint32Array

✅ Supports both WebGL buffer targets:

"ARRAY_BUFFER"

"ELEMENT_ARRAY_BUFFER"

✅ Automatically selects the correct JS TypedArray based on the Go slice type.

✅ Final Generic Upload Function

``` go
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

// Example Usages
```go
// For Vertex Positions ([]float32)
positionBuffer := UploadSliceToGL(gl, positions, "ARRAY_BUFFER", gl.Get("STATIC_DRAW"))

// For Triangle Indices ([]uint16)
indexBuffer := UploadSliceToGL(gl, indices, "ELEMENT_ARRAY_BUFFER", gl.Get("STATIC_DRAW"))

// For []uint32 index buffers (if using WebGL2)
if gl.Call("getExtension", "OES_element_index_uint").Truthy() {
    indexBuffer := UploadSliceToGL(gl, bigIndices, "ELEMENT_ARRAY_BUFFER", gl.Get("STATIC_DRAW"))
}
```

Automatically handles the conversion.

Automatically chooses the right JS TypedArray.

Works for all your common GPU buffer types in WebGL and WebGL2.


Here’s a quick summary of the common WebGL buffer types and their typical associated JavaScript TypedArrays (and corresponding Go slice types):

WebGL Buffer Types & Typical TypedArray Usage
WebGL Buffer Target	Data Usage	Typical JS TypedArray	Corresponding Go Slice Type
ARRAY_BUFFER	Vertex attributes (positions, colors, normals, texture coords)	Float32Array	[]float32
ELEMENT_ARRAY_BUFFER	Index data (which vertices to use for triangles, lines, etc.)	Uint16Array or Uint32Array (WebGL2)	[]uint16 or []uint32

Details:
ARRAY_BUFFER:
Stores vertex data attributes such as positions (x,y,z), colors (r,g,b,a), normals, texture coordinates, etc.
Uses Float32Array in JavaScript because vertex attributes are almost always floats.
In Go, this maps naturally to []float32.

ELEMENT_ARRAY_BUFFER:
Stores indices for vertex order when drawing primitives (triangles, lines, points).
Uses Uint16Array by default since WebGL1 requires 16-bit indices.
WebGL2 supports Uint32Array for larger meshes.
Corresponding Go types are []uint16 and []uint32.

Bonus Notes:
WebGL does not use integer types like int32, int16, or unsigned 8-bit (Uint8Array) for vertex attributes or indices generally.

For colors, sometimes Uint8Array is used with normalization, but usually colors are sent as Float32Array in the 0–1 range.

WebGL expects tightly packed typed arrays matching these types for maximum performance.

