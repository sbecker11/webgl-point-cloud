// wasm/event_handlers.go
package main

import (
	"syscall/js"
)

func setupEventHandlers(canvas, gl js.Value, camera *Camera) {
	canvas.Call("addEventListener", "mousedown", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		camera.HandleMouseDown(args[0].Get("clientX").Float(), args[0].Get("clientY").Float())
		return nil
	}))

	canvas.Call("addEventListener", "mousemove", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if camera.isMouseDown {
			camera.HandleMouseMove(args[0].Get("clientX").Float(), args[0].Get("clientY").Float())
		}
		return nil
	}))

	mouseUpOrLeave := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		camera.HandleMouseUp()
		return nil
	})
	canvas.Call("addEventListener", "mouseup", mouseUpOrLeave)
	canvas.Call("addEventListener", "mouseleave", mouseUpOrLeave)

	canvas.Call("addEventListener", "wheel", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		args[0].Call("preventDefault")
		camera.HandleMouseWheel(args[0].Get("deltaY").Float())
		return nil
	}), js.ValueOf(map[string]interface{}{"passive": false}))

	resizeFunc := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		width, height := js.Global().Get("innerWidth").Float(), js.Global().Get("innerHeight").Float()
		canvas.Set("width", width)
		canvas.Set("height", height)
		gl.Call("viewport", 0, 0, width, height)
		return nil
	})
	js.Global().Call("addEventListener", "resize", resizeFunc)
	resizeFunc.Call("call", js.Null()) // Initial call to set size
} 