package main

import (
	"log/slog"
	"os"
	"runtime"
	"strings"

	"github.com/XPD-Kasun/filters"
)

var Version = filters.VERSION_0_9_0

func main() {
	args := os.Args[1:]

	// configure slog with configurable log level
	level := slog.LevelInfo
	for _, arg := range args {
		if arg == "-debug" || arg == "--debug" || arg == "-v" {
			level = slog.LevelDebug
			break
		}
	}
	if os.Getenv("DEBUG") != "" {
		level = slog.LevelDebug
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: level,
	}))
	slog.SetDefault(logger)

	slog.Info("starting go-filters generator", "version", Version)
	slog.Info("GO version", "version", runtime.Version())
	slog.Debug("command line arguments", "args", args)

	gofile := os.Getenv("GOFILE")
	goline := os.Getenv("GOLINE")
	gopkg := os.Getenv("GOPACKAGE")
	slog.Debug("environment variables", "GOFILE", gofile, "GOLINE", goline, "GOPACKAGE", gopkg)

	ctx := &GenCtx{
		line: goline,
		path: gofile,
		pkg:  gopkg,
	}

	var targetFileName string
	for _, arg := range args {
		if strings.HasPrefix(arg, "-") {
			continue
		}
		targetFileName = arg
		break
	}
	if targetFileName == "" {
		targetFileName = gofile + "_query.go"
	}

	slog.Info("generating filter queries", "targetFile", targetFileName)
	gen := NewGenerator(ctx)

	gen.Generate()
	slog.Debug("generated code preview", "codeLength", len(gen.GetCode()))
	gen.Sync(targetFileName)
}
