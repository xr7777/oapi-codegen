// CLI contract-test client for the running stdhttp petstore-expanded server.
// Start the server, then run:
//
//	go run ./cmd/specmatic --port 8080
package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

const specmaticImage = "specmatic/specmatic:latest"

func main() {
	port := flag.String("port", "8080", "Port of the running server")
	flag.Parse()

	docker, err := exec.LookPath("docker")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Docker is required to run Specmatic")
		os.Exit(1)
	}
	_, sourceFile, _, _ := runtime.Caller(0)
	exampleDirectory := filepath.Clean(filepath.Join(filepath.Dir(sourceFile), "..", "..", ".."))
	cmd := exec.Command(docker,
		"run", "--rm",
		"--add-host", "host.docker.internal:host-gateway",
		"--volume", exampleDirectory+":/workspace",
		"--workdir", "/workspace/stdhttp",
		"--env", "PETSTORE_BASE_URL=http://host.docker.internal:"+*port,
		"--env", "PETSTORE_SWAGGER_URL=http://host.docker.internal:"+*port+"/openapi.json",
		specmaticImage,
		"test", "--config", "specmatic.yaml",
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Specmatic contract tests failed: %v\n", err)
		os.Exit(1)
	}
}
