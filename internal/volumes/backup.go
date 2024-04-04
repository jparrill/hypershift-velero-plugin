package volumes

import (
	"context"

	commonplug "github.com/jparrill/hypershift-velero-plugin/internal/common"
	"github.com/sirupsen/logrus"
	v1 "github.com/vmware-tanzu/velero/pkg/apis/velero/v1"
	"github.com/vmware-tanzu/velero/pkg/plugin/velero"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	logHeader = "[volume-backup]"
)

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
	return "hypershiftVolumesBackupPlugin"
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
		},
	}, nil
}

// Execute allows the ItemAction to perform arbitrary logic with the item being backed up,
// in this case, setting a custom annotation on the item being backed up.
func (p *BackupPlugin) Execute(item runtime.Unstructured, backup *v1.Backup) (runtime.Unstructured, []velero.ResourceIdentifier, error) {
	p.log.Debugf("%s Entering Hypershift Volumes backup plugin", logHeader)
	ctx := context.Context(context.TODO())

	p.log.Debugf("%s Getting client", logHeader)
	client, err := commonplug.GetClient()
	if err != nil {
		return nil, nil, err
	}

	// Pausing NodePools
	if err := commonplug.ManagePauseNodepools(ctx, client, p.log, "true", logHeader, backup.Spec.IncludedNamespaces); err != nil {
		return nil, nil, err
	}

	// Pausing HostedClusters
	if err := commonplug.ManagePauseHostedCluster(ctx, client, p.log, "true", logHeader, backup.Spec.IncludedNamespaces); err != nil {
		return nil, nil, err
	}

	return item, nil, nil
}
