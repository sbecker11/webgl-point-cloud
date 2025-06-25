# Project: WebGL Point Cloud
[github.com/sbecker11/webgl-point-cloud](https://github.com/sbecker11/webgl-point-cloud)

[![Sphere Demo](./media/Sphere86KB.png)](./media/Sphere1.4MB.mp4)

*Click the image above to see a video of the demo.*

This project is a WebAssembly-based application for visualizing 3D data, written in Go. It renders a point cloud sphere with interactive controls and serves as a foundation for more advanced data visualization tasks.

## Features
- **Interactive 3D View**: Click and drag to rotate the scene. A damping effect provides smooth deceleration.
- **Go + WebAssembly**: The core rendering logic is written in Go and compiled to WebAssembly, running directly in the browser.
- **Custom Math Package**: Includes a `glf32` package for 3D graphics-focused linear algebra (vector and matrix operations).
- **Responsive Design**: The main `index.html` page and the WebGL canvas are responsive and support system-level dark mode.

## Project Structure:
```bash
webgl-point-cloud/
├── styles.css            <-- NEW: Styles for the main page
├── index.html            <-- Main project page
├── main.go               <-- Go HTTP server
├── go.mod                <-- Go module file (for both server and glf32 package)
├── go.sum
├── glf32/                <-- Custom linear algebra package
│   ├── glf32.go
│   ├── glf32_test.go
│   ├── glf32_wasm.go
│   └── README.md
└── wasm/                 <-- Directory for WebAssembly related files
    ├── wasm_main.go      <-- WebGL application source
    ├── index.html        <-- HTML page to load the WASM app
    └── wasm_exec.js      <-- Go's WASM glue code (copied here)
    └── main.wasm         <-- Compiled WebGL application (output of wasm_main.go)
```

## Compilation and Execution Steps:  
Initialize Go Module (if not already done):
If this is a new project, in your-project-name/, run:

```bash
go mod init github.com/sbecker11/webgl-point-cloud
```

## Update wasm/wasm_main.go Import Path:  
Make sure the import for glf32 in wasm/wasm_main.go is correct relative to your go.mod file. If your go.mod is module your-project-name, then the import in wasm_main.go should be:
```go
import "github.com/sbecker11/webgl-point-cloud/glf32"
```
## Copy wasm_exec.js:  
Copy the installed `wasm_exec.js` from `~/wasm/` directory.
```bash
cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" webgl-point-cloud/wasm/
```
If you can't find `wasm_exec.jg` under `$GOROOT` then get it from using: 
```bash
curl -o wasm/wasm_exec.js https://raw.githubusercontent.com/golang/go/master/misc/wasm/wasm_exec.js
```
## Tidy Your Go Modules:  
In your project root, run:   
```bash
go mod tidy
```
This ensures that the local glf32 package is correctly recognized by both `main.go` and `wasm/wasm_main.go`.

## Compile the WebAssembly Application:  
Navigate to the wasm/ directory under your project-root:  
```bash
cd webgl-point-cloud/wasm/
```
Then compile:  
```bash
GOOS=js GOARCH=wasm go build -o main.wasm wasm_main.go  
```
This will create `main.wasm` in the same folder `webgl-point-cloud/wasm/`.

## Compile the HTTP Server:  
Navigate back to the project root:
```bash
cd webgl-point-cloud/
```
Then compile the server:  
```bash
go build -o server main.go
```
This creates an executable `server` (or `server.exe` on Windows) in your project root.

## Run the Server:  
```Bash
./server
```
You'll see Server running at `http://localhost:8080`.

## View in Browser:  
Open your web browser and go to [http://localhost:8080/wasm/index.html](http://localhost:8080/wasm/index.html).

Click and drag the mouse on the canvas to rotate the scene.

## Notes:  
1.  **Save the `glf32` package:**
    Create a directory named `glf32` inside your project root.
    Save the code provided previously (from "glf32/glf32.go") into `your-project/glf32/glf32.go`.

2.  **Update `go.mod`:**
    If you don't have a `go.mod` file, create one in your project root:
    `go mod init your-project-name` (replace `your-project-name` with your actual project name, e.g., `github.com/yourusername/webgl-sphere-project`)

    Then, in `main.go`, change the import `glf32` to reflect your module path:
    `import "your-project-name/glf32"` (e.g., `import "github.com/yourusername/webgl-sphere-project/glf32"`)

    Run `go mod tidy` in your project root to ensure dependencies are resolved.

3.  **HTML Setup:** Ensure you have a valid `wasm/index.html` file to load your WebAssembly application. The server is configured to serve the entire `wasm` directory.

4.  **`wasm_exec.js`:** Copy this file from your Go installation into the `wasm/` directory:
    `cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" ./wasm/`

5.  **Build the WASM:**
    `GOOS=js GOARCH=wasm go build -o wasm/main.wasm wasm/wasm_main.go`

6.  **Serve:**
    You'll need a simple HTTP server to serve the files. You can use the one provided (`server.go`) or a standard one like Python's:
    `python -m http.server 8080`
    Then open your browser to `http://localhost:8080/wasm/`.

This setup fully integrates your `glf32` package, uses the correct column-major matrix logic, and renders your animated sphere and axes in WebGL.
