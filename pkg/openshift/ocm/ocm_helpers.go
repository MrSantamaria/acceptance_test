package ocm

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/MrSantamaria/acceptance_test/pkg/assets"
	"github.com/MrSantamaria/acceptance_test/pkg/helpers"
	ocmsdk "github.com/openshift-online/ocm-sdk-go"
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

func BackplaneLogin(clusterIDorName string) error {
	var stdout, stderr bytes.Buffer

	cmd := exec.Command("ocm",
		"backplane",
		"login",
		clusterIDorName)

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Print env variables
	fmt.Println("BACKPLANE_CONFIG:", os.Getenv("BACKPLANE_CONFIG"))

	cmd.Run()
	if cmd.Stderr != nil {
		errMsg := stderr.String()
		if strings.Contains(errMsg, "Your Backplane CLI is not up to date.") ||
			strings.Contains(errMsg, "level=warning msg=\"Env KUBE_PS1_CLUSTER_FUNCTION is not detected.") {
			return nil
		}
		return fmt.Errorf("error executing ocm backplane login:\nStandard Error: %s", stderr.String())
	}

	fmt.Println("Backplane Login Successful")
	return nil
}

func SetUpOcmBinary() error {
	repoOwner, repoName := "openshift-online", "ocm-cli"
	// Get the latest release of the ocm-cli
	release, err := helpers.GetLatestRelease(repoOwner, repoName)
	if err != nil {
		fmt.Println("Error:", err)
		return err
	}

	selectedURL, err := helpers.SelectVersionByRuntime(release)
	if err != nil {
		fmt.Println("Error:", err)
		return err
	}

	// Extract the file name from the URL to use as the local file name
	tokens := strings.Split(selectedURL, "/")
	fileName := tokens[len(tokens)-1]

	ocmBinaryPath, err := helpers.DownloadRelease(selectedURL, fileName)
	if err != nil {
		fmt.Println("Error:", err)
		return err
	}

	err = helpers.SetupBinary(ocmBinaryPath)
	if err != nil {
		fmt.Println("Error:", err)
		return err
	}

	err = helpers.CheckBinary(ocmBinaryPath)
	if err != nil {
		fmt.Println("Error:", err)
		return err
	}

	return nil
}
