package rndr

import (
	"fmt"
	"path/filepath"

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
		return TemplateDefinition{}, errors.New("template renderer has to be specified, got none")
	}

	// TODO(bwplotka): Add validation for renderers.
	switch {
	case t.Renderer.Jsonnet != nil:
		fmt.Println(t.Renderer.Jsonnet.Function)
		t.Renderer.Jsonnet.Function = abs(t.Renderer.Jsonnet.Function, tmplDir)
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
	Go    *GoTemplateAPI
	Proto *ProtoTemplateAPI
}

type GoTemplateAPI struct {
	// Default is a <full package path>.<public function> to be invoked to get valid struct filled in Entry
	Default string
	// Struct is a <full package path>.<public struct> name that should be used as the entry point for API struct.
	Struct string
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
	Jsonnet *JsonnetTemplateRenderer
	// Helm allows to configure a renderer that is able to take helm chart and input in YAMl and render output files.
	Helm *HelmTemplateRenderer
	// Process allows to configure a renderer that is able to execute process with YAMl passed by stdin or envvar and render output files.
	// `rndr` expects output resources to be rendered in stdout.
	Process *ProcessTemplateRenderer
}

type JsonnetTemplateRenderer struct {
	// Function represents a local or absolute path to .jsonnet file with single `function(values) {` to be an entry point for jsonnet template.
	Function string
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
