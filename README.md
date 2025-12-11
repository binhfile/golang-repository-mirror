# Athens Prefill Tool

A Go CLI tool to download Go modules and their dependencies, and package them in [Athens](https://github.com/gomods/athens) disk storage format for offline usage.

## 📝 Project Development Notes

This project demonstrates best practices for complex software development including:
- **Todo List Tracking**: See `TODO_TRACKING_GUIDE.md` for guidance on tracking progress on complex tasks
- **Todo Templates**: See `TODO_TEMPLATES.md` for ready-to-use templates for common task types
- **Comprehensive Testing**: Unit tests for critical functionality
- **Clean Architecture**: Well-organized code structure with clear separation of concerns

## Purpose

`athens-prefill` automates the process of:
1. Reading a list of Go modules from a configuration file
2. Resolving all dependencies recursively using `go list -m -json all`
3. Packaging modules in Athens disk storage format
4. Supporting concurrent processing for faster packaging
5. Being idempotent (skipping already-packaged modules)

This is useful for creating offline Go module caches that can be deployed to machines without internet access.

## Building

```bash
go build ./cmd/athens-prefill
```

This will create an executable named `athens-prefill` (or `athens-prefill.exe` on Windows).

## Usage

### Basic Example

```bash
export ATHENS_DISK_STORAGE_ROOT=/data/athens-storage
./athens-prefill --modules ./modules.txt
```

### With All Options

```bash
./athens-prefill \
  --modules ./modules.txt \
  --storage-root /data/athens-storage \
  --work-dir /tmp/athens-work \
  --concurrency 8 \
  --log-level debug
```

```
./athens-prefill.exe --modules ./modules.txt --storage-root C:\Users\ngoth\AppData\Local\Temp\athens-storage --log-level debug --concurrency 1
```

### Flags

- `--modules, -m` (required): Path to the modules list file
- `--storage-root, -s`: Athens disk storage root directory (default: `ATHENS_DISK_STORAGE_ROOT` env var)
- `--work-dir, -w`: Temporary work directory (default: system temp directory)
- `--concurrency, -j`: Number of concurrent workers (default: 4)
- `--log-level`: Logging level - `debug`, `info`, `warn`, `error` (default: `info`)

## modules.txt Format

The modules file should contain one Go module per line. Each line can be:

- `github.com/gin-gonic/gin@v1.9.1` (with specific version)
- `google.golang.org/grpc` (without version - latest will be resolved)

Lines starting with `#` are treated as comments and ignored. Empty lines are also ignored.

Example `modules.txt`:

```text
github.com/gin-gonic/gin@v1.9.1
github.com/sirupsen/logrus@v1.9.3
# This is a comment - latest version will be used
google.golang.org/grpc
```

## Output Structure

The tool creates an Athens disk storage structure:

```
<storage_root>/
  github.com/
    gin-gonic/
      gin/
        v1.9.1/
          go.mod
          v1.9.1.info
          source.zip
    sirupsen/
      logrus/
        v1.9.3/
          go.mod
          v1.9.3.info
          source.zip
```

Each module directory contains:
- `go.mod`: Module definition file
- `<version>.info`: JSON metadata with version and timestamp
- `source.zip`: Packaged module source code

## Offline Usage

To use the created cache offline:

1. Copy the `storage_root` directory to your offline machine
2. Configure Athens to use disk storage mode:

```yaml
# athens.yml
StorageType: disk
Storage:
  Disk:
    RootPath: /path/to/copied/storage_root
```

3. Run Athens proxy pointing to this storage

## Implementation Details

### Architecture

The project is organized as follows:

- `cmd/athens-prefill/main.go`: Entry point
- `internal/cli/`: CLI argument parsing using Cobra
- `internal/gomod/`: Go module parsing and validation
- `internal/resolver/`: Dependency resolution using `go list -m -json all`
- `internal/packer/`: Packaging modules into Athens format
- `internal/worker/`: Worker pool for concurrent processing
- `internal/log/`: Structured logging

### Dependency Resolution

The resolver:
1. Creates a temporary `go.mod` file
2. Adds all input modules as require statements
3. Runs `go mod tidy` to resolve all dependencies
4. Parses `go list -m -json all` output
5. Filters out main module and invalid versions

### Packing

For each module:
1. Checks if `source.zip` already exists (idempotent)
2. Copies `go.mod` from the module directory
3. Creates `<version>.info` with JSON metadata
4. Creates `source.zip` by zipping the module directory (excluding `.git` and cache files)

### Concurrency

Uses a worker pool pattern:
- Producer thread loads modules and submits pack jobs
- N worker goroutines process jobs from a shared queue
- Thread-safe logging and directory creation
- Graceful handling of concurrent operations

## Requirements

- Go 1.21 or later
- `go` command available in PATH
- Network access for dependency resolution (first run)

## Error Handling

- If module packing fails, the error is logged but doesn't stop the entire process
- Summary is printed at the end showing:
  - Total modules processed
  - Number successfully packed
  - Number failed (with details)

## Testing

Run the included tests:

```bash
go test ./...
```

Tests cover:
- Module list parsing
- Version validation
- File operations
- Path construction

## Troubleshooting

### "go list -m -json failed"
Ensure you have Go 1.21+ installed and `go` is in your PATH.

### "Permission denied" when creating storage directories
Ensure you have write permissions to the storage root directory.

### Module not found
Check that the module exists and is publicly available. The resolver requires network access for initial dependency resolution.

## Contributing

This is a personal project tool. Feel free to extend and customize as needed.

## License

MIT
