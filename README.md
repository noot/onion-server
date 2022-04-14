# hidden service server

A CLI that will host a static website as a .onion hidden service.

Comes with an additional binary that can be used to generate vanity .onion addresses.

## Requirements

- go1.17
- tor 0.4.x
	- download source here https://www.torproject.org/download/tor/
	- extract files (`tar -xzf`) and navigate to directory 
	- `./configure && make && sudo make install`. Check that tor is installed with `tor --version`.

## Usage

### Build

```
make build
```

This places the binaries `onioncli` and `onionaddress` in the project root.

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
./onioncli --private-key=service.key --serve-dir ~/my-website
```

You can also turn on debug logs with `--log=debug`.

#### Vanity addresses

To find a vanity address and its private key:
```bash
./onionaddress --prefix <some-prefix> --count=3
```

This will serach for and print 3 .onion addresses with the given prefix and their corresponding private keys. The private keys can be used with `onioncli --private-key=<keyfile>`.

Note: for 4-letter prefixes and less, this process is quite quick. For 5-letter prefixes, it took around ~30 minutes on my machine to find 1 address, and this grows exponentially the longer the prefix gets.