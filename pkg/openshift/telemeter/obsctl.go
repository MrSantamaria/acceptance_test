package telemeter

import (
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
	OidcOfflineAccess bool
	Tenant            string
	LogLevel          string
}

type environmentConfig struct {
	intStage obsctlConfig
	prod     obsctlConfig
}

var (
	stageContextAPI = "https://observatorium.api.stage.openshift.com"
	config          = environmentConfig{
		intStage: obsctlConfig{
			ContextName:       "staging-api-x",
			ContextApi:        "https://observatorium.api.stage.openshift.com/",
			OidcAudience:      "observatorium-telemeter-staging",
			OidcClientID:      viper.GetString("telemeterClientID"),
			OidcClientSecret:  viper.GetString("telemeterSecret"),
			OidcIssuerURL:     "https://sso.redhat.com/auth/realms/redhat-external",
			OidcOfflineAccess: false,
			Tenant:            "telemeter",
			LogLevel:          "debug",
		},
		// TODO: Update this to production once we have a production environment
		prod: obsctlConfig{
			ContextName:       "production-api-x",
			ContextApi:        "https://observatorium.api.openshift.com/",
			OidcAudience:      "observatorium-telemeter-production",
			OidcClientID:      "observatorium-sdtcs-production",
			OidcClientSecret:  "placeHolder",
			OidcIssuerURL:     "https://sso.redhat.com/auth/realms/redhat-external",
			OidcOfflineAccess: false,
			Tenant:            "telemeter",
			LogLevel:          "debug",
		},
	}
)

func GetObsctlConfig() obsctlConfig {
	switch viper.GetString("environment") {
	case "int":
		return config.intStage
	case "stage":
		return config.intStage
	case "prod":
		return config.prod
	default:
		fmt.Println("environment not set, defaulting to int")
	}

	return config.intStage
}

func CliCheck() bool {
	cmd := exec.Command("obsctl")
	err := cmd.Run()
	if err == nil {
		return true
	}

	fmt.Println("obscli is not configured.")

	return false
}

func AddObsctlContext(contextName, contextUrl string) error {
	// Run the obsctl command
	cmd := exec.Command("obsctl", "context", "api", "add", "--name="+contextName, "--url="+contextUrl)

	// Set the command's standard input, output, and error streams to be the same as the current process
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Run the command
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error running obsctl command: %v", err)
	}

	return nil
}

func ObsctlLogin(api, audience, clientID, clientSecret, issuerURL, offlineAccess, tenant, logLevel string) error {

	// Run the obsctl login command
	cmd := exec.Command("obsctl", "login",
		"--api="+api,
		"--oidc.audience="+audience,
		"--oidc.client-id="+clientID,
		"--oidc.client-secret="+clientSecret,
		"--oidc.issuer-url="+issuerURL,
		"--oidc.offline-access="+offlineAccess,
		"--tenant="+tenant,
		"--log.level="+logLevel,
	)

	// Set the command's standard input, output, and error streams to be the same as the current process
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Run the command
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("error running obsctl login command: %v", err)
	}

	return nil
}
