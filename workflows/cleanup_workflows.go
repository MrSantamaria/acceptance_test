package workflows

import (
	"fmt"

	"github.com/MrSantamaria/acceptance_test/pkg/openshift/telemeter"
	"github.com/spf13/viper"
)

func CleanUp() error {
	// TODO: Update how this function is called once the telemeter config is pointer based
	err := telemeter.ObsctlLogout(telemeter.SetObsctlConfig(viper.GetString("environment")))
	if err != nil {
		fmt.Println("ERROR: Failed to logout local Telemeter instance")
		return err
	}

	return nil
}
