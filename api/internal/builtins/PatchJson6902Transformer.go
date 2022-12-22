// Code generated by pluginator on PatchJson6902Transformer; DO NOT EDIT.
// pluginator {unknown  1970-01-01T00:00:00Z  }

package builtins

import (
	"fmt"

	jsonpatch "github.com/evanphx/json-patch"
	"github.com/pkg/errors"
	"sigs.k8s.io/kustomize/api/filters/patchjson6902"
	"sigs.k8s.io/kustomize/api/ifc"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/kio/kioutil"
	"sigs.k8s.io/yaml"
)

type PatchJson6902TransformerPlugin struct {
	ldr          ifc.Loader
	decodedPatch jsonpatch.Patch
	Target       *types.Selector `json:"target,omitempty" yaml:"target,omitempty"`
	Path         string          `json:"path,omitempty" yaml:"path,omitempty"`
	JsonOp       string          `json:"jsonOp,omitempty" yaml:"jsonOp,omitempty"`
}

// noinspection GoUnusedGlobalVariable

func (p *PatchJson6902TransformerPlugin) Config(
	h *resmap.PluginHelpers, c []byte) (err error) {
	p.ldr = h.Loader()
	err = yaml.Unmarshal(c, p)
	if err != nil {
		return err
	}
	if p.Target.Name == "" {
		return fmt.Errorf("must specify the target name")
	}
	if p.Path == "" && p.JsonOp == "" {
		return fmt.Errorf("empty file path and empty jsonOp")
	}
	if p.Path != "" {
		if p.JsonOp != "" {
			return fmt.Errorf("must specify a file path or jsonOp, not both")
		}
		rawOp, err := p.ldr.Load(p.Path)
		if err != nil {
			return err
		}
		p.JsonOp = string(rawOp)
		if p.JsonOp == "" {
			return fmt.Errorf("patch file '%s' empty seems to be empty", p.Path)
		}
	}
	if p.JsonOp[0] != '[' {
		// if it doesn't seem to be JSON, imagine
		// it is YAML, and convert to JSON.
		op, err := yaml.YAMLToJSON([]byte(p.JsonOp))
		if err != nil {
			return err
		}
		p.JsonOp = string(op)
	}
	p.decodedPatch, err = jsonpatch.DecodePatch([]byte(p.JsonOp))
	if err != nil {
		return errors.Wrapf(err, "decoding %s", p.JsonOp)
	}
	if len(p.decodedPatch) == 0 {
		return fmt.Errorf(
			"patch appears to be empty; file=%s, JsonOp=%s", p.Path, p.JsonOp)
	}
	return err
}

func (p *PatchJson6902TransformerPlugin) Transform(m resmap.ResMap) error {
	if p.Target == nil {
		return fmt.Errorf("must specify a target for patch %s", p.JsonOp)
	}
	resources, err := m.Select(*p.Target)
	if err != nil {
		return err
	}

	if len(resources) == 0 {
		return fmt.Errorf("patchesJson6902 target not found for %s", p.Target.ResId)
	}

	for _, res := range resources {
		internalAnnotations := kioutil.GetInternalAnnotations(&res.RNode)

		err = res.ApplyFilter(patchjson6902.Filter{
			Patch: p.JsonOp,
		})
		if err != nil {
			return err
		}

		annotations := res.GetAnnotations()
		for key, value := range internalAnnotations {
			annotations[key] = value
		}
		err = res.SetAnnotations(annotations)
		if err != nil {
			return err
		}
	}
	return nil
}

func NewPatchJson6902TransformerPlugin() resmap.TransformerPlugin {
	return &PatchJson6902TransformerPlugin{}
}
