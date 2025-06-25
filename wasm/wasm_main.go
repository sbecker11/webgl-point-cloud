// wasm/wasm_main.go
package main

import (
	"fmt"
	"math/rand"
	"syscall/js"
	"time"

	"github.com/sbecker11/webgl-point-cloud/glf32"
)

var camera *Camera

func main() {
	js.Global().Call("setTimeout", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		go mainLogic()
		return nil
	}), 100)
	<-make(chan bool)
}

func mainLogic() {
	rand.Seed(time.Now().UnixNano())
	js.Global().Get("console").Call("log", "WASM module started")

	canvas := js.Global().Get("document").Call("getElementById", "canvas")
	gl := canvas.Call("getContext", "webgl")
	if gl.IsUndefined() {
		js.Global().Call("alert", "WebGL not supported")
		return
	}

	gl.Call("enable", gl.Get("DEPTH_TEST"))
	gl.Call("enable", gl.Get("BLEND"))
	gl.Call("blendFunc", gl.Get("SRC_ALPHA"), gl.Get("ONE_MINUS_SRC_ALPHA"))
	gl.Call("clearColor", 0.0, 0.1, 0.25, 1.0)

	camera = NewCamera(3.0)
	setupEventHandlers(canvas, gl, camera)

	pointProgram, pointMvpLoc, posLoc, colorLoc, err := setupPointShaders(gl)
	if err != nil {
		js.Global().Get("console").Call("error", "Point shader setup error: "+err.Error())
		return
	}
	lineProgram, lineMvpLoc, err := setupLineShaders(gl)
	if err != nil {
		js.Global().Get("console").Call("error", "Line shader setup error: "+err.Error())
		return
	}

	numPoints := 5000
	redCoords, redColors := generateNormalCluster(numPoints, glf32.Vec3{0.5, 0.5, 0.5}, 0.2, glf32.Vec3{1, 0, 0})
	greenCoords, greenColors := generateNormalCluster(numPoints, glf32.Vec3{-0.5, -0.5, 0.5}, 0.2, glf32.Vec3{0, 1, 0})
	blueCoords, blueColors := generateNormalCluster(numPoints, glf32.Vec3{0.0, 0.5, -0.5}, 0.2, glf32.Vec3{0, 0, 1})
	redPosVBO, redColorVBO := createVBO(gl, redCoords), createVBO(gl, redColors)
	greenPosVBO, greenColorVBO := createVBO(gl, greenCoords), createVBO(gl, greenColors)
	bluePosVBO, blueColorVBO := createVBO(gl, blueCoords), createVBO(gl, blueColors)

	axisCoords, axisColors := generateAxes(1.5)
	gridCoords, gridColors := generateGrid(1.5, 10)
	axisPosVBO, axisColorVBO := createVBO(gl, axisCoords), createVBO(gl, axisColors)
	gridPosVBO, gridColorVBO := createVBO(gl, gridCoords), createVBO(gl, gridColors)
	numAxisVertices := len(axisCoords) / 3
	numGridVertices := len(gridCoords) / 3

	var renderFrame js.Func
	renderFrame = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		camera.ApplyInertia()
		aspect := float32(canvas.Get("width").Float() / canvas.Get("height").Float())
		projMatrix := glf32.Perspective(45.0, aspect, 0.1, 100.0)
		viewMatrix := camera.GetViewMatrix()
		mvpMatrix := glf32.MultiplyMatrices(projMatrix, viewMatrix)

		gl.Call("clear", gl.Get("COLOR_BUFFER_BIT").Int()|gl.Get("DEPTH_BUFFER_BIT").Int())

		gl.Call("useProgram", lineProgram)
		gl.Call("uniformMatrix4fv", lineMvpLoc, false, sliceToJsFloat32Array(mvpMatrix[:]))
		gl.Call("enableVertexAttribArray", posLoc)
		gl.Call("enableVertexAttribArray", colorLoc)
		drawObject(gl, posLoc, colorLoc, gridPosVBO, gridColorVBO, gl.Get("LINES"), numGridVertices)
		drawObject(gl, posLoc, colorLoc, axisPosVBO, axisColorVBO, gl.Get("LINES"), numAxisVertices)

		gl.Call("useProgram", pointProgram)
		gl.Call("uniformMatrix4fv", pointMvpLoc, false, sliceToJsFloat32Array(mvpMatrix[:]))
		gl.Call("enableVertexAttribArray", posLoc)
		gl.Call("enableVertexAttribArray", colorLoc)
		drawObject(gl, posLoc, colorLoc, redPosVBO, redColorVBO, gl.Get("POINTS"), numPoints)
		drawObject(gl, posLoc, colorLoc, greenPosVBO, greenColorVBO, gl.Get("POINTS"), numPoints)
		drawObject(gl, posLoc, colorLoc, bluePosVBO, blueColorVBO, gl.Get("POINTS"), numPoints)

		js.Global().Call("requestAnimationFrame", renderFrame)
		return nil
	})
	js.Global().Call("requestAnimationFrame", renderFrame)
}

func setupPointShaders(gl js.Value) (program, mvpLoc, posLoc, colorLoc js.Value, err error) {
	pointSize := 2.0
	vertShader := `attribute vec4 aPosition; attribute vec4 aColor; uniform mat4 uMvpMatrix; varying vec4 vColor; void main() { gl_Position = uMvpMatrix * aPosition; gl_PointSize = ` + fmt.Sprintf("%.1f", pointSize) + `; vColor = aColor; }`
	fragShader := `precision mediump float; varying vec4 vColor; void main() { gl_FragColor = vColor; }`

	program, err = createShaderProgram(gl, vertShader, fragShader)
	if err != nil {
		return js.Null(), js.Null(), js.Null(), js.Null(), err
	}

	posLoc = gl.Call("getAttribLocation", program, "aPosition")
	colorLoc = gl.Call("getAttribLocation", program, "aColor")
	mvpLoc = gl.Call("getUniformLocation", program, "uMvpMatrix")
	return
}

func setupLineShaders(gl js.Value) (program, mvpLoc js.Value, err error) {
	vertShader := `attribute vec4 aPosition; attribute vec4 aColor; uniform mat4 uMvpMatrix; varying vec4 vColor; void main() { gl_Position = uMvpMatrix * aPosition; vColor = aColor; }`
	fragShader := `precision mediump float; varying vec4 vColor; void main() { gl_FragColor = vColor; }`

	program, err = createShaderProgram(gl, vertShader, fragShader)
	if err != nil {
		return js.Null(), js.Null(), err
	}

	mvpLoc = gl.Call("getUniformLocation", program, "uMvpMatrix")
	return
}