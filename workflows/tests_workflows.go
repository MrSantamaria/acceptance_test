package workflows

import (
	"fmt"

	"github.com/MrSantamaria/acceptance_test/pkg/openshift/ocm"
	"github.com/MrSantamaria/acceptance_test/pkg/openshift/telemeter"
	"github.com/spf13/viper"
)

func AcceptanceTest() error {
	var err error

	clusterIDs, err := ocm.GetManagementAndServiceClusterIDs()
	if err != nil {
		return err
	}

	// Print the clusterIDs
	fmt.Printf("ClusterIDs: %+v\n", clusterIDs)
	clusterOperationIDs, err := ocm.TransformClusterIDsToOperationIDs(clusterIDs)
	if err != nil {
		return err
	}

	/*
		1. We will gather a list of clusters that match the clusterDeploymentSelectors - Done
		2. We will grab the list of clusterIDs to verify with Telemeter on the csv_succeeded and csv_abnormal
		3. We will return a pass/fail depending on csv_succeeded > 0 and csv_abnormal == 0
	*/

	searchResults, err := telemeter.ObsctlSearchQuery("csv_succeeded{name=~\"" + viper.GetString("operator") + ".*" + viper.GetString("imagetag") + "\"}[" + viper.GetString("telemeterSearchTime") + "]")
	if err != nil {
		return err
	}
	if telemeter.ObsctlProccessSearchResult(searchResults, clusterOperationIDs) < 1 {
		return fmt.Errorf("csv_succeeded count is 0")
	}

	searchResults, err = telemeter.ObsctlSearchQuery("csv_abnormal{name=~\"" + viper.GetString("operator") + ".*" + viper.GetString("imagetag") + "\"}[" + viper.GetString("telemeterSearchTime") + "]")
	if err != nil {
		return err
	}
	if telemeter.ObsctlProccessSearchResult(searchResults, clusterIDs) > 0 {
		return fmt.Errorf("csv_abnormal count is greater than 0")
	}

	return nil
}
