package constant

import (
	_ "embed"

	appsv1 "k8s.io/api/apps/v1"
	appsv1ac "k8s.io/client-go/applyconfigurations/apps/v1"
	"sigs.k8s.io/yaml"
)

const (
	Namespace              = "sandbox"
	StatefulSetName        = "test-sts"
	FieldManagerClient     = "test-sts-client"
	FieldManagerController = "test-sts-controller"
)

//go:embed template.yaml
var manifest []byte

func GetAppsV1StatefulSet() *appsv1.StatefulSet {
	sts := appsv1.StatefulSet{}
	_ = yaml.Unmarshal([]byte(manifest), &sts) // TODO: error check
	return &sts
}

func GetAppsV1StatefulSetAC() *appsv1ac.StatefulSetApplyConfiguration {
	sts := appsv1ac.StatefulSetApplyConfiguration{}
	_ = yaml.Unmarshal([]byte(manifest), &sts) // TODO: error check
	return &sts
}
