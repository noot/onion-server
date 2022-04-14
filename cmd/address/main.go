package main

import (
	"encoding/hex"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/urfave/cli/v2"
)

var app = &cli.App{
	Name:  "onionaddress",
	Usage: "vanity onion address generator. for example, to find 5 addresses with the prefix \"fun\":\n\t$ ./onionaddress --prefix fun --count=5",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "prefix",
			Usage: "designated prefix to search for",
		},
		&cli.UintFlag{
			Name:  "max",
			Usage: "maximum number of iterations per goroutine; if --count is set, this is ignored. default=65536",
		},
		&cli.UintFlag{
			Name:  "grs",
			Usage: "number of goroutines to use for search. default=1",
		},
		&cli.UintFlag{
			Name:  "count",
			Usage: "how many addresses with the matching prefix to find. if set, ignores --max and runs until that many addresses are found",
		},
		&cli.BoolFlag{
			Name:  "no-prefix",
			Usage: "don't search for a specific prefix, but print all addresses and keys found",
		},
	},
	Action: run,
}

func main() {
	err := app.Run(os.Args)
	if err != nil {
		panic(err)
	}
}

func run(c *cli.Context) error {
	max := uint64(c.Uint("max"))
	if max == 0 {
		max = 65536
	}

	prefix := c.String("prefix")
	if len(prefix) == 0 && !c.Bool("no-prefix") {
		return fmt.Errorf("must provide --prefix; if no prefix is desired, use the --no-prefix option")
	}

	count := c.Uint("count")
	if count != 0 {
		max = ^uint64(0)
	}

	grs := int(c.Uint("grs"))
	if grs == 0 {
		grs = 1
	}

	start := time.Now()

	var found uint
	var wg sync.WaitGroup
	var mu sync.Mutex

	wg.Add(grs)

	for i := 0; i < grs; i++ {
		go func() {
			defer wg.Done()
			for j := uint64(0); j < max; j++ {
				if found >= count && count != 0 {
					break
				}

				addr, priv, err := GenerateAddress()
				if err != nil {
					continue
				}

				if addr[:len(prefix)] != prefix {
					continue
				}

				mu.Lock()
				fmt.Fprintf(os.Stdout, "%d %s.onion\t%s\n", found, addr, hex.EncodeToString(priv))
				found++
				mu.Unlock()
			}
		}()
	}

	wg.Wait()

	duration := time.Since(start)
	fmt.Printf("duration: %dms\n", duration.Milliseconds())
	return nil
}
