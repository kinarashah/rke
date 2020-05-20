package templates

import (
	"bytes"
	"encoding/json"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig"
	"github.com/ghodss/yaml"
	"github.com/rancher/rke/util"
	"github.com/sirupsen/logrus"
)

var VersionedTemplate = map[string]map[string]string{
	"calico": map[string]string{
		"v1.15":   CalicoTemplateV115,
		"v1.14":   CalicoTemplateV113,
		"v1.13":   CalicoTemplateV113,
		"default": CalicoTemplateV112,
	},
	"canal": map[string]string{
		"v1.15":   CanalTemplateV115,
		"v1.14":   CanalTemplateV113,
		"v1.13":   CanalTemplateV113,
		"default": CanalTemplateV112,
	},
	"flannel": map[string]string{
		"v1.15":   FlannelTemplateV115,
		"default": FlannelTemplate,
	},
}

func CompileTemplateFromMap(tmplt string, configMap interface{}) (string, error) {
	out := new(bytes.Buffer)
	templateFuncMap := sprig.TxtFuncMap()
	templateFuncMap["toYaml"] = ToYAML
	t := template.Must(template.New("compiled_template").Parse(tmplt))
	if err := t.Execute(out, configMap); err != nil {
		return "", err
	}
	return out.String(), nil
}

func GetVersionedTemplates(templateName string, k8sVersion string) string {

	versionedTemplate := VersionedTemplate[templateName]
	if t, ok := versionedTemplate[util.GetTagMajorVersion(k8sVersion)]; ok {
		return t
	}
	return versionedTemplate["default"]
}

func ToYAML(v interface{}) string {
	data, err := json.Marshal(v)
	if err != nil {
		// Swallow errors inside of a template so it doesn't affect remaining template lines
		logrus.Errorf("[ToYAML] Error marshaling %v: %v", v, err)
		return ""
	}
	yamlData, err := yaml.JSONToYAML(data)
	if err != nil {
		// Swallow errors inside of a template so it doesn't affect remaining template lines
		logrus.Errorf("[ToYAML] Error converting json to yaml for %v: %v ", string(data), err)
		return ""
	}
	return strings.TrimSuffix(string(yamlData), "\n")
}