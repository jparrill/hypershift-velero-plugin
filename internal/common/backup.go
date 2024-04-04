package common

import (
	"context"

	"github.com/sirupsen/logrus"
	v1 "github.com/vmware-tanzu/velero/pkg/apis/velero/v1"
	"github.com/vmware-tanzu/velero/pkg/plugin/velero"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	logHeader = "[common-backup]"
)

// BackupPlugin is a backup item action plugin for Hypershift common objects.
type BackupPlugin struct {
	log            logrus.FieldLogger
	DataUploadDone bool
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
	p.log.Debugf("%s Entering Hypershift common backup plugin", logHeader)
	var err error

	ctx := context.Context(context.TODO())

	p.log.Debugf("%s Getting Client", logHeader)
	client, err := GetClient()
	if err != nil {
		return nil, nil, err
	}

	if !p.DataUploadDone {
		if item.GetObjectKind().GroupVersionKind().Kind == "Secret" {
			p.log.Infof("%s Secret section reached", logHeader)
			// This function will wait before the secrets got backed up.
			// This is a workaround because of the limitations of velero plugins and hooks.
			// We need to think how to acomplish that in a better way in the final solution.
			if p.DataUploadDone, err = WaitForDataUpload(ctx, client, p.log, backup); err != nil {
				return nil, nil, err
			}
		}
	}

	if p.DataUploadDone {
		// Unpausing NodePools
		if err := ManagePauseNodepools(ctx, client, p.log, "false", logHeader, backup.Spec.IncludedNamespaces); err != nil {
			return nil, nil, err
		}

		// Unpausing HostedClusters
		if err := ManagePauseHostedCluster(ctx, client, p.log, "false", logHeader, backup.Spec.IncludedNamespaces); err != nil {
			return nil, nil, err
		}
	}

	return item, nil, nil
}
