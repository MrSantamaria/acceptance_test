package hive

import (
	"fmt"

	"github.com/openshift/hive/apis"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// GetClientConfig gets the config for the REST client.
func GetClientConfig() (*rest.Config, error) {
	rules := clientcmd.NewDefaultClientConfigLoadingRules()
	kubeconfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules, &clientcmd.ConfigOverrides{})
	return kubeconfig.ClientConfig()
}

// GetClient returns a new dynamic controller-runtime client.
func GetClient(scheme *runtime.Scheme) (client.Client, error) {
	cfg, err := GetClientConfig()
	if err != nil {
		return nil, err
	}

	// Create a dynamic controller-runtime client.
	dynamicClient, err := client.New(cfg, client.Options{Scheme: scheme})
	if err != nil {
		return nil, err
	}

	return dynamicClient, nil
}

func GetHiveClient() (client.Client, error) {
	scheme := runtime.NewScheme()
	if err := apis.AddToScheme(scheme); err != nil {
		fmt.Printf("Error adding APIs to scheme: %v\n", err)
		return nil, err
	}

	// Get the Kubernetes client using in-cluster configuration.
	HiveClient, err := GetClient(scheme) // Pass the scheme to GetClient function
	if err != nil {
		fmt.Printf("Error creating dynamic client: %v\n", err)
		return nil, err
	}

	return HiveClient, nil
}
