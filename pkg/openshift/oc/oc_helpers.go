package ocCli

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/MrSantamaria/acceptance_test/pkg/helpers"
)

type ClusterServiceVersion struct {
	Items []struct {
		Status struct {
			Phase string `json:"phase"`
		} `json:"status"`
		Spec struct {
			Version string `json:"version"`
		} `json:"spec"`
	} `json:"items"`
}

type MatchExpression struct {
	Key      string   `json:"key"`
	Operator string   `json:"operator"`
	Values   []string `json:"values"`
}

type ClusterDeploymentSelector struct {
	MatchExpressions []MatchExpression `json:"matchExpressions"`
}

type SelectorSyncSet struct {
	Spec struct {
		ClusterDeploymentSelector `json:"clusterDeploymentSelector"`
	} `json:"spec"`
}

type ClusterDeployment struct {
	Namespace string `json:"NAMESPACE"`
	Name      string `json:"NAME"`
	Infraid   string `json:"INFRAID"`
	// Add other fields as needed
}

func GetClusterServiceVersionPhaseVersionShortSha(ctx context.Context, operatorName string) (string, string, string, error) {
	cmd := exec.CommandContext(ctx,
		"oc",
		"get",
		"csv",
		"-n",
		"openshift-"+operatorName,
		"-l",
		"operators.coreos.com/"+operatorName+".openshift-"+operatorName+"=",
		"-ojson")

	// Print the command being executed
	fmt.Printf("Executing command: %s\n", cmd.String())

	// Create a buffer to capture the 'oc' command output
	var ocOutput bytes.Buffer
	cmd.Stdout = &ocOutput
	var ocStderr bytes.Buffer
	cmd.Stderr = &ocStderr

	// Run the 'oc' command and capture its output
	if err := cmd.Run(); err != nil {
		return "", "", "", fmt.Errorf("error executing the 'oc' command: %v\nStandard Error: %s", err, ocStderr.String())
	}

	// Parse the JSON output directly
	var csv ClusterServiceVersion
	if err := json.Unmarshal(ocOutput.Bytes(), &csv); err != nil {
		return "", "", "", fmt.Errorf("failed to parse JSON: %v", err)
	}

	if len(csv.Items) == 0 {
		return "", "", "", fmt.Errorf("no ClusterServiceVersion found")
	}

	version := strings.Split(csv.Items[0].Spec.Version, "-")

	// Return the phase of the first ClusterServiceVersion
	return csv.Items[0].Status.Phase, version[1], "", nil
}

func IsClusterConnected() bool {
	cmd := exec.Command("oc", "whoami")

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return false
	}

	fmt.Println("Cluster connection is active")
	return true
}

func PrintNamespaces() error {
	cmd := exec.Command("oc", "get", "namespaces")

	// Redirect the command output to the current process's standard output
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Run the 'oc get namespaces' command
	err := cmd.Run()

	// Check for errors and return them, if any
	if err != nil {
		return fmt.Errorf("error executing oc get namespaces: %v", err)
	}

	return nil
}

func GetSelectorSyncSetLabels(ctx context.Context, sss string) (string, error) {
	cmd := exec.Command(
		"oc",
		"get",
		"selectorsyncset",
		sss,
		"-ojson")

	// Capture the command output
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("error executing oc get selectorsyncset: %v", err)
	}

	var sssData SelectorSyncSet

	err = json.Unmarshal(output, &sssData)
	if err != nil {
		return "", fmt.Errorf("error parsing JSON: %v", err)
	}

	var labelLines []string
	for _, expr := range sssData.Spec.ClusterDeploymentSelector.MatchExpressions {
		// Build the label line for each match expression
		line := fmt.Sprintf("%s %s (%s)", expr.Key, expr.Operator, strings.Join(expr.Values, ","))
		labelLines = append(labelLines, line)
	}

	// Join the label lines with a comma separator
	labelLine := strings.Join(labelLines, ", ")

	return labelLine, nil
}

func GetClusterDeployments(ctx context.Context, labels string) ([]string, error) {
	cmd := exec.Command(
		"oc",
		"get",
		"clusterdeployments",
		"--all-namespaces",
		"-l",
		labels,
	)

	// Capture the command output
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("error executing oc get clusterdeployments: %v", err)
	}

	// Parse the Comma Separated Values (CSV) output
	r := csv.NewReader(strings.NewReader(string(output)))
	records, err := r.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("error parsing CSV: %v", err)
	}

	var clusterDeployments []string
	for _, record := range records[1:] { // Skip the header row
		clusterDeployment := strings.TrimSpace(record[0])
		clusterDeployments = append(clusterDeployments, clusterDeployment)
	}

	return clusterDeployments, nil
}

func CliCheck() bool {
	cmd := exec.Command("oc")
	err := cmd.Run()
	if err == nil {
		return true
	}

	fmt.Println("oc is not installed, attempting to install via Homebrew")
	fmt.Println("https://docs.openshift.com/container-platform/4.8/cli_reference/openshift_cli/getting-started-cli.html#cli-installing-cli-brew_cli-developer-commands")

	err = helpers.InstallFromHomeBrew("openshift-cli")

	return err == nil
}
