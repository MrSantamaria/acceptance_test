package ocm

import (
	"bytes"
	"context"
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

func SetupOcmClient(ctx context.Context) (*ocmClient, error) {
	token := ctx.Value("ocmToken").(string)
	env := ctx.Value("env").(string)

	connection, err := ocmsdk.NewConnectionBuilder().
		URL(env).
		Tokens(token).
		BuildContext(ctx)
	if err != nil {
		return nil, err
	}

	return &ocmClient{connection}, nil
}

func Login(ctx context.Context, ocmBinaryPath string, token string, environment string) error {
	// Check if the ocm binary path is empty
	if ocmBinaryPath == "" {
		return fmt.Errorf("ocm binary path cannot be empty")
	}
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
	cmd := exec.CommandContext(ctx, ocmBinaryPath, "login", "--token", token, "--url", env[environment])

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

func GetManagementAndServiceClusters(ctx context.Context, ocmBinaryPath string, region, sector string) ([]string, []string, error) {
	var stdout, stderr bytes.Buffer
	var managementClusters, serviceClusters []string

	cmd := exec.CommandContext(ctx, ocmBinaryPath,
		"list",
		"clusters",
		"--parameter",
		"search=name like '%hs-mc%'",
		"--parameter",
		"search=region.id = '"+region+"'",
		"--columns", "ID",
		"--no-headers")

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return nil, nil, fmt.Errorf("error executing the command: %v\nStandard Error: %s", err, stderr.String())
	}

	// Append IDs into the managementClusters slice
	for _, id := range bytes.Split(stdout.Bytes(), []byte("\n")) {
		if len(id) > 0 {
			managementClusters = append(managementClusters, string(id))
		}
	}

	// Reset the buffer
	stdout.Reset()
	stderr.Reset()

	cmd = exec.CommandContext(ctx, ocmBinaryPath,
		"list",
		"clusters",
		"--parameter",
		"search=name like '%hs-sc%'",
		"--parameter",
		"search=region.id = '"+region+"'",
		"--columns", "ID",
		"--no-headers")

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		return nil, nil, fmt.Errorf("error executing the command: %v\nStandard Error: %s", err, stderr.String())
	}

	// Append IDs into the serviceClusters slice
	for _, id := range bytes.Split(stdout.Bytes(), []byte("\n")) {
		if len(id) > 0 {
			serviceClusters = append(serviceClusters, string(id))
		}
	}

	return managementClusters, serviceClusters, nil
}

func BackplaneLogin(ctx context.Context, ocmBinaryPath string, clusterIDorName string) error {
	var stdout, stderr bytes.Buffer

	cmd := exec.CommandContext(ctx, ocmBinaryPath,
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
