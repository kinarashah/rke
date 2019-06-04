package rke

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/rancher/kontainer-driver-metadata/rke/templates"

	"github.com/rancher/types/apis/management.cattle.io/v3"
	projectv3 "github.com/rancher/types/apis/project.cattle.io/v3"
	"github.com/rancher/types/image"
)

const (
	rkeDataFilePath = "./data/data.json"
)

// Data to be written in dataFilePath, dynamically populated on init() with the latest versions
type Data struct {

	// K8sVersionServiceOptions - service options per k8s version
	K8sVersionServiceOptions  map[string]v3.KubernetesServicesOptions
	K8sVersionRKESystemImages map[string]v3.RKESystemImages

	// Addon Templates per K8s version / "default" where nothing changes for k8s version
	K8sVersionedTemplates map[string]map[string]string

	// K8sVersionToRKEVersions - min/max RKE versions per k8s version
	K8sVersionToRKEVersions map[string]v3.RKEVersions

	// K8sVersionToRancherVersions - min/max Rancher versions per k8s version
	K8sVersionToRancherVersions map[string]v3.RancherVersion

	//Default K8s version for every rancher version
	RancherDefaultK8sVersions map[string]string

	K8sVersionWindowsSystemImages   map[string]v3.WindowsSystemImages
	K8sVersionWindowsServiceOptions map[string]v3.KubernetesServicesOptions

	// ToolsSystemImages default images for alert, pipeline, logging, globaldns
	ToolsSystemImages ToolsSystemImages
}

type ToolsSystemImages struct {
	AlertSystemImages    v3.AlertSystemImages
	PipelineSystemImages projectv3.PipelineSystemImages
	LoggingSystemImages  v3.LoggingSystemImages
	AuthSystemImages     v3.AuthSystemImages
}

var (
	DriverData Data
	m          = image.Mirror

	// Default k8s version per rke version
	RKEDefaultK8sVersions map[string]string
)

func InitRKE(writeData bool) {
	DriverData = Data{
		K8sVersionRKESystemImages: loadK8sRKESystemImages(),
	}

	for version, images := range DriverData.K8sVersionRKESystemImages {
		longName := "rancher/hyperkube:" + version
		if !strings.HasPrefix(longName, images.Kubernetes) {
			panic(fmt.Sprintf("For K8s version %s, the Kubernetes image tag should be a substring of %s, currently it is %s", version, version, images.Kubernetes))
		}
	}

	DriverData.K8sVersionServiceOptions = loadK8sVersionServiceOptions()

	DriverData.K8sVersionToRKEVersions = loadK8sRKEVersions()

	DriverData.K8sVersionToRancherVersions = loadK8sRancherVersions()

	DriverData.K8sVersionedTemplates = templates.LoadK8sVersionedTemplates()

	RKEDefaultK8sVersions = loadRKEDefaultK8sVersions()

	for _, defaultK8s := range RKEDefaultK8sVersions {
		if _, ok := DriverData.K8sVersionRKESystemImages[defaultK8s]; !ok {
			panic(fmt.Sprintf("Default K8s version %v is not found in the K8sVersionToRKESystemImages", defaultK8s))
		}
	}

	// init Windows versions
	DriverData.K8sVersionWindowsSystemImages = loadK8sVersionWindowsSystemimages()
	DriverData.K8sVersionWindowsServiceOptions = loadK8sVersionWindowsServiceOptions()

	DriverData.ToolsSystemImages = getToolsSystemImages()

	DriverData.RancherDefaultK8sVersions = loadRancherDefaultK8sVersions()

	if writeData {
		//todo: more optimization on how data is stored in file
		strData, _ := json.MarshalIndent(DriverData, "", " ")
		jsonFile, err := os.Create(rkeDataFilePath)
		if err != nil {
			panic(fmt.Errorf("err %v", err))
		}
		jsonFile.Write(strData)
		jsonFile.Close()
	}

}

func getToolsSystemImages() ToolsSystemImages {
	return ToolsSystemImages{
		AlertSystemImages: v3.AlertSystemImages{
			AlertManager:       m("prom/alertmanager:v0.15.2"),
			AlertManagerHelper: m("rancher/alertmanager-helper:v0.0.2"),
		},
		PipelineSystemImages: projectv3.PipelineSystemImages{
			Jenkins:       m("rancher/pipeline-jenkins-server:v0.1.0"),
			JenkinsJnlp:   m("jenkins/jnlp-slave:3.10-1-alpine"),
			AlpineGit:     m("rancher/pipeline-tools:v0.1.9"),
			PluginsDocker: m("plugins/docker:17.12"),
			Minio:         m("minio/minio:RELEASE.2018-05-25T19-49-13Z"),
			Registry:      m("registry:2"),
			RegistryProxy: m("rancher/pipeline-tools:v0.1.9"),
			KubeApply:     m("rancher/pipeline-tools:v0.1.9"),
		},
		LoggingSystemImages: v3.LoggingSystemImages{
			Fluentd:                       m("rancher/fluentd:v0.1.11"),
			FluentdHelper:                 m("rancher/fluentd-helper:v0.1.2"),
			LogAggregatorFlexVolumeDriver: m("rancher/log-aggregator:v0.1.4"),
		},
		AuthSystemImages: v3.AuthSystemImages{
			KubeAPIAuth: m("rancher/kube-api-auth:v0.1.3"),
		},
	}
}
