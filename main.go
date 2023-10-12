package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/MrSantamaria/acceptance_test/pkg/helpers"
	ocCli "github.com/MrSantamaria/acceptance_test/pkg/openshift/oc"
	"github.com/MrSantamaria/acceptance_test/pkg/openshift/ocm"
	"github.com/openshift/hive/apis"
	v1 "github.com/openshift/hive/apis/hive/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	repoOwner = "openshift-online"
	repoName  = "ocm-cli"
	imageTag  = "c599108"            //This image tag is the latest image tag for the aws-vpce-operator matching app-interface.
	ctx       = context.Background() // Ctx as a placeholder this is meant for the test_main class once we are ready to test.
	//pairedSSS      = "aws-vpce-operator-hypershift-sss-us-west-2-main"
	//namespace      = "cluster-scope"
	failedClusters []string
)

func main() {
	// Grab the token from the environment
	ocmToken := os.Getenv("OCM_TOKEN")
	//imageTag := os.Getenv("IMAGE_TAG")

	// All the old code to set up OCM and OC
	tmpDir, err := os.MkdirTemp("", "")
	if err != nil {
		fmt.Printf("failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Change the working directory to the temporary directory
	err = os.Chdir(tmpDir)
	if err != nil {
		fmt.Println("Error changing working directory:", err)
		return
	}

	// From here we will download the needed binaries for our execution

	// Get the latest release of the ocm-cli
	release, err := helpers.GetLatestRelease(repoOwner, repoName)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	selectedURL, err := helpers.SelectVersionByRuntime(release)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Extract the file name from the URL to use as the local file name
	tokens := strings.Split(selectedURL, "/")
	fileName := tokens[len(tokens)-1]

	ocmBinaryPath, err := helpers.DownloadRelease(selectedURL, fileName)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	err = helpers.SetupBinary(ocmBinaryPath)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	err = helpers.CheckBinary(ocmBinaryPath)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	err = ocm.Login(ctx, ocmBinaryPath, ocmToken, "int")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Get the latest release of the oc-cli
	if !ocCli.CliCheck() {
		return
	}

	err = helpers.CheckBinary("oc")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

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
	clusterDeploymentsList := &v1.ClusterDeploymentList{}
	//clusterDeployments := &v1.ClusterDeployment{}

	err = customClient.Get(context.TODO(), client.ObjectKey{Namespace: "cluster-scope", Name: "aws-vpce-operator-hypershift-sss-us-west-2-main"}, sss)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("--- SelectorSyncSet.ClusterDeploymentSelector.MatchExpressions ---")
	fmt.Println(sss.Spec.ClusterDeploymentSelector.MatchExpressions)

	labelSelectors, err := metav1.LabelSelectorAsSelector(&sss.Spec.ClusterDeploymentSelector)
	if err != nil {
		fmt.Println(err)
	}

	//Print out the labelSelectors
	fmt.Println("--- labelSelectors ---")
	fmt.Println(labelSelectors)

	err = customClient.List(context.TODO(), clusterDeploymentsList, &client.ListOptions{LabelSelector: labelSelectors})
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("--- ClusterDeploymentList.Items ---")
	fmt.Println(clusterDeploymentsList.Items)

	for _, clusterDeployment := range clusterDeploymentsList.Items {
		fmt.Println("--- ClusterDeployment ---")
		fmt.Println(clusterDeployment.Name)
		err = ocm.BackplaneLogin(ctx, ocmBinaryPath, clusterDeployment.Name)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		if !ocCli.IsClusterConnected() {
			fmt.Println("Error:", err)
			return
		}

		phase, version, _, err := ocCli.GetClusterServiceVersionPhaseVersionShortSha(ctx, "aws-vpce-operator")
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		if phase == "Succeeded" && version == imageTag {
			fmt.Println("Cluster is ready")
		} else {
			failedClusters = append(failedClusters, clusterDeployment.Name)
		}
	}

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
