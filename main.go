package main

import (
	plugcommon "github.com/jparrill/hypershift-velero-plugin/internal/common"
	"github.com/sirupsen/logrus"
	"github.com/vmware-tanzu/velero/pkg/plugin/framework"
)

func main() {
	framework.NewServer().
		RegisterBackupItemAction("openshift.io/hypershift-common-backup-plugin", newCommonBackupPlugin).
		RegisterRestoreItemAction("openshift.io/hypershift-common-restore-plugin", newCommonRestorePlugin).
		Serve()
}

func newCommonBackupPlugin(logger logrus.FieldLogger) (interface{}, error) {
	return plugcommon.NewBackupPlugin(logger), nil
}

func newCommonRestorePlugin(logger logrus.FieldLogger) (interface{}, error) {
	return plugcommon.NewRestorePlugin(logger), nil
}
