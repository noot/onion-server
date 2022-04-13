package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/cretz/bine/tor"
	"github.com/cretz/bine/torutil/ed25519"
	logging "github.com/ipfs/go-log"
	"github.com/urfave/cli/v2"
)

const (
	defaultPrivateKeyFile = "service.key"
)

var log = logging.Logger("cmd")

var app = &cli.App{
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "datadir",
			Usage: "data directory used by program. if empty, uses data-dir-*.",
		},
		&cli.StringFlag{
			Name:  "private-key",
			Usage: "path to private key file. if not set, generates a new private key and writes it to service.key",
		},
		&cli.StringFlag{
			Name:  "log",
			Usage: "logging level. one of crit|error|warn|info|debug",
		},
		&cli.StringFlag{
			Name:  "serve-dir",
			Usage: "path to static website to serve",
		},
	},
	Action: run,
}

func getPrivateKey(c *cli.Context) (ed25519.PrivateKey, error) {
	pkFile := c.String("private-key")
	if pkFile != "" {
		log.Debugf("reading private key from file %s", pkFile)
		pkHexBytes, err := ioutil.ReadFile(filepath.Clean(pkFile))
		if err != nil {
			return nil, err
		}

		pkStr := string(pkHexBytes)
		pkBytes, err := hex.DecodeString(pkStr)
		if err != nil {
			return nil, err
		}

		if len(pkBytes) != ed25519.PrivateKeySize {
			return nil, fmt.Errorf("invalid private key size: got %d, expected %d", len(pkBytes), ed25519.PrivateKeySize)
		}

		return ed25519.PrivateKey(pkBytes), nil
	}

	log.Debugf("generating new private key and writing to file %s", defaultPrivateKeyFile)
	pk, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	pkStr := hex.EncodeToString(pk.PrivateKey())
	err = ioutil.WriteFile(defaultPrivateKeyFile, []byte(pkStr), os.ModePerm)
	if err != nil {
		return nil, err
	}

	return pk.PrivateKey(), nil
}

func setLogLevel(c *cli.Context) error {
	const (
		levelError = "error"
		levelWarn  = "warn"
		levelInfo  = "info"
		levelDebug = "debug"
	)

	level := c.String("log")
	if level == "" {
		level = levelInfo
	}

	switch level {
	case levelError, levelWarn, levelInfo, levelDebug:
	default:
		return fmt.Errorf("invalid log level")
	}

	_ = logging.SetLogLevel("cmd", level)
	return nil
}

func run(c *cli.Context) error {
	err := setLogLevel(c)
	if err != nil {
		return err
	}

	serveDir := c.String("serve-dir")
	if serveDir == "" {
		return fmt.Errorf("must provide --serve-dir (static website to serve)")
	}

	log.Info("Starting and registering onion service, please wait...")

	startConf := &tor.StartConf{
		NoAutoSocksPort: true,
		DataDir:         c.String("datadir"),
	}
	if c.String("log") == "debug" {
		// if debug is enabled, write all logs to stdout
		startConf.DebugWriter = os.Stdout
	}

	t, err := tor.Start(context.Background(), startConf)
	if err != nil {
		return fmt.Errorf("failed to start tor: %w", err)
	}

	defer t.Close()

	// Wait at most a few minutes to publish the service
	listenCtx, listenCancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer listenCancel()

	pk, err := getPrivateKey(c)
	if err != nil {
		return fmt.Errorf("failed to get private key: %w", err)
	}

	// Create a v3 onion service to listen on any port but show as 80
	onion, err := t.Listen(listenCtx, &tor.ListenConf{
		Version3:    true,
		RemotePorts: []int{80},
		Key:         pk.KeyPair(),
	})
	if err != nil {
		return fmt.Errorf("unable to create onion service: %w", err)
	}
	defer onion.Close()

	log.Infof("Open Tor browser and navigate to http://%v.onion\n", onion.ID)
	log.Infof("Press enter to exit")

	// Serve the current folder from HTTP
	errCh := make(chan error, 1)
	go func() {
		errCh <- http.Serve(onion, NewHandler(onion, serveDir))
	}()

	// End when enter is pressed
	go func() {
		fmt.Scanln()
		errCh <- nil
	}()

	if err = <-errCh; err != nil {
		return fmt.Errorf("failed serving: %w", err)
	}

	return nil
}

func main() {
	err := app.Run(os.Args)
	if err != nil {
		panic(err)
	}
}

var _ http.Handler = &Handler{}

type Handler struct {
	onion    *tor.OnionService
	serveDir string
}

func NewHandler(onion *tor.OnionService, serveDir string) *Handler {
	return &Handler{
		onion:    onion,
		serveDir: serveDir,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method is not supported.", http.StatusNotFound)
		return
	}

	err := http.Serve(h.onion, http.FileServer(http.Dir(h.serveDir)))
	if err != nil {
		log.Errorf("failed to serve: %w", err)
	}
}
