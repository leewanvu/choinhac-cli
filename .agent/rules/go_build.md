---
description: Rule for go build commands
---
When executing `go build`, ensure that the output binary is stored in the `build/` directory.
For example: instead of `go build -o player ./cmd/player`, use `go build -o build/player ./cmd/player`.
