package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/MrSantamaria/acceptance_test/pkg/helpers"
	ocCli "github.com/MrSantamaria/acceptance_test/pkg/openshift/oc"
	"github.com/MrSantamaria/acceptance_test/pkg/openshift/ocm"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// Trigger quay build
var (
	repoOwner      = "openshift-online"
	repoName       = "ocm-cli"
	imageTag       = "e3cc340"            //This image tag is the latest image tag for the aws-vpce-operator matching app-interface.
	ctx            = context.Background() // Ctx as a placeholder this is meant for the test_main class once we are ready to test.
	pairedSSS      = "aws-vpce-operator-hypershift-sss-us-west-2-main"
	failedClusters []string
)

// There is a missing part I want to request from APP-SRE and that is the comsumption of promotion data.
/*
	In the interim you can assume that we will be passing the following data to the test:
	$OPERATOR_NAME-$ENV-$REGION
	IE: aws-vpce-operator-integration-us-east-2
	With this we should be able to query the region but not the sector as that is part of the passed parameters.
	We could pass this as part of the parameters but preferably consuming from the published data would be better.
	Ref: https://gitlab.cee.redhat.com/service/app-interface/-/blob/master/data/services/osd-operators/cicd/saas/saas-aws-vpce-operator/hypershift-deploy.yaml#L47
	Ref Thread: https://redhat-internal.slack.com/archives/CCRND57FW/p1690576939634289

*/

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Error: Missing arguments")
		os.Exit(1)
	}

	ocmToken := os.Args[1]

	// This first part will be used to set up the environment for the test
	// Everything will run from a new directory created in the system's temporary directory
	// We will download OCM and OC binaries to this directory and use them for the test

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

	// Call the inClusterConfig function to set up the oc-cli
	config, err := rest.InClusterConfig()
	if err != nil {
		fmt.Println("InClusterConfig failed")
		fmt.Println("Error:", err)
		panic(err.Error())
	}
	fmt.Println("InClusterConfig set up successfully")

	// creates the clientset
	_, err = kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Printf("Clientset failed, err: %v\n", err)
		panic(err.Error())
	}
	fmt.Println("Clientset set up successfully")

	// Set the KUBECONFIG environment variable
	os.Setenv("KUBECONFIG", config.BearerTokenFile)
	fmt.Printf("KUBECONFIG set to %s\n", config.BearerTokenFile)

	// Moved this into a function underneath just to not lose the progress.
	//acceptance_test_through_ocm(ctx, ocmBinaryPath)

	// This next Approach uses the labelSelectors from inside Hive to get the clusters
	// Ref: https://redhat-internal.slack.com/archives/CMK13BP4J/p1690569091850269?thread_ts=1690564815.068249&cid=CMK13BP4J
	// In order for this approach to work the script will be need to be aware of app-interface.
	// We can either attach a component to query app-interface, this shouldn't be a problem since the acceptance test
	// needs to run behind the VPN in order to use backplane.
	// I will be developing a solution that gets args passed in to accomplish the same thing in hopes of consumed promotion data.

	acceptance_test_through_hive(ctx, ocmBinaryPath)

	// This line exist for a debug breakpoint - Diego
	fmt.Println("Hello, world!")
}

func acceptance_test_through_hive(ctx context.Context, ocmBinaryPath string) {

	cmd := exec.Command(
		"oc",
		"get",
		"nodes")

	// Capture the command output
	output, err := cmd.Output()
	if err != nil {
		fmt.Println("Error:", err)
	}

	fmt.Println(string(output))

	// These commands run in the context of the cluster that is running the acceptance test.
	// For Dustin - Since this runs in a Hive in production can we load a kubeconfig unto the pod through a configMap?
	// If we can't load a kubeconfig how are we going to run the oc commands? - Diego
	labels, err := ocCli.GetSelectorSyncSetLabels(ctx, pairedSSS)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	clusterDeployments, err := ocCli.GetClusterDeployments(ctx, labels)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	for _, clusterDeployment := range clusterDeployments {
		fields := strings.Fields(clusterDeployment)
		if len(fields) > 2 {
			clusterName := fields[1]
			fmt.Println(clusterName)

			err = ocm.BackplaneLogin(ctx, ocmBinaryPath, clusterName)
			if err != nil {
				fmt.Println("Error:", err)
				continue
			}

			if !ocCli.IsClusterConnected() {
				fmt.Println("Error:", err)
				continue
			}

			phase, version, _, err := ocCli.GetClusterServiceVersionPhaseVersionShortSha(ctx, clusterName, "aws-vpce-operator")
			if err != nil {
				fmt.Println("Error:", err)
				continue
			}

			if phase == "Succeeded" && version == imageTag {
				fmt.Println("Cluster is ready")
			} else {
				failedClusters = append(failedClusters, clusterName)
			}

			if len(failedClusters) > 0 {
				fmt.Println("Failed Clusters:")
				for _, cluster := range failedClusters {
					fmt.Println(cluster)
				}
			} else {
				fmt.Println("All clusters are ready")
			}

		}
	}
}

func acceptance_test_through_ocm(ctx context.Context, ocmBinaryPath string) {
	managementClusters, serviceClusters, err := ocm.GetManagementAndServiceClusters(ctx, ocmBinaryPath, "us-west-2", "")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Print the management clusters
	fmt.Println("Management Clusters:")
	for _, cluster := range managementClusters {
		fmt.Println(cluster)
		err = ocm.BackplaneLogin(ctx, ocmBinaryPath, cluster)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		if !ocCli.IsClusterConnected() {
			fmt.Println("Error:", err)
			return
		}

		phase, version, _, err := ocCli.GetClusterServiceVersionPhaseVersionShortSha(ctx, cluster, "aws-vpce-operator")
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		if phase == "Succeeded" && version == imageTag {
			fmt.Println("Cluster is ready")
		} else {
			failedClusters = append(failedClusters, cluster)
		}
	}

	fmt.Println("Service Clusters:")
	// Print the service clusters
	for _, cluster := range serviceClusters {
		fmt.Println(cluster)
		err = ocm.BackplaneLogin(ctx, ocmBinaryPath, cluster)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		if !ocCli.IsClusterConnected() {
			fmt.Println("Error:", err)
			return
		}

		phase, version, _, err := ocCli.GetClusterServiceVersionPhaseVersionShortSha(ctx, cluster, "aws-vpce-operator")
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		if phase == "Succeeded" && version == imageTag {
			fmt.Println("Cluster is ready")
		} else {
			failedClusters = append(failedClusters, cluster)
		}
	}

	if len(failedClusters) > 0 {
		fmt.Println("Failed Clusters:")
		for _, cluster := range failedClusters {
			fmt.Println(cluster)
		}
	} else {
		fmt.Println("All clusters are ready")
	}
}
