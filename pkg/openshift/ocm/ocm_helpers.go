package ocm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/MrSantamaria/acceptance_test/pkg/assets"
	"github.com/MrSantamaria/acceptance_test/pkg/helpers"
	ocmsdk "github.com/openshift-online/ocm-sdk-go"
	"github.com/spf13/viper"
)

var (
	Ocm *ocmClient
	env map[string]string = map[string]string{
		"int":   "https://api.integration.openshift.com",
		"stage": "https://api.stage.openshift.com",
		"prod":  "https://api.openshift.com"}
	backplaneConfig map[string]string = map[string]string{
		"int":   "config.int.json",
		"stage": "config.stage.json",
		"prod":  "config.prod.json"}
)

// OCM is a wrapper around the OCM client.
type ocmClient struct {
	*ocmsdk.Connection
}

func CliCheck() bool {
	cmd := exec.Command("ocm")
	err := cmd.Run()
	if err == nil {
		return true
	}

	fmt.Println("ocm cli is not configured.")
	fmt.Println("This application requires the ocm cli to be installed for the backplane login to work.")
	fmt.Println("https://github.com/openshift-online/ocm-cli")

	return false
}

func Login(token string, environment string) error {
	// Check if the token is empty
	if token == "" {
		return fmt.Errorf("token cannot be empty")
	}

	// Check if the specified environment is valid
	if _, ok := env[environment]; !ok {
		return fmt.Errorf("env %s is not a valid environment", environment)
	}

	// Create the config file needed for backplane operations
	backplaneFile, err := helpers.CopyFileToCurrentDir(assets.Assets, backplaneConfig[environment])
	if err != nil {
		return err
	}

	// Set the BACKPLANE_CONFIG environment variable
	helpers.SetEnvVariables(fmt.Sprintf("BACKPLANE_CONFIG:%s", backplaneFile))

	// Prepare the command
	cmd := exec.Command("ocm", "login", "--token", token, "--url", env[environment])

	// Create a buffer to capture the standard output and standard error of the command
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Execute the command
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("error executing ocm login using token: %v\nStandard Error: %s", err, stderr.String())
	}

	// If the command execution is successful, you can access the captured output as:
	// output := stdout.String()

	return nil
}

func GetManagementAndServiceClusterIDs() ([]string, error) {
	var stdoutManagement, stderrManagement, stdoutService, stderrService bytes.Buffer
	var clusterIDs []string

	// Command for management clusters
	cmdManagement := exec.Command("ocm", "get", "/api/osd_fleet_mgmt/v1/management_clusters")
	cmdManagement.Stdout = &stdoutManagement
	cmdManagement.Stderr = &stderrManagement

	err := cmdManagement.Run()
	if err != nil {
		return nil, fmt.Errorf("error executing ocm get management clusters: %v\nStandard Error: %s", err, stderrManagement.String())
	}

	idsManagement, err := parseJsonDataForClusterIDs(stdoutManagement.String(), "ManagementCluster")
	if err != nil {
		return nil, err
	}

	clusterIDs = append(clusterIDs, idsManagement...)

	// Command for service clusters
	cmdService := exec.Command("ocm", "get", "/api/osd_fleet_mgmt/v1/service_clusters")
	cmdService.Stdout = &stdoutService
	cmdService.Stderr = &stderrService

	err = cmdService.Run()
	if err != nil {
		return nil, fmt.Errorf("error executing ocm get service clusters: %v\nStandard Error: %s", err, stderrService.String())
	}

	idsService, err := parseJsonDataForClusterIDs(stdoutService.String(), "ServiceCluster")
	if err != nil {
		return nil, err
	}

	clusterIDs = append(clusterIDs, idsService...)

	return clusterIDs, nil
}

func TransformClusterIDsToOperationIDs(clusterIDs []string) ([]string, error) {
	var operationIDs []string

	for _, clusterID := range clusterIDs {
		operationID, err := getOperationIdFromClusterId(clusterID)
		if err != nil {
			return nil, err
		}

		operationIDs = append(operationIDs, operationID)
	}

	return operationIDs, nil
}

func parseJsonDataForClusterIDs(jsonData, clusterKind string) ([]string, error) {
	var data map[string]interface{}
	var clusterIDs []string
	regionSelector, sectorSelector, err := createOCMSelectors(viper.GetStringSlice("selectors"))
	if err != nil {
		return nil, err
	}

	err = json.NewDecoder(strings.NewReader(jsonData)).Decode(&data)
	if err != nil {
		return nil, err
	}

	items, ok := data["items"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("error accessing 'items' array")
	}

	for _, item := range items {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			fmt.Println("Error accessing item in 'items' array")
			continue
		}

		region, regionExists := itemMap["region"].(string)
		if region != regionSelector && regionExists {
			continue
		}
		sector, sectorExists := itemMap["sector"].(string)
		if sector != sectorSelector && sectorExists {
			continue
		}
		kind, kindExists := itemMap["kind"].(string)
		if kind != clusterKind && kindExists {
			continue
		}

		clusterIDs = append(clusterIDs, itemMap["id"].(string))
	}

	return clusterIDs, nil

}

// Create OCM selectors based on AWS regions and Openshift sectors
func createOCMSelectors(selectors []string) (string, string, error) {
	var regionSelector, sectorSelector string

	for _, selector := range selectors {
		trimmedSelector := strings.TrimSpace(selector)
		if helpers.IsAWSRegion(trimmedSelector) {
			regionSelector = trimmedSelector
			continue
		}
		if helpers.IsOpenshiftSector(trimmedSelector) {
			sectorSelector = trimmedSelector
			continue
		}

		fmt.Printf("Selector %s is not a valid AWS region or Openshift sector\nWe will ignore it for now. Reach out the SD-CICADA team if this shouldn't be the case\n", trimmedSelector)
	}

	if regionSelector == "" || sectorSelector == "" {
		return "", "", fmt.Errorf("error creating OCM selectors")
	}

	return regionSelector, sectorSelector, nil
}

func getOperationIdFromClusterId(clusterId string) (string, error) {
	var stdout, stderr bytes.Buffer

	cmd := exec.Command("ocm", "get", "cluster", clusterId)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("error executing ocm get cluster: %v\nStandard Error: %s", err, stderr.String())
	}

	// operationId, err := stdout.ReadString(
	// if err != nil {
	// 	return "", err
	// }

	operationId := strings.TrimSpace(stdout.String())

	return operationId, nil
}
