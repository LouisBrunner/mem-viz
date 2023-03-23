// TODO: rebrand as mem-viz with dsc-viz as a separate tool

# `dsc-viz`

This tool allows to display the format of a dyld shared cache (DSC) file. It is also able to read the kernel-provided DSC in its memory.

## Installation

```sh
go install github.com/LouisBrunner/mem-viz@latest
```

## Usage

```
Usage of dsc-viz:
      --from-arch arm64                       scan your system to find the cache for the given architecture (e.g. arm64)
      --from-current-arch                     scan your system to find the cache for the current architecture
      --from-file ./dyld_shared_cache_arm64   file to fetch from, e.g. ./dyld_shared_cache_arm64
      --from-memory                           fetch from memory
      --logging-level string                  logrus log level for internal debugging, e.g. "debug" (default "error")
      --output string                         output format, one of: "graphviz", "latex", "markdown", "text", "ascii", "json" (default "text")
```

You can use `--from-memory` or `--from-current-arch` to let the tool fetch the DSC from your system (respectively from memory or from a file on disk). Otherwise you can use `--from-file` to specify a file to fetch from or `--from-arch` to scan your system but for a specific architecture.

### Output formats

A wide-range of output formats is supported.

#### `text` (default)

TODO: unimplemented

#### `ascii`

TODO: unimplemented

#### `graphviz`

TODO: unimplemented

#### `markdown`

TODO: unimplemented

#### `latex`

TODO: unimplemented

#### `json`

TODO:
