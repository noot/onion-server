# hidden service server

A CLI that will host a static website as a hidden service.

## Requirements

- go1.17
- tor 0.4.x
	- download source here https://www.torproject.org/download/tor/
	- extract files (`tar -xzf`) and navigate to directory 
	- `./configure && make && sudo make install`. Check that tor is installed with `tor --version`.

## Usage

### Build

```
cd cmd/ && go build -o onioncli && mv onioncli .. && cd ..
```

This places the binary `onioncli` in the project root.

### Run

Instead of building the project, you can also run it:
```
go run ./cmd/... [flags]
```

### Usage

To serve a static website:
```bash
./onioncli --serve-dir ~/my-website
$ 2022-04-13T10:44:44.217-0400	INFO	cmd	cmd/main.go:153	Open Tor browser and navigate to http://7ukuzklqxkwesfs3dla5zzj3bsjb6v2rx25bq3fr662qistclpixgxqd.onion
```

If you have run the CLI before and have a server private key already (by default stored in `service.key`), you can pass it to the CLI so that the .onion address used will be the same as before.

```bash
./onioncli --private-key=service.key --serve-dir ~/my-website/
```

You can also turn on debug logs with `--log=debug`.