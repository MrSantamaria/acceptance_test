package workflows

import (
	"fmt"

	"github.com/MrSantamaria/acceptance_test/pkg/openshift/ocm"
	"github.com/spf13/viper"
)

func SetUp(ocmToken, environment string) error {
	var errs []error
	var err error

	if !ocm.CliCheck() {
		errs = append(errs, fmt.Errorf("ocm-cli is not installed"))
	}

	err = validateRequiredVars()
	if err != nil {
		errs = append(errs, err)
	}

	err = ocm.Login(ocmToken, environment)
	if err != nil {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return fmt.Errorf("Acceptance Test Setup Failed: %v", errs)
	}

	return nil
}

func validateRequiredVars() error {
	var errs []error

	if len(viper.GetString("token")) == 0 {
		errs = append(errs, fmt.Errorf("token is required"))
	}

	if len(viper.GetString("environment")) == 0 {
		errs = append(errs, fmt.Errorf("environment is required"))
	}

	if len(viper.GetStringSlice("selectors")) == 0 {
		errs = append(errs, fmt.Errorf("selectors are required"))
	}

	/*

		if len(viper.GetString("operator")) == 0 {
			errs = append(errs, fmt.Errorf("operator is required"))
		}

		if len(viper.GetString("imagetag")) == 0 {
			errs = append(errs, fmt.Errorf("imagetag is required"))
		}
	*/

	if len(errs) > 0 {
		return fmt.Errorf("failed to validate required vars: %v", errs)
	}

	return nil
}
