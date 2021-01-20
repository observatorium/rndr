package rndr

import (
	"path/filepath"

	"github.com/observatorium/rndr/pkg/rndr/golang"
	"github.com/observatorium/rndr/pkg/rndr/jsonnet"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

type TemplateDefinition struct {
	Name    string
	Authors string

	// API is an input definition that will be used to validate template input YAML against and generate
	// Custom Resource Definitions for Kubernetes.
	// It's recommended to define your definitions.
	// Otherwise if empty, no validation will be in place as well as rndr will fail when Kubernetes operator is requested.
	API TemplateAPI

	// Renderer is a mandatory expanding engine that converts input to desired deployment resources (e.g as Kuberentes YAMLs)
	Renderer TemplateRenderer
}

// ParseTemplate parses TemplateDefinition from bytes.
func ParseTemplate(b []byte, tmplDir string) (TemplateDefinition, error) {
	t := TemplateDefinition{}
	if err := yaml.Unmarshal(b, &t); err != nil {
		return TemplateDefinition{}, errors.Wrapf(err, "parsing template content %q", string(b))
	}

	if t.Name == "" {
		return TemplateDefinition{}, errors.New("name not specified, but required")
	}

	if t.Authors == "" {
		return TemplateDefinition{}, errors.New("authors not specified, but required")
	}

	switch {
	case t.API.Go != nil:
		if t.API.Go.Struct == "" {
			return TemplateDefinition{}, errors.New("api.go.struct not specified, but required")
		}
	case t.API.Proto != nil:
		if t.API.Proto.Message == "" {
			return TemplateDefinition{}, errors.New("api.proto.message not specified, but required")
		}
		if t.API.Proto.File == "" {
			return TemplateDefinition{}, errors.New("api.proto.file not specified, but required")
		}
		t.API.Proto.File = abs(t.API.Proto.File, tmplDir)
	default:
		return TemplateDefinition{}, errors.New("template api has to be specified, got none")
	}

	// TODO(bwplotka): Add validation for renderers.
	switch {
	case t.Renderer.Jsonnet != nil:
		if len(t.Renderer.Jsonnet.Functions) == 0 {
			return TemplateDefinition{}, errors.New("jsonnet template renderer has to have at least single function file specified, got none")
		}
		for i := range t.Renderer.Jsonnet.Functions {
			t.Renderer.Jsonnet.Functions[i] = abs(t.Renderer.Jsonnet.Functions[i], tmplDir)
		}

	case t.Renderer.Helm != nil:
	case t.Renderer.Process != nil:
		t.Renderer.Process.Command = abs(t.Renderer.Process.Command, tmplDir)
	default:
		return TemplateDefinition{}, errors.New("template renderer has to be specified, got none")
	}
	return t, nil
}

func abs(path string, relDir string) string {
	if relDir == "" {
		return path
	}

	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(relDir, path)
}

type TemplateAPI struct {
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
	Helm *HelmTemplateRenderer
	// Process allows to configure a renderer that is able to execute process with YAMl passed by stdin or envvar and render output files.
	// `rndr` expects output resources to be rendered in stdout.
	Process *ProcessTemplateRenderer
}

type HelmTemplateRenderer struct {
	// Chart is a chart name.
	Chart string
	// Repo is an URL to the chart repository if remote.
	Repo string
	// Version is a chart version within the repo.
	Version string
}

type ProcessTemplateRenderer struct {
	Command   string
	Arguments []string
	// InputEnvVar controls the name of variable with input YAML content e.g `INPUT`.
	// If empty template input YAML is passed via stdin.
	InputEnvVar string
}
