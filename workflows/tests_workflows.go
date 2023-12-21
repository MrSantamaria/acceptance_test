package workflows

import (
	"fmt"

	"github.com/MrSantamaria/acceptance_test/pkg/openshift/ocm"
	"github.com/MrSantamaria/acceptance_test/pkg/openshift/telemeter"
	"github.com/spf13/viper"
)

/*
1. We will gather a list of clusters that match the clusterDeploymentSelectors - Done
2. We will grab the list of clusterIDs to verify with Telemeter on the csv_succeeded and csv_abnormal
3. We will return a pass/fail depending on csv_succeeded > 0 and csv_abnormal == 0
*/
func AcceptanceTest() error {
	var err error
	testStatus := "PASSED"

	clusterIDs, err := ocm.GetManagementAndServiceClusterIDs()
	if err != nil {
		return err
	}

	clusterExternalIDs, err := ocm.GetExternalIdFromClusterId(clusterIDs)
	if err != nil {
		return err
	}
	fmt.Printf("Cluster External IDs: %+v\n", clusterExternalIDs)

	for _, clusterID := range clusterExternalIDs {

		searchResults, err := telemeter.ObsctlSearchQuery("csv_succeeded{_id=\"" + clusterID + "\",	name=~\"" + viper.GetString("operator") + ".*" + viper.GetString("imagetag") + "\"}[" + viper.GetString("telemeterSearchTime") + "]")
		if err != nil {
			return err
		}
		if telemeter.ObsctlProccessSearchResult(searchResults) < 1 {
			testStatus = "FAILED"
			return fmt.Errorf("csv_succeeded count is 0")
		}

		searchResults, err = telemeter.ObsctlSearchQuery("csv_abnormal{_id=\"" + clusterID + "\", name=~\"" + viper.GetString("operator") + ".*" + viper.GetString("imagetag") + "\"}[" + viper.GetString("telemeterSearchTime") + "]")
		if err != nil {
			return err
		}
		if telemeter.ObsctlProccessSearchResult(searchResults) > 0 {
			testStatus = "FAILED"
			return fmt.Errorf("csv_abnormal count is greater than 0")
		}
	}

	fmt.Printf("Acceptance Test %s for: %s %s environment: %s selectors: %v\n",
		testStatus,
		viper.GetString("operator"),
		viper.GetString("imagetag"),
		viper.GetString("environment"),
		viper.GetStringSlice("selectors"))

	return nil
}
