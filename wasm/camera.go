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
		rotationX:    0.3, // Start with a slight tilt
		rotationY:    -0.5, // Start with a slight rotation
		zoom:         1.0,
		velocityX:    0,
		velocityY:    0,
		damping:      0.90,
		isMouseDown:  false,
		minRotationX: -math.Pi / 2 * 0.999, // Clamp just before the poles
		maxRotationX: math.Pi / 2 * 0.999,
		minZoom:      0.1,
		maxZoom:      10.0,
	}
}

func (c *Camera) GetViewMatrix() glf32.Mat4 {
	// Calculate camera position using spherical coordinates.
	// This is the standard, stable way for an orbit camera.
	effectiveDistance := c.distance / c.zoom
	camX := effectiveDistance * float32(math.Sin(float64(c.rotationY))*math.Cos(float64(c.rotationX)))
	camY := effectiveDistance * float32(math.Sin(float64(c.rotationX)))
	camZ := effectiveDistance * float32(math.Cos(float64(c.rotationY))*math.Cos(float64(c.rotationX)))
	position := glf32.Vec3{camX, camY, camZ}

	// The world's up vector. Clamping rotationX prevents the camera's forward
	// vector from becoming parallel to 'up', which is what caused all crashes.
	up := glf32.Vec3{0, 1, 0}
	target := glf32.Vec3{0, 0, 0}

	// With the corrected LookAt function, this is now stable and reliable.
	return glf32.LookAt(position, target, up)
}

func (c *Camera) ApplyInertia() {
	if !c.isMouseDown && (c.velocityX != 0 || c.velocityY != 0) {
		c.rotationY += c.velocityX * 0.01
		c.rotationX += c.velocityY * 0.01
		c.wrapAngles()
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

func (c *Camera) wrapAngles() {
	// Keep rotationY between 0 and 2*PI to prevent floating point instability.
	c.rotationY = float32(math.Mod(float64(c.rotationY), 2*math.Pi))
	if c.rotationY < 0 {
		c.rotationY += 2 * math.Pi
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

	// Invert rotationY for intuitive horizontal rotation
	// Add to rotationX for intuitive vertical rotation
	c.rotationY -= float32(dx) * 0.01
	c.rotationX += float32(dy) * 0.01
	c.wrapAngles()
	c.clampRotation()

	// Update velocity for inertia, matching the rotation direction
	c.velocityX = -float32(dx) * 0.5
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