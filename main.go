package main

import (
	plugcommon "github.com/jparrill/hypershift-velero-plugin/internal/common"
	plugvols "github.com/jparrill/hypershift-velero-plugin/internal/volumes"
	"github.com/sirupsen/logrus"
	"github.com/vmware-tanzu/velero/pkg/plugin/framework"
)

func main() {
	framework.NewServer().
		RegisterBackupItemAction("hypershift.openshift.io/common-backup-plugin", newCommonBackupPlugin).
		RegisterRestoreItemAction("hypershift.openshift.io/common-restore-plugin", newCommonRestorePlugin).
		RegisterBackupItemAction("hypershift.openshift.io/volumes-backup-plugin", newVolumesBackupPlugin).
		RegisterRestoreItemAction("hypershift.openshift.io/volumes-restore-plugin", newVolumesRestorePlugin).
		Serve()
}

func newCommonBackupPlugin(logger logrus.FieldLogger) (interface{}, error) {
	return plugcommon.NewBackupPlugin(logger), nil
}

func newCommonRestorePlugin(logger logrus.FieldLogger) (interface{}, error) {
	return plugcommon.NewRestorePlugin(logger), nil
}

func newVolumesBackupPlugin(logger logrus.FieldLogger) (interface{}, error) {
	return plugvols.NewBackupPlugin(logger), nil
}

func newVolumesRestorePlugin(logger logrus.FieldLogger) (interface{}, error) {
	return plugvols.NewRestorePlugin(logger), nil
}
