# Go Module Clone - Flow Sequence Diagram

## Main Workflow Sequence

```mermaid
sequenceDiagram
    participant User
    participant CLI as CLI Handler
    participant Logger as Logger
    participant FSys as File System
    participant Parser as Module Parser
    participant Resolver as Dependency Resolver
    participant GoTools as Go Tools<br/>(go get, go list)
    participant ModCache as Module Cache
    participant Packer as Module Packer
    participant Worker as Worker Pool
    participant Storage as Athens Storage

    User->>CLI: Execute go-mod-clone<br/>--modules modules.txt<br/>--storage-root /data

    CLI->>Logger: Initialize logger<br/>(set log level)
    Logger-->>CLI: Ready

    CLI->>FSys: Validate/Create<br/>storage root
    FSys-->>CLI: Storage ready

    CLI->>FSys: Create/Validate<br/>work directory
    FSys-->>CLI: Work dir ready

    CLI->>FSys: Read modules.txt file
    FSys-->>CLI: File content

    CLI->>Parser: Parse module list
    Parser->>Parser: Parse each line<br/>Extract path@version
    Parser-->>CLI: []ModuleSpec
    Logger->>Logger: Log: Loaded N modules

    CLI->>Resolver: NewResolver(workDir)
    Resolver-->>CLI: Resolver instance

    CLI->>Resolver: ResolveDependencies(specs)

    loop For each module (BFS queue)
        Resolver->>FSys: Create temp dir<br/>resolve-temp-{id}
        FSys-->>Resolver: Temp dir ready

        Resolver->>FSys: Write go.mod<br/>(module temp)
        FSys-->>Resolver: go.mod created

        Resolver->>GoTools: go get -d {module}@{version}
        GoTools->>ModCache: Download module
        ModCache-->>GoTools: Downloaded
        GoTools-->>Resolver: Success

        Resolver->>GoTools: go mod download all
        GoTools->>ModCache: Download all dependencies
        ModCache-->>GoTools: All downloaded
        GoTools-->>Resolver: Success

        Resolver->>GoTools: go mod download -json all
        GoTools-->>Resolver: JSON output<br/>(paths, versions)

        Resolver->>GoTools: go list -m -json all
        GoTools-->>Resolver: JSON output<br/>(modules with versions)

        Resolver->>Resolver: Parse JSON output<br/>Extract module info<br/>Filter valid semver

        Resolver->>Resolver: Build module map<br/>Path@Version → Module

        Resolver->>Resolver: Queue new dependencies<br/>Add to toProcess queue<br/>Mark as processed
    end

    Resolver-->>CLI: []Module (all resolved)
    Logger->>Logger: Log: Resolved N modules

    CLI->>Packer: NewPacker(storageRoot)
    Packer-->>CLI: Packer instance

    CLI->>Worker: NewPool(concurrency)
    Worker-->>CLI: Worker pool ready

    par Parallel Packing
        loop For each module
            CLI->>Worker: Submit pack job

            par Worker processes modules
                Worker->>Packer: Pack(module)
                Packer->>FSys: Check if source.zip exists
                alt Already packed
                    Packer-->>Worker: Skip (idempotent)
                else First time
                    Packer->>ModCache: Read module directory
                    ModCache-->>Packer: Module files

                    Packer->>Storage: Create Athens path<br/>github.com/user/pkg/v1.0.0/
                    Storage-->>Packer: Path ready

                    Packer->>Storage: Copy go.mod file
                    Storage-->>Packer: Copied

                    Packer->>Storage: Create {version}.info<br/>(JSON metadata)
                    Storage-->>Packer: Created

                    Packer->>FSys: Zip module directory<br/>(exclude .git, cache)
                    FSys-->>Packer: source.zip

                    Packer->>Storage: Save source.zip
                    Storage-->>Packer: Saved

                    Packer-->>Worker: Success/Failure
                end
            end
        end
    end

    Worker->>Worker: Wait for all jobs
    Worker-->>CLI: Complete

    CLI->>Logger: Print summary<br/>Total, Packed, Failed

    alt All success
        CLI-->>User: Exit 0 (success)
    else Some failures
        Logger->>Logger: Log failed modules
        CLI-->>User: Exit 1 (failure)
    end
```

## Server Mode Workflow

```mermaid
sequenceDiagram
    participant User
    participant CLI as CLI Handler
    participant Server as HTTP Server
    participant Storage as Athens Storage
    participant FileServ as File Server

    User->>CLI: Execute go-mod-clone server<br/>--storage-root /data<br/>--port 3000

    CLI->>Server: NewServer(root, host, port)
    Server-->>CLI: Server instance

    CLI->>Server: Start()
    Server->>FileServ: Create file server<br/>for storage root
    FileServ-->>Server: Ready

    Server->>Server: Listen on host:port

    loop Incoming requests
        User->>Server: GET /github.com/user/pkg/@v/v1.0.0.mod
        Server->>Storage: Serve file<br/>github.com/user/pkg/v1.0.0/go.mod
        Storage-->>Server: File content
        Server-->>User: HTTP 200<br/>go.mod content

        User->>Server: GET /github.com/user/pkg/@v/v1.0.0.zip
        Server->>Storage: Serve file<br/>github.com/user/pkg/v1.0.0/source.zip
        Storage-->>Server: File content
        Server-->>User: HTTP 200<br/>source.zip
    end
```

## Dependency Resolution Detail (Recursive BFS)

```mermaid
sequenceDiagram
    participant Input as Input Specs
    participant Queue as Processing Queue
    participant Processed as Processed Map
    participant Resolver as resolveModule()
    participant GoMod as go list/get
    participant Result as Resolved Modules

    Input->>Queue: Add initial specs

    loop While queue not empty
        Queue->>Processed: Check if module processed
        alt Already processed
            Queue-->>Queue: Skip
        else New module
            Queue->>Resolver: resolveModule(spec)

            Resolver->>GoMod: Download & resolve module
            GoMod->>GoMod: go get, go list all deps
            GoMod-->>Resolver: Dependencies found

            Resolver->>Result: Add module to result map
            Result-->>Resolver: Added

            loop For each dependency
                Resolver->>Processed: Check if processed
                alt Not yet processed
                    Resolver->>Queue: Queue dependency
                    Processed->>Processed: Mark as processed
                else Already processed
                    Resolver-->>Resolver: Skip (cycle detection)
                end
            end

            Resolver-->>Queue: Complete
        end
    end

    Queue-->>Result: All modules resolved
```

## Error Handling Flow

```mermaid
sequenceDiagram
    participant CLI as go-mod-clone
    participant Operation as Operation
    participant Logger as Logger
    participant Recovery as Recovery

    CLI->>Operation: Execute operation

    alt Success
        Operation-->>CLI: Result
        CLI->>Logger: Log success
    else Error occurs
        Operation->>Operation: Catch error
        Operation->>Logger: Log error

        alt Critical error
            Logger-->>CLI: Error returned
            CLI->>Recovery: Cleanup<br/>(remove temp dir)
            Recovery-->>CLI: Cleanup done
            CLI-->>CLI: Exit with code 1
        else Non-critical error<br/>(e.g., pack failure)
            Logger->>Logger: Log failure details
            Operation-->>CLI: Continue with other items
            CLI->>Logger: Track failure count
            Logger-->>CLI: Include in summary
        end
    end
```

---

## Process Summary

### Main Command Flow
1. **Initialization**: Setup logger, storage, work directory
2. **Parsing**: Read and parse modules.txt
3. **Dependency Resolution**: Use go tools to resolve all transitive dependencies
4. **Parallel Packing**: Pack modules into Athens format using worker pool
5. **Summary**: Report results with success/failure counts

### Key Features
- **Recursive Resolution**: BFS approach finds all transitive dependencies
- **Parallel Processing**: Worker pool speeds up packing (default 4 workers)
- **Idempotent**: Skips already-packaged modules
- **Error Resilience**: Continues on individual pack failures, reports summary
- **Server Mode**: Serves packaged modules as HTTP Go module proxy
- **Concurrency**: Thread-safe with mutexes protecting shared state
