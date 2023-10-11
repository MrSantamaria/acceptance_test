package main

import (
	"context"
	"fmt"

	"github.com/openshift/hive/apis"
	v1 "github.com/openshift/hive/apis/hive/v1"
	"k8s.io/apimachinery/pkg/runtime"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func main() {
	//kubeconfig := "/Users/dsantama/Documents/GitHub/acceptance_test/kubeconfig"
	// Create a new scheme and add necessary APIs to it.
	scheme := runtime.NewScheme()
	if err := apis.AddToScheme(scheme); err != nil {
		fmt.Printf("Error adding APIs to scheme: %v\n", err)
		return
	}

	// Get the Kubernetes client using in-cluster configuration.
	customClient, err := GetClient(scheme) // Pass the scheme to GetClient function
	if err != nil {
		fmt.Printf("Error creating dynamic client: %v\n", err)
		return
	}

	sss := &v1.SelectorSyncSet{}

	err = customClient.Get(context.TODO(), client.ObjectKey{Namespace: "cluster-scope", Name: "aws-vpce-operator-hypershift-sss-us-west-2-main"}, sss)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("--- SelectorSyncSet ---")
	fmt.Println(sss)

	fmt.Println("--- SelectorSyncSet.ClusterDeploymentSelector.MatchLabels ---")
	fmt.Println(sss.Spec.ClusterDeploymentSelector.MatchLabels)

	fmt.Println("--- SelectorSyncSet.ClusterDeploymentSelector.MatchExpressions ---")
	fmt.Println(sss.Spec.ClusterDeploymentSelector.MatchExpressions)

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

// GetClientConfig gets the config for the REST client.
func GetClientConfig() (*restclient.Config, error) {
	rules := clientcmd.NewDefaultClientConfigLoadingRules()
	kubeconfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules, &clientcmd.ConfigOverrides{})
	return kubeconfig.ClientConfig()
}

// Create a mock object that satisfies the v1.SelectorSyncSet interface
type MockSelectorSyncSet struct {
	// Define fields that match the structure of v1.SelectorSyncSet
	Field1 string
	Field2 int
	// Add other fields as needed
}
