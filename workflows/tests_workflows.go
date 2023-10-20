package workflows

import (
	"context"
	"fmt"

	"github.com/MrSantamaria/acceptance_test/pkg/openshift/hive"
	"github.com/MrSantamaria/acceptance_test/pkg/openshift/oc"
	"github.com/MrSantamaria/acceptance_test/pkg/openshift/ocm"
	v1 "github.com/openshift/hive/apis/hive/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func AcceptanceTest(token, environment, namespace, operator, pairedSSS, imageTag string) error {
	failedClusters := []string{}
	sss := &v1.SelectorSyncSet{}
	clusterDeploymentsList := &v1.ClusterDeploymentList{}
	//clusterDeployments := &v1.ClusterDeployment{}

	hiveClient, err := hive.GetHiveClient()
	if err != nil {
		fmt.Println("Error:", err)
		return err
	}

	err = hiveClient.Get(context.TODO(), client.ObjectKey{Namespace: namespace, Name: pairedSSS}, sss)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("--- SelectorSyncSet.ClusterDeploymentSelector.MatchExpressions ---")
	fmt.Println(sss.Spec.ClusterDeploymentSelector.MatchExpressions)

	labelSelectors, err := metav1.LabelSelectorAsSelector(&sss.Spec.ClusterDeploymentSelector)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("--- labelSelectors ---")
	fmt.Println(labelSelectors)

	err = hiveClient.List(context.TODO(), clusterDeploymentsList, &client.ListOptions{LabelSelector: labelSelectors})
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("--- ClusterDeployments ---")
	fmt.Println(clusterDeploymentsList.Items)

	for _, clusterDeployment := range clusterDeploymentsList.Items {
		fmt.Println("--- ClusterDeployment ---")
		fmt.Println(clusterDeployment.Name)
		err = ocm.BackplaneLogin(clusterDeployment.Name)
		if err != nil {
			fmt.Println("Error:", err)
			return err
		}

		if !oc.IsClusterConnected() {
			fmt.Println("Error: Cluster is not connected. Cluster is considered failed.")
			failedClusters = append(failedClusters, clusterDeployment.Name)
			continue
		}

		phase, version, _, err := oc.GetClusterServiceVersionPhaseVersionShortSha(operator)
		if err != nil {
			fmt.Println("Error:", err)
			failedClusters = append(failedClusters, clusterDeployment.Name)
			continue
		}

		if phase == "Succeeded" && version == imageTag {
			fmt.Printf("Cluster %s has passed acceptance testing\n", clusterDeployment.Name)
		} else {
			failedClusters = append(failedClusters, clusterDeployment.Name)
			continue
		}
	}

	if len(failedClusters) > 0 {
		return fmt.Errorf("the following clusters have failed: %v", failedClusters)
	}

	return nil
}
