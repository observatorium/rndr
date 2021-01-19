package main

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/brancz/locutus/render/file"
	"github.com/brancz/locutus/render/jsonnet"
	"github.com/brancz/locutus/rollout"
	"github.com/bwplotka/mdox/pkg/version"
	"github.com/efficientgo/tools/core/pkg/clilog"
	"github.com/efficientgo/tools/core/pkg/errcapture"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/oklog/run"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/expfmt"
	"gopkg.in/alecthomas/kingpin.v2"
"github.com/openproto/protoconfig/go/kingpinv2"
	"k8s.io/client-go/kubernetes"
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

func main() {
	app := kingpin.New(filepath.Base(os.Args[0]), `Deployments Template Toolkit.`).Version(version.Version)
	logLevel := app.Flag("log.level", "Log filtering level.").
		Default("info").Enum("error", "warn", "info", "debug")
	logFormat := app.Flag("log.format", "Log format to use.").
		Default(logFormatCLILog).Enum(logFormatLogfmt, logFormatJSON, logFormatCLILog)
	tmpl := kingpinv2.Flag(app, "template", "Deployment template")



	ctx, cancel := context.WithCancel(context.Background())
	registerFmt(ctx, app)
	registerWeb(ctx, app)

	cmd, runner := app.Parse()
	logger := setupLogger(*logLevel, *logFormat)

	var g run.Group
	g.Add(func() error {
		// TODO(bwplotka): Move to customized better setup function.
		return runner(ctx, logger)
	}, func(err error) {
		cancel()
	})
/*
   	cmd := app.Command("fmt", "Formats in-place given markdown files uniformly following GFM (Github Flavored Markdown: https://github.github.com/gfm/). Example: mdox fmt *.md")
   	files := cmd.Arg("files", "Markdown file(s) to process.").Required().ExistingFiles()
   	checkOnly := cmd.Flag("check", "If true, fmt will not modify the given files, instead it will fail if files needs formatting").Bool()

   	disableGenCodeBlocksDirectives := cmd.Flag("code.disable-directives", `If false, fmt will parse custom fenced code directives prefixed with 'mdox-gen' to autogenerate code snippets. For example:
   	`+"```"+`<lang> mdox-gen-exec="<executable + arguments>"
   This directive runs executable with arguments and put its stderr and stdout output inside code block content, replacing existing one.`).Bool()
   	anchorDir := cmd.Flag("anchor-dir", "Anchor directory for all transformers. PWD is used if flag is not specified.").ExistingDir()
   	linksLocalizeForAddress := cmd.Flag("links.localize.address-regex", "If specified, all HTTP(s) links that target a domain and path matching given regexp will be transformed to relative to anchor dir path (if exists)."+
   		"Absolute path links will be converted to relative links to anchor dir as well.").Regexp()
   	// TODO(bwplotka): Add cache in file?
   	linksValidateEnabled := cmd.Flag("links.validate", "If true, all links will be validated").Short('l').Bool()
   	linksValidateExceptDomains := cmd.Flag("links.validate.without-address-regex", "If specified, all links will be validated, except those matching the given target address.").Default(`^$`).Regexp()

   	cmd.Run(func(ctx context.Context, logger log.Logger) (err error) {
   		var opts []mdformatter.Option
   		if !*disableGenCodeBlocksDirectives {
   			opts = append(opts, mdformatter.WithCodeBlockTransformer(mdgen.NewCodeBlockTransformer()))

   		if *checkOnly {
   			diff, err := mdformatter.IsFormatted(ctx, logger, *files, opts...)
   			if err != nil {
   				return err
   			}
   			if len(diff) == 0 {
   				return nil
   			}
   			return errors.Errorf("files not formatted: %v", diff.String())
   		}
   		return mdformatter.Format(ctx, logger, *files, opts...)
   	})
   }
   	cmd := app.Command("web", "Tools for generating static HTML website based on https://gohugo.io/ on every PR with preview")
   	genCmd := cmd.Command("gen", "Generate versioned docs")

   	_ = genCmd.Arg("files", "Markdown file(s) to process.").Required().ExistingFiles()

   	// TODO(bwplotka): Generate versioned docs used for Hugo. Open question: Hugo specific? How to adjust links? Should fmt and this be
   	// the same?
   	cmd.Run(func(ctx context.Context, logger log.Logger) error {
   		return errors.New("not implemented")
   	})
   }

 */
	//var renderer rollout.Renderer
	//{
	//	switch renderProviderName {
	//	case "jsonnet":
	//		renderer = jsonnet.NewRenderer(logger, rendererJsonnetEntrypoint, sources)
	//	case "file":
	//		renderer = file.NewRenderer(logger, rendererFileDirectory, rendererFileRollout)
	//	default:
	//		logger.Log("msg", "failed to find render provider")
	//		return 1
	//	}
	//}

	// Listen for termination signals.
	{
		cancel := make(chan struct{})
		g.Add(func() error {
			return interrupt(logger, cancel)
		}, func(error) {
			close(cancel)
		})
	}

	if err := g.Run(); err != nil {
		if *logLevel == "debug" {
			// Use %+v for github.com/pkg/errors error to print with stack.
			level.Error(logger).Log("err", fmt.Sprintf("%+v", errors.Wrapf(err, "%s command failed", cmd)))
			os.Exit(1)
		}
		level.Error(logger).Log("err", errors.Wrapf(err, "%s command failed", cmd))
		os.Exit(1)
	}
}

func interrupt(logger log.Logger, cancel <-chan struct{}) error {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	select {
	case s := <-c:
		level.Info(logger).Log("msg", "caught signal. Exiting.", "signal", s)
		return nil
	case <-cancel:
		return errors.New("canceled")
	}
}