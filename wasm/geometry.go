// wasm/geometry.go
package main

import (
	"math"
	"math/rand"
	"github.com/sbecker11/webgl-point-cloud/glf32"
)


// generateNormalCluster creates a cluster of points with a normal (Gaussian) distribution.
func generateNormalCluster(numPoints int, center glf32.Vec3, stdDev float32, color glf32.Vec3) (coords []float32, colors []float32) {
	coords = make([]float32, 0, numPoints*3)
	colors = make([]float32, 0, numPoints*4) // 4 components for RGBA

	for i := 0; i < numPoints; i++ {
		u1, u2 := rand.Float32(), rand.Float32()
		mag := stdDev * float32(math.Sqrt(-2.0*math.Log(float64(u1))))
		z0 := mag * float32(math.Cos(2.0*math.Pi*float64(u2)))
		z1 := mag * float32(math.Sin(2.0*math.Pi*float64(u2)))

		u3, u4 := rand.Float32(), rand.Float32()
		mag2 := stdDev * float32(math.Sqrt(-2.0*math.Log(float64(u3))))
		z2 := mag2 * float32(math.Cos(2.0*math.Pi*float64(u4)))

		coords = append(coords, center[0]+z0, center[1]+z1, center[2]+z2)
		colors = append(colors, color[0], color[1], color[2], 1.0) // Add alpha
	}
	return coords, colors
}

// --- Geometry Generation ---

func generateAxes(size float32) ([]float32, []float32) {
	vertices := []float32{
		// X-axis (red)
		-size, 0, 0, size, 0, 0,
		// Y-axis (green)
		0, -size, 0, 0, size, 0,
		// Z-axis (blue)
		0, 0, -size, 0, 0, size,
	}
	colors := []float32{
		// X-axis
		1, 0, 0, 1, 1, 0, 0, 1,
		// Y-axis
		0, 1, 0, 1, 0, 1, 0, 1,
		// Z-axis
		0, 0, 1, 1, 0, 0, 1, 1,
	}
	return vertices, colors
}

func generateGrid(size float32, divisions int) ([]float32, []float32) {
	var vertices []float32
	var colors []float32
	step := size / float32(divisions)
	gridColor := []float32{0.4, 0.4, 0.4, 1.0}

	for i := -divisions; i <= divisions; i++ {
		if i == 0 {
			continue // Don't draw over the axes
		}
		pos := float32(i) * step

		// Lines for the XZ plane (y=0)
		vertices = append(vertices, -size, 0, pos, size, 0, pos) // Parallel to X
		vertices = append(vertices, pos, 0, -size, pos, 0, size) // Parallel to Z
		
		// Lines for the XY plane (z=0)
		vertices = append(vertices, -size, pos, 0, size, pos, 0) // Parallel to X
		vertices = append(vertices, pos, -size, 0, pos, size, 0) // Parallel to Y
		
		// Lines for the YZ plane (x=0)
		vertices = append(vertices, 0, pos, -size, 0, pos, size) // Parallel to Z
		vertices = append(vertices, 0, -size, pos, 0, size, pos) // Parallel to Y

		// Add colors for the 6 lines (12 vertices) we just added
		for j := 0; j < 12; j++ {
			colors = append(colors, gridColor...)
		}
	}
	return vertices, colors
} 