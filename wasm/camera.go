// wasm/camera.go
package main

import (
	"math"
	"github.com/sbecker11/webgl-point-cloud/glf32"
)

type Camera struct {
	distance         float32
	rotationX        float32
	rotationY        float32
	zoom             float32
	velocityX        float32
	velocityY        float32
	damping          float32
	isMouseDown      bool
	lastMouseX       float64
	lastMouseY       float64
	minRotationX     float32
	maxRotationX     float32
	minZoom          float32
	maxZoom          float32
}

func NewCamera(distance float32) *Camera {
	return &Camera{
		distance:     distance,
		rotationX:    0,
		rotationY:    0,
		zoom:         1.0,
		velocityX:    0,
		velocityY:    0,
		damping:      0.90,
		isMouseDown:  false,
		minRotationX: -math.Pi / 2 * 0.999,
		maxRotationX: math.Pi / 2 * 0.999,
		minZoom:      0.1,
		maxZoom:      10.0,
	}
}

func (c *Camera) GetViewMatrix() glf32.Mat4 {
	// Build the view matrix by inverting the camera's transformations.
	// 1. Start with an identity matrix.
	matrix := glf32.Identity()

	// 2. Translate out to the camera's distance along the Z-axis.
	matrix = glf32.MultiplyMatrices(matrix, glf32.Translate(0, 0, -c.distance/c.zoom))

	// 3. Rotate around the X-axis (pitch).
	matrix = glf32.MultiplyMatrices(matrix, glf32.RotateX(c.rotationX))

	// 4. Rotate around the Y-axis (yaw).
	matrix = glf32.MultiplyMatrices(matrix, glf32.RotateY(c.rotationY))

	return matrix
}

func (c *Camera) ApplyInertia() {
	if !c.isMouseDown && (c.velocityX != 0 || c.velocityY != 0) {
		c.rotationY += c.velocityX * 0.01
		c.rotationX += c.velocityY * 0.01 // Corrected direction for inertia
		c.velocityX *= c.damping
		c.velocityY *= c.damping
		c.clampRotation()
	}
}

func (c *Camera) clampRotation() {
	if c.rotationX > c.maxRotationX {
		c.rotationX = c.maxRotationX
	}
	if c.rotationX < c.minRotationX {
		c.rotationX = c.minRotationX
	}
}

func (c *Camera) HandleMouseDown(x, y float64) {
	c.isMouseDown = true
	c.lastMouseX = x
	c.lastMouseY = y
	c.velocityX = 0
	c.velocityY = 0
}

func (c *Camera) HandleMouseUp() {
	c.isMouseDown = false
}

func (c *Camera) HandleMouseMove(x, y float64) {
	if !c.isMouseDown {
		return
	}
	dx := x - c.lastMouseX
	dy := y - c.lastMouseY

	c.rotationY += float32(dx) * 0.01
	c.rotationX += float32(dy) * 0.01 // Corrected direction for mouse tilt
	c.clampRotation()

	// Update velocity for inertia
	c.velocityX = float32(dx) * 0.5
	c.velocityY = float32(dy) * 0.5

	c.lastMouseX = x
	c.lastMouseY = y
}

func (c *Camera) HandleMouseWheel(deltaY float64) {
	if deltaY < 0 {
		c.zoom *= 1.1
	} else {
		c.zoom /= 1.1
	}
	c.zoom = float32(math.Max(float64(c.minZoom), math.Min(float64(c.zoom), float64(c.maxZoom))))
} 