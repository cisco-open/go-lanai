# Profiler Utilities

## Important Notes

- This package is designed for troubleshooting performance related issues such as
  - Memory Leak
  - High CPU Usage
- This package **MUST NOT** be used in production.

<br>

## How To Use

This package leverage GO `runtime/pprof` and `net/http/pprof` packages. 

For more efficient troubleshooting, this package should be used together with GO benchmarking techniques and suitable `pprof` visualization tools.

### Serve `pprof` Data via HTTP

To enable `pprof` over HTTP, add following code before application start, usually in `init()` of application's main package:

```go
package main
import "github.com/cisco-open/go-lanai/pkg/profiler"
func init() {
	profiler.Use()
}
```

The `profiler` package will enable `runtime/pprof` and install following endpoints under application's `server.context-path`:

- GET `/${server.contex-path}/debug/pprof/[pprof_profile]`: Dump compressed raw data as `gz`. 
  
  Common `[pprof_profile]` values are
  - `heap`
  - `block`
  - `goroutine`
  - `threadcreate`
  - `mutex`

- GET `/${server.contex-path}/debug/pprof/cmdline`: Returns running program's command line, with arguments separated by NUL bytes.
- GET `/${server.contex-path}/debug/pprof/profile`: Returns the pprof-formatted cpu profile.
- GET `/${server.contex-path}/debug/pprof/symbol`: Looks up the program counters listed in the request, responding with a table mapping program counters to function names
- GET `/${server.contex-path}/debug/pprof/trace`: Returns the execution trace in binary form

More details about GO `pprof` can be found at:
- [Profiling Go Programs](https://go.dev/blog/pprof) 
- [net/http/pprof](https://pkg.go.dev/net/http/pprof) 
- [runtime/pprof](https://pkg.go.dev/runtime/pprof)

### Visualize `pprof` Data

There are many tools available to visualize data produced by `pprof`. Here is an exmaple:

Tool: [Google pprof GO implementation](https://github.com/google/pprof)

Prerequisites:
1. Go development kit (recommended 1.16+)
2. [Graphviz](https://www.graphviz.org/). To install on MacOS:

   ```
   brew install graphviz
   ```

Install `google/pprof` CLI:
```shell
go install github.com/google/pprof@latest
```

Example Usage (Investigating AuthService's memory allocation):

```shell
pprof -http=localhost:6061 http://localhost:8900/auth/debug/pprof/heap
```

Command above would start a web service at `localhost:6061` and render various views of current heap snapshot.

Use `pprof --help` for detailed usage.


### Real-Time Performance Monitoring

This folder also includes a performance monitoring utility that serve an HTML page of real-time metrics charts. 
To enable it:

```go
package main
import "github.com/cisco-open/go-lanai/pkg/profiler/monitor"
func init() {
	monitor.Use()
}
```

The charts can be accessed via HTTP `http://server:port/<context-path>/debug/charts/`

**Note**: This is an experimental feature and may consume fare amount of resources.
