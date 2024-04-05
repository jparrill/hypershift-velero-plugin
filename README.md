# hypershift-velero-plugin
Hosted Control Planes Velero plugin to cover backup and restore of the Control Planes objects

## Pre-reqs

- Have MCE or Hypershift Operator working with at least 1 HostedCluster deployed
- S3 bucket to store the VolumeSnapshot and the Openshift Objects

## Quickstart

1. Apply this manifest in a Connected Management Cluster

```yaml
---
apiVersion: v1
kind: Namespace
metadata:
  name: openshift-adp
  labels:
    openshift.io/cluster-monitoring: "true"
  annotations:
    workload.openshift.io/allowed: management
---
apiVersion: operators.coreos.com/v1
kind: OperatorGroup
metadata:
  name: redhat-oadp-operator-operatorgroup
  namespace: openshift-adp
spec:
  targetNamespaces:
  - openshift-adp
---
apiVersion: operators.coreos.com/v1alpha1
kind: Subscription
metadata:
  name: redhat-oadp-operator
  namespace: openshift-adp
spec:
  channel: "stable-1.3"
  name: redhat-oadp-operator
  source: redhat-operators
  sourceNamespace: openshift-marketplace
  startingCSV: oadp-operator.v1.3.0
  installPlanApproval: Automatic
```

This will deploy the OADP Operator.

2. Now we need to create the `DataProtectionApplication` which is the instance that will deploy Velero and node agents.

```yaml
---
apiVersion: oadp.openshift.io/v1alpha1
kind: DataProtectionApplication
metadata:
  name: dpa-instance
  namespace: openshift-adp
spec:
  backupLocations:
    - name: default
      velero:
        provider: aws
        default: true
        objectStorage:
          bucket: jparrill-oadp
          prefix: objects
        config:
          region: us-west-2
          profile: "default"
        credential:
          key: cloud
          name: cloud-credentials
  snapshotLocations:
    - velero:
        provider: aws
        config:
          region: us-west-2
          profile: "default"
        credential:
          key: cloud
          name: cloud-credentials
  configuration:
    nodeAgent:
      enable: true
      uploaderType: kopia
    velero:
      defaultPlugins:
        - openshift
        - aws
        - csi
      customPlugins:
        - name: hypershift-velero-plugin
          image: quay.io/jparrill/hypershift-velero-plugin:main
      resourceTimeout: 2h
```

You can change the provider but the customPlugin part is necessary to use the code we've don in this repo.

Once created this CR, some pods will pop up in the `openshift-adp` namespace which implies:

- The node agents for VolumeSnapshots
- The velero main pod which it will mount your new plugin as a side container
- And the openshift-adp-controller-manager

3. Now we can create our main objects which are `backup` and `restore` ones.

```yaml
---
apiVersion: velero.io/v1
kind: Backup
metadata:
  name: csi-hc-backup
  namespace: openshift-adp
  labels:
    velero.io/storage-location: default
spec:
  hooks: {}
  includedNamespaces:
  - clusters
  - clusters-jparrill-hosted
  includedResources:
  - pv
  - pvc
  - hostedcluster
  - nodepool
  - secrets
  - hostedcontrolplane
  - cluster
  - awscluster
  - awsmachinetemplate
  - awsmachine
  - machinedeployment
  - machineset
  - machine
  excludedResources: []
  storageLocation: default
  ttl: 720h0m0s
  snapshotMoveData: true
  datamover: "velero"
```

The deployment of this new object will trigger the backup process creating other multiple objects. You can check the logs in the velero pod and see how your/this plugin modifies the natural flow of ADP backup.

4. Check the backup process

```bash
oc get backup -n openshift-adp -ojsonpath='{items[0].status}'
```

The backup object is stored in the `openshift-adp` namespace and you can check also the resources:

- **VolumeSnapshot:** It's the CR in charge to create the ETCD Backup and store it in the `BackupStorageLocations` (another CR)
- **DataUpload:** It's the CR in charge to upload a successfully VolumeSnapshot created, meanwhile the backup CR contains the parameter `snapshotMoveData: true` and `datamover: "velero"`. The `status.phase` field could tell you how it's progressing.
- **BackupRepository:** It's another CR created by default which takes care of the S3 bucket folder structure. Sometimes it gets blocked if you modify the S3 bucket folder structure using the browser.

5. Advices

- To handle the backup deletion I recommend to use velero CLI or use the container with this alias:

```bash
alias velero='oc -n openshift-adp exec deployment/velero -c velero -it -- ./velero'
velero backup delete --all
```

These commands will make sure the BackupRepository does not get blocked during you back and forth retries during the development.

- Make sure you've followed the official doc to configure the necessary AWS and OCP objects to make velero work like IAM, Poolicy file, etc... For more info you can check the [official doc](https://docs.openshift.com/container-platform/4.15/backup_and_restore/application_backup_and_restore/installing/installing-oadp-aws.html) or this [small guide](https://github.com/jparrill/poc-oadp-ho/blob/main/assets/oadp/README.md)

## Development

### Important make targets

- `make`: Compiles the plugin in local
- `make docker-build`: Creates the container and tag it
- `make docker-push`: Pushes the container to quay.io

### Dev Flow (without Tilt)

- Change the code
- `make docker-build docker-push`
- `oc delete pod -n openshift-adp -ldeploy=velero`
- If you need to clean the previous backup: `velero backup delete  --all`
- Apply the `backup` CR again

