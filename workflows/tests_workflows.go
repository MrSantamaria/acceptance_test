package workflows

import (
	"fmt"

	"github.com/MrSantamaria/acceptance_test/pkg/openshift/ocm"
)

func AcceptanceTest() error {
	var err error

	clusterIDs, err := ocm.GetManagementAndServiceClusterIDs()
	if err != nil {
		return err
	}

	// All clusters on [0] are management clusters
	// All clusters on [1] are service clusters
	// Print the clusterIDs
	fmt.Printf("ClusterIDs: %+v\n", clusterIDs)

	/*
		1. We will gather a list of clusters that match the clusterDeploymentSelectors - Done
		2. We will grab the list of clusterIDs to verify with Telemeter on the csv_succeeded and csv_abnormal
		3. We will return a pass/fail depending on csv_succeeded > 0 and csv_abnormal == 0
	*/

	return nil
}
