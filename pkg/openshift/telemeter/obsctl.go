package telemeter

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/viper"
)

type obsctlConfig struct {
	ContextName       string
	ContextApi        string
	OidcAudience      string
	OidcClientID      string
	OidcClientSecret  string
	OidcIssuerURL     string
	OidcOfflineAccess string
	Tenant            string
	LogLevel          string
}

type environmentConfig struct {
	intStage obsctlConfig
	prod     obsctlConfig
}

type obsctlSearchResult struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Metric struct {
				CSV  string `json:"__name__"`
				ID   string `json:"_id"`
				Name string `json:"name"`
			} `json:"metric"`
			Values [][]interface{} `json:"values"`
		} `json:"result"`
	} `json:"data"`
}

var (
	config = environmentConfig{
		intStage: obsctlConfig{
			ContextName:       "staging-api-x",
			ContextApi:        "https://observatorium.api.stage.openshift.com/",
			OidcAudience:      "observatorium-telemeter-staging",
			OidcClientID:      "",
			OidcClientSecret:  "",
			OidcIssuerURL:     "https://sso.redhat.com/auth/realms/redhat-external",
			OidcOfflineAccess: "false",
			Tenant:            "telemeter",
			LogLevel:          "debug",
		},
		// TODO: Update this to production once we have a production environment
		prod: obsctlConfig{
			ContextName:       "production-api-x",
			ContextApi:        "https://observatorium.api.openshift.com/",
			OidcAudience:      "observatorium-telemeter-production",
			OidcClientID:      "",
			OidcClientSecret:  "",
			OidcIssuerURL:     "https://sso.redhat.com/auth/realms/redhat-external",
			OidcOfflineAccess: "false",
			Tenant:            "telemeter",
			LogLevel:          "debug",
		},
	}
)

func CliCheck() bool {
	cmd := exec.Command("obsctl")
	err := cmd.Run()
	if err == nil {
		return true
	}

	fmt.Println("obscli is not configured.")

	return false
}

func SetObsctlConfig(environment string) obsctlConfig {
	switch environment {
	case "int":
		return config.intStage
	case "stage":
		return config.intStage
	case "prod":
		return config.prod
	default:
		fmt.Println("ERROR: Invalid environment. Please use int, stage, or prod.")
		return obsctlConfig{}
	}
}

func addObsctlContext(telemeterConfig *obsctlConfig) error {
	var ErrSkipContextCreation = fmt.Errorf("Context already exists. Skipping context creation.")
	cmd := exec.Command("obsctl", "context", "api", "add", "--name="+telemeterConfig.ContextName, "--url="+telemeterConfig.ContextApi)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if errors.As(err, &ErrSkipContextCreation) {
		fmt.Println("Context already exists. Skipping context creation.")
		return nil
	}
	if err != nil {
		return fmt.Errorf("error running obsctl context command: %v", err)
	}

	return nil
}

func updateObsctlConfig(telemeterConfig *obsctlConfig) error {
	telemeterConfig.OidcClientID = viper.GetString("TELEMETER_CLIENT_ID")
	if len(telemeterConfig.OidcClientID) == 0 {
		return fmt.Errorf("TELEMETER_CLIENT_ID is required")
	}
	telemeterConfig.OidcClientSecret = viper.GetString("TELEMETER_SECRET")
	if len(telemeterConfig.OidcClientSecret) == 0 {
		return fmt.Errorf("TELEMETER_SECRET is required")
	}

	return nil
}

func ObsctlLogin(telemeterConfig obsctlConfig) error {
	err := updateObsctlConfig(&telemeterConfig)
	if err != nil {
		return fmt.Errorf("error updating obsctl config: %v", err)
	}
	err = addObsctlContext(&telemeterConfig)
	if err != nil {
		return fmt.Errorf("error adding obsctl context: %v", err)
	}

	cmd := exec.Command("obsctl", "login",
		"--api="+telemeterConfig.ContextName,
		"--oidc.audience="+telemeterConfig.OidcAudience,
		"--oidc.client-id="+telemeterConfig.OidcClientID,
		"--oidc.client-secret="+telemeterConfig.OidcClientSecret,
		"--oidc.issuer-url="+telemeterConfig.OidcIssuerURL,
		"--oidc.offline-access="+telemeterConfig.OidcOfflineAccess,
		"--tenant="+telemeterConfig.Tenant,
		"--log.level="+telemeterConfig.LogLevel,
	)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Printf("Running command: %s\n", cmd.Args)
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("error running obsctl login command: %v", err)
	}

	return nil
}

func ObsctlSetContext(searchQuery string) error {
	cmd := exec.Command("obsctl", "")

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Printf("Running command: %s\n", cmd.Args)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("error running obsctl context command: %v", err)
	}

	return nil
}

func ObsctlSearchQuery(searchQuery string) (obsctlSearchResult, error) {
	var obsctlSearchResult obsctlSearchResult
	cmd := exec.Command("obsctl", "metrics", "query", searchQuery)

	fmt.Printf("Running command: %s\n", cmd.Args)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return obsctlSearchResult, fmt.Errorf("error running obsctl metrics query command: %v", err)
	}

	err = json.Unmarshal(output, &obsctlSearchResult)
	if err != nil {
		return obsctlSearchResult, fmt.Errorf("error unmarshalling obsctl metrics query output: %v", err)
	}

	return obsctlSearchResult, nil
}

func ObsctlProccessSearchResult(searchResult obsctlSearchResult) int {
	var csvCount int

	// TODO: Rewrite this to be more efficient
	for _, result := range searchResult.Data.Result {
		if result.Metric.CSV == "csv_succeeded" {
			csvCount++
		}
		if result.Metric.CSV == "csv_abnormal" {
			csvCount++
		}
	}

	return csvCount
}

func ObsctlLogout(telemeterConfig obsctlConfig) error {
	cmd := exec.Command("obsctl", "logout",
		"--api="+telemeterConfig.ContextName,
		"--tenant="+telemeterConfig.Tenant,
	)

	// Set the command's standard input, output, and error streams to be the same as the current process
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Printf("Running command: %s\n", cmd.Args)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("error running obsctl logout command: %v", err)
	}

	return nil
}
