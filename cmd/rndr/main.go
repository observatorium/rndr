package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"github.com/efficientgo/tools/core/pkg/clilog"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/observatorium/rndr/pkg/version"
	"github.com/oklog/run"
	"github.com/openproto/protoconfig/go/kingpinv2"
	"github.com/pkg/errors"
	"gopkg.in/alecthomas/kingpin.v2"
)

const (
	logFormatLogfmt = "logfmt"
	logFormatJSON   = "json"
	logFormatCLILog = "clilog"
)

func setupLogger(logLevel, logFormat string) log.Logger {
	var lvl level.Option
	switch logLevel {
	case "error":
		lvl = level.AllowError()
	case "warn":
		lvl = level.AllowWarn()
	case "info":
		lvl = level.AllowInfo()
	case "debug":
		lvl = level.AllowDebug()
	default:
		panic("unexpected log level")
	}
	switch logFormat {
	case logFormatJSON:
		return level.NewFilter(log.NewJSONLogger(log.NewSyncWriter(os.Stderr)), lvl)
	case logFormatLogfmt:
		return level.NewFilter(log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr)), lvl)
	case logFormatCLILog:
		fallthrough
	default:
		return level.NewFilter(clilog.New(log.NewSyncWriter(os.Stderr)), lvl)
	}
}

type rndrConfig struct {
	logger log.Logger
	outDir string
	tmpl   *kingpinv2.PathOrContent
}

func main() {
	app := kingpin.New(filepath.Base(os.Args[0]), `Deployments Template Toolkit.`).Version(version.Version)
	logLevel := app.Flag("log.level", "Log filtering level.").
		Default("info").Enum("error", "warn", "info", "debug")
	logFormat := app.Flag("log.format", "Log format to use.").
		Default(logFormatCLILog).Enum(logFormatLogfmt, logFormatJSON, logFormatCLILog)

	rc := rndrConfig{}
	rc.tmpl = kingpinv2.Flag(app, "template", "Template definition YAML as defined in github.com/observatorium/rndr/pkg/rndr.TemplateDefinition").Required().PathOrContent()
	app.Flag("output", "Directory where to put output files").Short('o').Default("./rndr-out").ExistingDirVar(&rc.outDir)

	var g run.Group

	k8sCmd := registerKubernetesCommand(app, &g, rc)
	_ = registerKubernetesOperatorCommand(k8sCmd, &g, rc)
	_ = registerKubernetesHelmCommand(k8sCmd, &g, rc)

	cmd, err := app.Parse(os.Args[1:])
	if err != nil {
		fmt.Println("error", err)
		os.Exit(1)
	}
	rc.logger = setupLogger(*logLevel, *logFormat)

	g.Add(run.SignalHandler(context.Background(), syscall.SIGINT, syscall.SIGTERM))
	if err := g.Run(); err != nil {
		if *logLevel == "debug" {
			// Use %+v for github.com/pkg/errors error to print with stack.
			level.Error(rc.logger).Log("err", fmt.Sprintf("%+v", errors.Wrapf(err, "%s command failed", cmd)))
			os.Exit(1)
		}
		level.Error(rc.logger).Log("err", errors.Wrapf(err, "%s command failed", cmd))
		os.Exit(1)
	}
}
