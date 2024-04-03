package common

import (
	"context"
	"fmt"
	"time"

	hyperv1 "github.com/openshift/hypershift/api/hypershift/v1beta1"
	"github.com/sirupsen/logrus"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/rest"
	cr "sigs.k8s.io/controller-runtime"
	crclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func getMetadataAndAnnotations(item runtime.Unstructured) (metav1.Object, map[string]string, error) {
	metadata, err := meta.Accessor(item)
	if err != nil {
		return nil, nil, err
	}

	annotations := metadata.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}

	return metadata, annotations, nil
}

//func GetClient() (*kubernetes.Clientset, error) {
//	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
//	configOverrides := &clientcmd.ConfigOverrides{}
//	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
//	clientConfig, err := kubeConfig.ClientConfig()
//	if err != nil {
//		return nil, errors.WithStack(err)
//	}
//
//	client, err := kubernetes.NewForConfig(clientConfig)
//	if err != nil {
//		return nil, errors.WithStack(err)
//	}
//
//	return client, nil
//}

// GetClient creates a controller-runtime client for Kubernetes
func GetClient() (crclient.Client, error) {
	config, err := GetConfig()
	if err != nil {
		return nil, fmt.Errorf("unable to get kubernetes config: %w", err)
	}
	client, err := crclient.New(config, crclient.Options{Scheme: CustomScheme})
	if err != nil {
		return nil, fmt.Errorf("unable to get kubernetes client: %w", err)
	}
	return client, nil
}

func GetConfig() (*rest.Config, error) {
	cfg, err := cr.GetConfig()
	if err != nil {
		return nil, err
	}
	cfg.QPS = 100
	cfg.Burst = 100
	return cfg, nil
}

func WaitForPausedPropagated(ctx context.Context, client crclient.Client, log logrus.FieldLogger, hc *hyperv1.HostedCluster) (string, error) {
	waitCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	err := wait.PollUntilContextCancel(waitCtx, 5*time.Second, true, func(ctx context.Context) (bool, error) {
		hcp := &hyperv1.HostedControlPlane{}
		if err := client.Get(ctx, types.NamespacedName{Name: hc.Name, Namespace: hc.Namespace}, hcp); err != nil {
			if apierrors.IsNotFound(err) {
				return true, nil
			}
			log.Error(err, "failed to get HostedControlPlane", "namespace", hc.Namespace, "name", hc.Name)
			return false, err
		}

		if hcp.Spec.PausedUntil != nil {
			log.Debug("HostedControlPlane is paused", "namespace", hc.Namespace, "name", hc.Name)
			return true, nil
		}

		return false, nil
	})

	if err != nil {
		log.Error(err, "Giving up, HCP was not updated in the expecteed timeout", "namespace", hc.Namespace, "name", hc.Name)
		return "", err
	}

	return "", err
}
