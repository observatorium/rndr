package rndr

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/go-kit/kit/log"
	"github.com/observatorium/rndr/pkg/rndr/engines/golang"
	"github.com/observatorium/rndr/pkg/rndr/engines/helm"
	"github.com/observatorium/rndr/pkg/rndr/engines/jsonnet"
	"github.com/observatorium/rndr/pkg/rndr/rndrapi"
	"github.com/pkg/errors"
)

type Template struct {
	// API is an input definition that will be used to validate template input YAML against or generate
	// required by packaging definitions e.g Custom Resource Definitions for Kubernetes.
	API API

	// Renderer is a mandatory expanding engine that converts input to desired output (e.g as Kubernetes YAMLs)
	Renderer TemplateRenderer
}

type API struct {
	// One of.
	Go    *golang.TemplateAPI
	Proto *ProtoTemplateAPI
}

type ProtoTemplateAPI struct {
	// Message is a name of root proto Message to be assumed as entry point for API in .proto file.
	Message string
	// File is destination to .proto file on local filesystem.
	File string
}

type TemplateRenderer struct {
	// One of.
	// Jsonnet allows to configure a renderer that is able to take jsonnet entry point file and input in YAMl and render output files.
	// `rndr` expects output resources to be rendered in stdout.
	Jsonnet *jsonnet.TemplateRenderer
	// Helm allows to configure a renderer that is able to take helm chart and input in YAMl and render output files.
	Helm *helm.TemplateRenderer
	// Process allows to configure a renderer that is able to execute process with YAMl passed by stdin or envvar and render output files.
	// `rndr` expects output resources to be rendered in stdout.
	Process *ProcessTemplateRenderer
}

type ProcessTemplateRenderer struct {
	Command   string
	Arguments []string
	// InputEnvVar controls the name of variable with input YAML content e.g `INPUT`.
	// If empty template input YAML is passed via stdin.
	InputEnvVar string
}

// RenderTemplate renders files based on template and values.
func RenderTemplate(_ context.Context, logger log.Logger, name string, t Template , valuesYAML []byte, outDir string) (err error) {
	// TODO(bwplotka): Parse values & validate through API (!).
	// TODO(bwplotka): Allow passing more parameters (e.g kubernetes options).
	var objectGroups rndrapi.Groups

	switch {
	case t.Renderer.Jsonnet != nil:
		objectGroups, err = jsonnet.Render(logger, name, *t.Renderer.Jsonnet, valuesYAML)
	case t.Renderer.Helm != nil:
		objectGroups, err = helm.Render(logger, name, *t.Renderer.Helm, valuesYAML)
	case t.Renderer.Process != nil:
		return errors.Errorf("process renderer is not implemented")
	default:
		return errors.Errorf("no renderer was specified")
	}
	if err != nil {
		return err
	}

	// TODO(bwplotka): Allow different dirs?
	for name, resources := range objectGroups {
		dir := filepath.Join(outDir, name)
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return err
		}
		for i, r := range resources {
			if err := ioutil.WriteFile(filepath.Join(dir, fmt.Sprintf("%d-%v.yaml", i, r.Item)), r.Object, os.ModePerm); err != nil {
				return err
			}
		}
	}
	return nil
}


