# A binary for precompile benchmark

## Installation instructions

Sync the repo, check out the `runtime-design` branch and build the executable:

```
$ go build ./cmd/ethereum/...
```

You should then have an executable called `ethereum` in the same directory.

## Usage

```
$ ./ethereum -input "input data" <wasm file>
result: <whatever finish/revert was called with>
```
