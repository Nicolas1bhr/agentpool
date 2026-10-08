// Command agentpool is the AgentPool inference service: one stable model provider for
// OpenCode, backed by a pool of independently authenticated routes.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Nicolas1bhr/agentpool/internal/gateway"
	"github.com/Nicolas1bhr/agentpool/internal/version"
)

const usage = `agentpool — one model provider in front of a pool of AI accounts

Usage:
  agentpool serve [-addr 127.0.0.1:43819]   run the inference data plane (skeleton)
  agentpool version                         print build identity
  agentpool help                            show this help

Status: skeleton. Inference requests return 501 until milestone P1 lands.
`

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(stderr, usage)
		return 2
	}
	switch args[0] {
	case "serve":
		return serve(args[1:], stderr)
	case "version", "--version", "-v":
		fmt.Fprintln(stdout, version.String())
		return 0
	case "help", "--help", "-h":
		fmt.Fprint(stdout, usage)
		return 0
	default:
		fmt.Fprintf(stderr, "agentpool: unknown command %q\n\n%s", args[0], usage)
		return 2
	}
}

func serve(args []string, stderr io.Writer) int {
	fs := flag.NewFlagSet("serve", flag.ContinueOnError)
	fs.SetOutput(stderr)
	addr := fs.String("addr", gateway.DefaultAddr, "loopback address for the inference data plane")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	log := slog.New(slog.NewTextHandler(stderr, nil))
	srv, err := gateway.NewServer(*addr)
	if err != nil {
		log.Error("refusing to start", "err", err)
		return 1
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errc := make(chan error, 1)
	go func() { errc <- srv.ListenAndServe() }()
	log.Info("data plane listening", "addr", *addr, "version", version.Version, "stage", "skeleton")

	select {
	case err := <-errc:
		if !errors.Is(err, http.ErrServerClosed) {
			log.Error("server failed", "err", err)
			return 1
		}
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Error("shutdown", "err", err)
			return 1
		}
		log.Info("stopped")
	}
	return 0
}
