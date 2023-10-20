package workflows

import (
	"fmt"

	"github.com/MrSantamaria/acceptance_test/pkg/helpers"
	"github.com/MrSantamaria/acceptance_test/pkg/openshift/oc"
	"github.com/MrSantamaria/acceptance_test/pkg/openshift/ocm"
)

func SetUp() error {
	err := helpers.RunConfig.InitRunConfig()
	if err != nil {
		return err
	}

	if !ocm.CliCheck() {
		err = ocm.SetUpOcmBinary()
		if err != nil {
			return err
		}
	}

	return nil
}

func ValidateRequirements() error {
	var err []error

	if !oc.CliCheck() {
		err = append(err, fmt.Errorf("oc-cli is not installed"))
	}

	if !ocm.CliCheck() {
		err = append(err, fmt.Errorf("ocm-cli is not installed"))
	}

	// TODO: Validate that a client can be created for the cluster

	if len(err) > 0 {
		return fmt.Errorf("failed to validate required cli tools: %v", err)
	}

	return nil
}
