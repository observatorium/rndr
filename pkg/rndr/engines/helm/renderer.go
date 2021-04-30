package helm

import (
	"github.com/go-kit/kit/log"
	"github.com/observatorium/rndr/pkg/rndr/rndrapi"
	"github.com/pkg/errors"
)

type TemplateRenderer struct {
	// Chart is a chart name.
	Chart string
	// Repo is an URL to the chart repository if remote.
	Repo string
	// Version is a chart version within the repo.
	Version string
}

func Render(logger log.Logger, name string, c TemplateRenderer, valuesJSON []byte) (groups rndrapi.Groups, err error) {
	return nil, errors.Errorf("helm renderer is not implemented")
}
