package common

import (
	"github.com/sirupsen/logrus"
	"github.com/vmware-tanzu/velero/pkg/plugin/velero"
)

// RestorePlugin is a restore item action plugin for hypershift common objects.
type RestorePlugin struct {
	log logrus.FieldLogger
}

// NewRestorePlugin instantiates a RestorePlugin.
func NewRestorePlugin(log logrus.FieldLogger) *RestorePlugin {
	return &RestorePlugin{log: log}
}

// Name is required to implement the interface, but the Velero pod does not delegate this
// method -- it's used to tell velero what name it was registered under. The plugin implementation
// must define it, but it will never actually be called.
func (p *RestorePlugin) Name() string {
	return "hypershiftCommonRestorePlugin"
}

// AppliesTo returns information about which resources this action should be invoked for.
// The IncludedResources and ExcludedResources slices can include both resources
// and resources with group names. These work: "ingresses", "ingresses.extensions".
// A RestoreItemAction's Execute function will only be invoked on items that match the returned
// selector. A zero-valued ResourceSelector matches all resources.
func (p *RestorePlugin) AppliesTo() (velero.ResourceSelector, error) {
	return velero.ResourceSelector{
		IncludedResources: []string{
			"pv",
			"pvc",
			"hostedcluster",
			"nodepool",
			"secrets",
			"hostedcontrolplane",
			"cluster",
			"awscluster",
			"awsmachinetemplate",
			"awsmachine",
			"machinedeployment",
			"machineset",
			"machine",
		},
	}, nil
}

// Execute sets a custom annotation on the item being restored.
func (p *RestorePlugin) Execute(input *velero.RestoreItemActionExecuteInput) (*velero.RestoreItemActionExecuteOutput, error) {
	p.log.Info("[common-restore] Entering Hypershift common restore plugin")

	metadata, annotations, err := getMetadataAndAnnotations(input.Item)
	if err != nil {
		return nil, err
	}
	name := metadata.GetName()
	p.log.Infof("[common-restore] common restore plugin for %s", name)

	annotations[CommonRestoreAnnotationName] = string(RestoreDone)

	if annotations[CommonBackupAnnotationName] != "" {
		delete(annotations, CommonBackupAnnotationName)
	}

	metadata.SetAnnotations(annotations)

	return velero.NewRestoreItemActionExecuteOutput(input.Item), nil
}
