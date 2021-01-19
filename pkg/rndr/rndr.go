package rndr

import (
	"context"

	"github.com/brancz/locutus/render/file"
	"github.com/brancz/locutus/render/jsonnet"
	"github.com/brancz/locutus/rollout"
	"github.com/go-kit/kit/log"
)

func Render(ctx context.Context, logger log.Logger, t TemplateDefinition, values []byte, outDir string) error {
	var renderer rollout.Renderer
	{
		switch renderProviderName {
		case "jsonnet":
			renderer = jsonnet.NewRenderer(logger, rendererJsonnetEntrypoint, sources)
		case "file":
			renderer = file.NewRenderer(logger, rendererFileDirectory, rendererFileRollout)
		default:
			logger.Log("msg", "failed to find render provider")
			return 1
		}
	}
}
