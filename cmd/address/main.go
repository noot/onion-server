package main

import (
	"encoding/hex"
	"fmt"
	"os"
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

	start := time.Now()

	var found uint
	for i := uint64(0); i < max; i++ {
		addr, priv, err := GenerateAddress()
		if err != nil {
			return err
		}

		if addr[:len(prefix)] != prefix {
			continue
		}

		fmt.Fprintf(os.Stdout, "%s.onion\t%s\n", addr, hex.EncodeToString(priv))
		found++
		if found == count && count != 0 {
			break
		}
	}

	duration := time.Since(start)
	fmt.Printf("duration: %dms\n", duration.Milliseconds())
	return nil
}
