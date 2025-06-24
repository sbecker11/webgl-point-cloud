### To make this runnable:
Project name: github.com/sbecker11/webgl-point-cloud

Project Structure:

webgl-point-cloud/
├── main.go               <-- Go HTTP server
├── go.mod                <-- Go module file (for both server and glf32 package)
├── go.sum
├── glf32/                <-- New linear algebra package
│   └── glf32.go
└── wasm/                 <-- Directory for WebAssembly related files
    ├── wasm_main.go      <-- Your WebGL application source
    ├── index.html        <-- HTML page to load the WASM app
    └── wasm_exec.js      <-- Go's WASM glue code (copied here)
    └── main.wasm         <-- Compiled WebGL application (output of wasm_main.go)

Compilation and Execution Steps:

Initialize Go Module (if not already done):
If this is a new project, in your-project-name/, run:

Bash

go mod init github.com/sbecker11/webgl-point-cloud

Update wasm/wasm_main.go Import Path:
Make sure the import for glf32 in wasm/wasm_main.go is correct relative to your go.mod file. If your go.mod is module your-project-name, then the import in wasm_main.go should be:

Go

import webgl-point-cloud/glf32
Copy wasm_exec.js:
Make sure wasm_exec.js is in your wasm/ directory.

Bash

cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" webgl-point-cloud/wasm/
Tidy Go Modules:
In your-project-name/, run:

Bash

go mod tidy
This ensures glf32 is correctly recognized by both main.go and wasm/wasm_main.go.

Compile the WebAssembly Application:
Navigate to the wasm/ directory:

Bash

cd your-project-name/wasm/
Then compile:

Bash

GOOS=js GOARCH=wasm go build -o main.wasm wasm_main.go
This will create main.wasm inside your-project-name/wasm/.

Compile the HTTP Server:
Navigate back to the project root:

Bash

cd .. # Or cd your-project-name/ if you're not there
Then compile the server:

Bash

go build -o server main.go
This creates an executable server (or server.exe on Windows) in your-project-name/.

Run the Server:

Bash

./server
You'll see Server running at http://localhost:8080.

View in Browser:
Open your web browser and go to http://localhost:8080/wasm/index.html.


1.  **Save the `glf32` package:**
    Create a directory named `glf32` inside your project root.
    Save the code provided previously (from "glf32/glf32.go") into `your-project/glf32/glf32.go`.

2.  **Update `go.mod`:**
    If you don't have a `go.mod` file, create one in your project root:
    `go mod init your-project-name` (replace `your-project-name` with your actual project name, e.g., `github.com/yourusername/webgl-sphere-project`)

    Then, in `main.go`, change the import `glf32` to reflect your module path:
    `import "your-project-name/glf32"` (e.g., `import "github.com/yourusername/webgl-sphere-project/glf32"`)

    Run `go mod tidy` in your project root to ensure dependencies are resolved.

3.  **HTML Setup:** You'll need an `index.html` file to load the WebAssembly.
    ```html
    <!DOCTYPE html>
    <html>
    <head>
        <meta charset="utf-8">
        <title>WebGL Point Cloud Viewer</title>
        <style>
            body { margin: 0; overflow: hidden; background-color: #000; }
            canvas { display: block; width: 100vw; height: 100vh; }
        </style>
        <script src="wasm_exec.js"></script>
        <script>
            // Ensure WebAssembly is supported
            if (!WebAssembly.instantiateStreaming) { 
                WebAssembly.instantiateStreaming = async (resp, importObject) => {
                    const source = await (await resp).arrayBuffer();
                    return await WebAssembly.instantiate(source, importObject);
                };
            }

            const go = new Go();
            async function runWASM() {
                try {
                    const result = await WebAssembly.instantiateStreaming(fetch("main.wasm"), go.importObject);
                    go.run(result.instance);
                } catch (err) {
                    console.error(err);
                    alert("Error loading WebAssembly: " + err.message);
                }
            }
            runWASM();
        </script>
    </head>
    <body>
        <canvas id="canvas"></canvas>
    </body>
    </html>
    ```

4.  **`wasm_exec.js`:** Copy this file from your Go installation:
    `cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" .`

5.  **Build the WASM:**
    `GOOS=js GOARCH=wasm go build -o main.wasm .`

6.  **Serve:**
    You'll need a simple HTTP server to serve the files (due to WASM security restrictions). You can use Python:
    `python -m http.server 8080`
    Then open your browser to `http://localhost:8080`.

This setup fully integrates your `glf32` package, uses the correct column-major matrix logic, and renders your animated sphere and axes in WebGL.
