package common

import (
	"context"
	"fmt"

	hyperv1 "github.com/openshift/hypershift/api/hypershift/v1beta1"
	"github.com/sirupsen/logrus"
	v1 "github.com/vmware-tanzu/velero/pkg/apis/velero/v1"
	"github.com/vmware-tanzu/velero/pkg/plugin/velero"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	crclient "sigs.k8s.io/controller-runtime/pkg/client"
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
	p.log.Info("[common-backup] Entering Hypershift common backup plugin")
	ctx := context.Context(context.TODO())

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

	p.log.Infof("[common-backup] Checking NodePool resources", metadata.GetName())
	client, err := GetClient()
	if err != nil {
		return nil, nil, err
	}

	nps := &hyperv1.NodePoolList{}
	if err := client.List(ctx, nps, crclient.InNamespace("clusters")); err != nil {
		return nil, nil, err
	}

	p.log.Info("[common-backup] Creating Test Secret per existent NodePool")

	for _, np := range nps.Items {
		p.log.Infof("[common-backup] Creating test Secret per existent NodePool %s", np.Name)
		if err := client.Create(ctx, crclient.Object(&corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("test-secret-%s", np.Name),
				Namespace: "clusters",
			},
			Data: map[string][]byte{
				"test": []byte("test"),
			},
		})); err != nil {
			if !apierrors.IsAlreadyExists(err) {
				return nil, nil, err
			}
		}
	}

	return item, nil, nil
}
