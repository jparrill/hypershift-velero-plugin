package volumes

import (
	commonplug "github.com/jparrill/hypershift-velero-plugin/internal/common"
	"github.com/sirupsen/logrus"
	"github.com/vmware-tanzu/velero/pkg/plugin/velero"
	"k8s.io/apimachinery/pkg/api/meta"
)

type RestorePlugin struct {
	log logrus.FieldLogger
}

func NewRestorePlugin(log logrus.FieldLogger) *RestorePlugin {
	return &RestorePlugin{log: log}
}

func (p *RestorePlugin) Name() string {
	return "hypershiftVolumeRestorePlugin"
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
		},
	}, nil
}

// Execute sets a custom annotation on the item being restored.
func (p *RestorePlugin) Execute(input *velero.RestoreItemActionExecuteInput) (*velero.RestoreItemActionExecuteOutput, error) {
	p.log.Info("[volume-restore] Entering Hypershift Volumes restore plugin")
	//ctx := context.Context(context.TODO())
	metadata, err := meta.Accessor(input.Item)
	if err != nil {
		return nil, err
	}

	annotations := metadata.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}

	name := metadata.GetName()
	p.log.Infof("[volume-restore] Volume restore plugin for %s", name)

	annotations[commonplug.CommonRestoreAnnotationName] = string(commonplug.RestoreDone)

	if annotations[commonplug.CommonBackupAnnotationName] != "" {
		delete(annotations, commonplug.CommonBackupAnnotationName)
	}

	metadata.SetAnnotations(annotations)

	return velero.NewRestoreItemActionExecuteOutput(input.Item), nil
}
