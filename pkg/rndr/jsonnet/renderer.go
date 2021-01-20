package jsonnet

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/brancz/locutus/render/jsonnet"
	"github.com/brancz/locutus/rollout"
	"github.com/efficientgo/tools/core/pkg/errcapture"
	"github.com/efficientgo/tools/core/pkg/logerrcapture"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

type TemplateRenderer struct {
	// Functions represent a local or absolute paths to .jsonnet files with
	// single `function(values) {` that renders manifests in right order.
	// Each function's manifests will be part of different groups allowing parallel rollout if requested.
	Functions []string
}

// TODO(bwplotka): This is bit fuzzy. Potentially we need more control on what is rolled when. Improve.
// TODO(bwplotka): I assume rollout groups allows paralelism. Check that.
var applyAllLocutusJsonnetTmpl = template.Must(template.New("").Parse(`
local values = import '{{ .LocutusVirtualConfigPath }}';

local groups = {
{{- range .Groups }}
  '{{ .Prefix }}': (import '{{ .FunctionFile }}')(values),
{{-  end }}
};

{
  objects: {
		[g + '#' + item]: groups[g]
		for g in std.objectFields(groups)
		for item in std.objectFields(groups[g])
  },
  rollout: {
    apiVersion: 'workflow.kubernetes.io/v1alpha1',
    kind: 'Rollout',
    metadata: {
	  template: '{{ .Name }}',
      name: 'rndr-generated-jsonnet',
    },
    spec: {
      groups: [
		{
		  name: g,
		  steps: [
			{
			  action: 'CreateOrUpdate',
			  object: g + '#' + item,
			}
			for item in std.objectFields(groups[g])
		  ],
		},
		for g in std.objectFields(groups)
	  ],
    },
  },
}`))

// locutusify takes function file path and creates Locutus boilerplate in specified location
// that takes locutus specific jsonnet file that imports and executes function file with values taken from
// locutus specific path specified by jsonnet.VirtualConfigPath.
// NOTE(bwplotka): We are reinventing invocation part for two reason:
// * To simplify input passing and not leak locutus existence. User does not need to know exact "virtual config path" which might be non-intuitive to learn
// about in the first place.
// * To make sure jsonnet templates can be run natively without mocking any input (e.g for testing purposes)
// * To reduce user knowledge about locutus rollout abstractions. `rndr` is meant as a tool which can be used reliably
// for both operator, helm and GitOps flows. If the operator flow is not necessary (for example for stateless services) we
// want to make sure no operator will be deployed. This significantly reduces simplifies that stack if it can be simplified.
// TODO(bwplotka): Potentially something to upstream on Locutus side.
func locutusify(entry string, templName string, functionFiles []string) (err error) {
	if err := os.RemoveAll(entry); err != nil {
		return err
	}

	f, err := os.Create(entry)
	if err != nil {
		return err
	}
	defer errcapture.Do(&err, f.Close, "close locutus entry")

	type group struct {
		Prefix       string
		FunctionFile string
	}

	g := make([]group, 0, len(functionFiles))
	for _, f := range functionFiles {
		// TODO(bwplotka): Ensure name clashes are handled.
		g = append(g, group{FunctionFile: f, Prefix: strings.TrimSuffix(filepath.Base(f), filepath.Ext(filepath.Base(f)))})
	}

	return applyAllLocutusJsonnetTmpl.Execute(f, struct {
		Name                     string
		LocutusVirtualConfigPath string
		Groups                   []group
	}{
		Name:                     templName,
		Groups:                   g,
		LocutusVirtualConfigPath: jsonnet.VirtualConfigPath,
	})
}

type Resource struct {
	Item       string
	ObjectYAML []byte
}

func Render(logger log.Logger, name string, c TemplateRenderer, valuesJSON []byte) (groups map[string][]Resource, err error) {
	tmpDir, err := ioutil.TempDir(os.TempDir(), "rndr")
	if err != nil {
		return nil, err
	}
	defer logerrcapture.Do(logger, func() error { return os.RemoveAll(tmpDir) }, "remove tmp dir")

	entry := filepath.Join(tmpDir, "main.jsonnet")
	if err := locutusify(entry, name, c.Functions); err != nil {
		return nil, errors.Wrap(err, "locutusify")
	}

	level.Debug(logger).Log("msg", "rendered jsonnet intermidiate file", "path", entry)

	res, err := jsonnet.NewRenderer(logger, entry, nil).Render(valuesJSON)
	if err != nil {
		return nil, errors.Wrapf(err, "render jsonnet from options %+v", c)
	}

	if res.Rollout == nil || len(res.Rollout.Spec.Groups) == 0 {
		return nil, errors.Errorf("no rollout resource rendered by locutus entry %v", entry)
	}
	ret := make(map[string][]Resource, len(res.Rollout.Spec.Groups))

	// We control boilerplate so we expect purely CreateOrUpdate actions.
	for _, g := range res.Rollout.Spec.Groups {
		for _, s := range g.Steps {
			if s.Action != (&rollout.CreateOrUpdateObjectAction{}).Name() {
				return nil, errors.Errorf("rollout step rendered by locutus has unexpected action %v", s)
			}

			if _, ok := res.Objects[s.Object]; !ok {
				return nil, errors.Errorf("rollout step rendered by locutus has object that does not exists %v", s.Object)
			}

			split := strings.Split(s.Object, "#")

			// TODO(bwplotka): Most likely we have to stick to JSON output.
			b := bytes.Buffer{}
			m := yaml.NewEncoder(&b)
			m.SetIndent(2)

			if err := m.Encode(res.Objects[s.Object].Object); err != nil {
				return nil, err
			}

			ret[g.Name] = append(ret[g.Name], Resource{Item: split[1], ObjectYAML: b.Bytes()})
		}
	}
	return ret, nil
}
