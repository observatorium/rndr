package rndr

import (
	"gopkg.in/yaml.v3"
)

type TemplateDefinition struct {
	Version  string
	Template string
	Authors  string

	// API is an input definition that will be used to validate template input YAML against and generate
	// Custom Resource Definitions for Kubernetes.
	// It's recommended to define your definitions.
	// Otherwise if empty, no validation will be in place as well as rndr will fail when Kubernetes operator is requested.
	API TemplateAPI

	// Renderer is a mandatory expanding engine that converts input to desired deployment resources (e.g as Kuberentes YAMLs)
	Renderer TemplateRenderer

	// Tells rndr and renderer where generated resources should land.
	Output TemplateOutput
}

// ParseTemplate parses TemplateDefinition from bytes.
func ParseTemplate(b []byte) (TemplateDefinition, error) {
	t := TemplateDefinition{}
	if err := yaml.Unmarshal(b, &t); err != nil {
		return TemplateDefinition{}, err
	}
	return t, nil
}

type TemplateAPI struct {
	// One of.
	Go    GoTemplateAPI
	Proto ProtoTemplateAPI
}

type GoTemplateAPI struct {
	Entry   string
	Package string
}

type ProtoTemplateAPI struct {
	Entry string
	File  string
}

// TODO(bwplotka): Allow building all into Go binary? What if files are too large to be part of binary?
type TemplateRenderer struct {
	// One of.
	// Jsonnet allows to configure a renderer that is able to take jsonnet entry point file and input in YAMl and render output files.
	Jsonnet JsonnetTemplateRenderer
	// Helm allows to configure a renderer that is able to take helm chart and input in YAMl and render output files.
	Helm HelmTemplateRenderer
	// Process allows to configure a renderer that is able to execute process with YAMl passed by stdin or envvar and render output files.
	Process ProcessTemplateRenderer
}

type TemplateOutput struct {
	// TODO(bwplotka): Allow defining custom, parallel deployment logics?
	// NOTE: Resources are meant to be deployed in the lexicographic order.
	Directories []string
}

type JsonnetTemplateRenderer struct {
	// entrypoint represents entry .jsonnet file to be executed.
	entrypoint string
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
