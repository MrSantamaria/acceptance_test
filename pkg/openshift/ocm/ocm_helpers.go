package ocm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"
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

// Cluster represents the structure of the JSON data
type Cluster struct {
	Kind  string `json:"kind"`
	Page  int    `json:"page"`
	Size  int    `json:"size"`
	Total int    `json:"total"`
	Items []Item `json:"items"`
}

// Item represents the structure of the "items" array in the JSON data
type Item struct {
	ID                         string                     `json:"id"`
	Kind                       string                     `json:"kind"`
	Href                       string                     `json:"href"`
	Name                       string                     `json:"name"`
	Status                     string                     `json:"status"`
	CloudProvider              string                     `json:"cloud_provider"`
	Region                     string                     `json:"region"`
	Sector                     string                     `json:"sector"` // Assuming you have "sector" field in your JSON
	ClusterManagementReference ClusterManagementReference `json:"cluster_management_reference"`
}

// ClusterManagementReference represents the structure of the "cluster_management_reference" field in the JSON data
type ClusterManagementReference struct {
	ClusterID string `json:"cluster_id"`
	Href      string `json:"href"`
}

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

	backplaneFile, err := helpers.CopyFileToCurrentDir(assets.Assets, backplaneConfig[environment])
	if err != nil {
		return err
	}

	helpers.SetEnvVariables(fmt.Sprintf("BACKPLANE_CONFIG:%s", backplaneFile))

	cmd := exec.Command("ocm", "login", "--token", token, "--url", env[environment])

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	fmt.Printf("Logging in to OCM for %s environment\n", environment)
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

	fmt.Printf("Running command: %s\n", cmdManagement.Args)
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

	fmt.Printf("Running command: %s\n", cmdService.Args)
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

func GetExternalIdFromClusterId(clusterIds []string) ([]string, error) {
	// The stdout of the ocm describe cluster $id command does not output a json, we will use a regex to parse the output
	var stdout, stderr bytes.Buffer
	var clusterExternalIds []string
	re := regexp.MustCompile(`External ID:\s+(\S+)`)

	for _, id := range clusterIds {
		cmd := exec.Command("ocm", "describe", "cluster", id)
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		fmt.Printf("Running command: %s\n", cmd.Args)
		err := cmd.Run()
		if err != nil {
			return clusterExternalIds, fmt.Errorf("error executing ocm describe cluster: %v\nStandard Error: %s", err, stderr.String())
		}

		match := re.FindStringSubmatch(stdout.String())
		// Check if there is a line matching the regex
		// If there is a match 'External ID' will be in match[0] and the external id will be in match[1]
		if len(match) == 2 {
			externalID := match[1]
			clusterExternalIds = append(clusterExternalIds, externalID)
		} else {
			fmt.Println("External ID not found.")
		}

		// Resets the buffer otherwise the next iteration will append to the previous stds
		stdout.Reset()
		stderr.Reset()
	}

	return clusterExternalIds, nil
}

func parseJsonDataForClusterIDs(jsonData, clusterKind string) ([]string, error) {
	var cluster Cluster
	var clusterIDs []string

	regionSelector, sectorSelector, err := createOCMSelectors(viper.GetStringSlice("selectors"))
	if err != nil {
		return nil, err
	}

	err = json.NewDecoder(strings.NewReader(jsonData)).Decode(&cluster)
	if err != nil {
		return nil, err
	}

	for _, item := range cluster.Items {
		if item.Region != regionSelector {
			continue
		}

		if item.Sector != sectorSelector {
			continue
		}

		if item.Kind != clusterKind {
			continue
		}

		clusterIDs = append(clusterIDs, item.ClusterManagementReference.ClusterID)
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
