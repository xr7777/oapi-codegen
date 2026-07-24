# Petstore Expanded Example

This example demonstrates [oapi-codegen](https://github.com/oapi-codegen/oapi-codegen) generating server stubs for several different Go HTTP frameworks from a single [OpenAPI 3.0 spec](petstore-expanded.yaml).

Each server variant is fully self-contained, in that it has its own generated types, its own
copy of the trivial in-memory database (a Go map), and its own handler implementation.
The backend is simple enough that copying it into each variant keeps every example 
readable on its own. A single shared test client (`common/client`) can exercise any of the variants over real HTTP.

## Directory Structure

```
petstore-expanded/
├── petstore-expanded.yaml          # Shared OpenAPI spec
├── common/
│   └── client/                     # Single test client, works against any variant
│       ├── main.go                 # CLI test client
│       ├── testclient/
│       │   └── testclient.go       # Reusable CRUD test sequence
│       └── openapi/
│           ├── generate.go         # go:generate for the client
│           ├── client.cfg.yaml     # Codegen config: client + models
│           └── client.gen.go       # Generated HTTP client (+ model types)
├── chi/                            # Chi (net/http compatible)
├── gorilla/                        # Gorilla/mux (net/http compatible)
├── stdhttp/                        # stdlib net/http
├── echo/                           # Echo v4
├── echo-v5/                        # Echo v5
├── gin/                            # Gin
├── fiber/                          # Fiber
├── iris/                           # Iris
└── strict/                         # Strict server (Chi + typed request/response objects)
```

Each server variant follows the same self-contained pattern:
- `api/server.cfg.yaml` — codegen config generating the server interface, model types, and embedded spec
- `api/generate.go` — `//go:generate` directive for the server code
- `api/petstore-server.gen.go` — generated server boilerplate and model types
- `server/store.go` — the in-memory CRUD store (copied into each variant)
- `server/server.go` — hand-written `ServerInterface` implementation backed by the local store
- `server/setup.go` — factory function that creates a fully configured server/app
- `petstore.go` — `main()` wiring (thin wrapper around `setup.go`)

## Generating Code

From the repository root, `make generate` regenerates everything. To regenerate a
single piece from the `examples/` directory:

```sh
# Generate a specific server variant (server + models)
cd examples/petstore-expanded/chi/api && go generate ./...

# Generate the shared test client
cd examples/petstore-expanded/common/client/openapi && go generate ./...
```

## Running a Server

```sh
cd examples/petstore-expanded/chi
go run . --port 8080
```

Replace `chi` with any variant: `gorilla`, `stdhttp`, `echo`, `echo-v5`, `gin`, `fiber`, `iris`, `strict`.

## Test Client

A single client executable verifies the behavior of any variant over real HTTP.
Start a server in one terminal, then point the client at it from another:

```sh
# Terminal 1: start any server variant
cd examples/petstore-expanded/chi && go run . --port 8080

# Terminal 2: run the test client against it
cd examples && go run ./petstore-expanded/common/client/ --port 8080
```

The client verifies: add pets, find by ID, 404 on missing pet, list/filter by tag, delete, and empty list after deletion.

## Contract Testing with Specmatic

The `stdhttp` variant also has an opt-in [Specmatic](https://specmatic.io/) contract test. It starts the real server and checks its HTTP responses against [petstore-expanded.yaml](petstore-expanded.yaml), rather than testing a mocked handler.

This is useful because compilation and handler unit tests cannot by themselves confirm that a running server returns the status codes, headers, and JSON shapes promised by the OpenAPI document. Contract testing catches that drift at the HTTP boundary. During this integration, it exposed two real mismatches: Fiber v3 returned `201 Created` for `POST /pets` when the contract specifies `200 OK`, and the `stdhttp` store encoded an empty pet list as `null` instead of `[]`.

### Run the contract test

Docker must be running; the test uses the pinned `specmatic/specmatic:2.50.1` image.

```sh
# From the repository root
make specmatic-test
```

Or run it directly from the example:

```sh
cd examples/petstore-expanded/stdhttp
go test -tags=specmatic -count=1 -v ./...
```

The test automatically:

- starts the `stdhttp` server on a free port;
- runs Specmatic in Docker against the live server; and
- shuts the server down and writes an HTML report to `stdhttp/build/reports/specmatic/test/html/index.html`.

The contract test is behind the `specmatic` build tag. Normal generation, test, and lint commands do not require Docker or run Specmatic.

### Adapting this to other business logic

The same pattern works for a service with real business logic, a database, or external dependencies:

1. Point `specmatic.yaml` at that service's OpenAPI document.
2. Start the real service in a test harness, Docker Compose stack, or test environment.
3. Prepare only the deterministic state required for successful scenarios, using public APIs, fixtures, or migrations.
4. Run Specmatic against the live base URL in CI.

This keeps the OpenAPI document as the source of truth while testing the actual HTTP boundary, not an in-memory mock.
