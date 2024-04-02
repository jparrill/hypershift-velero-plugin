package common

import (
	"github.com/sirupsen/logrus"
	v1 "github.com/vmware-tanzu/velero/pkg/apis/velero/v1"
	"github.com/vmware-tanzu/velero/pkg/plugin/velero"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
)

// BackupPlugin is a backup item action plugin for Hypershift common objects.
type BackupPlugin struct {
	log logrus.FieldLogger
}

// NewBackupPlugin instantiates BackupPlugin.
func NewBackupPlugin(log logrus.FieldLogger) *BackupPlugin {
	return &BackupPlugin{log: log}
}

// Name is required to implement the interface, but the Velero pod does not delegate this
// method -- it's used to tell velero what name it was registered under. The plugin implementation
// must define it, but it will never actually be called.
func (p *BackupPlugin) Name() string {
	return "hypershiftCommonBackupPlugin"
}

// AppliesTo returns information about which resources this action should be invoked for.
// The IncludedResources and ExcludedResources slices can include both resources
// and resources with group names. These work: "ingresses", "ingresses.extensions".
// A BackupPlugin's Execute function will only be invoked on items that match the returned
// selector. A zero-valued ResourceSelector matches all resources.
func (p *BackupPlugin) AppliesTo() (velero.ResourceSelector, error) {
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

// Execute allows the ItemAction to perform arbitrary logic with the item being backed up,
// in this case, setting a custom annotation on the item being backed up.
func (p *BackupPlugin) Execute(item runtime.Unstructured, backup *v1.Backup) (runtime.Unstructured, []velero.ResourceIdentifier, error) {
	p.log.Info("[common-backup] Entering Hypershift common restore plugin")

	metadata, err := meta.Accessor(item)
	if err != nil {
		return nil, nil, err
	}

	annotations := metadata.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}

	if annotations[CommonBackupAnnotationName] == "" {
		p.log.Infof("[common-backup] Setting annotation for item, %s", metadata.GetName())
		annotations[CommonBackupAnnotationName] = string(BackupStatusInProgress)
		metadata.SetAnnotations(annotations)
	}

	return item, nil, nil
}
