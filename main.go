package main

import (
	"fmt"
	"os"

	"github.com/MrSantamaria/acceptance_test/workflows"
)

func main() {

	config, err := workflows.SetUp()
	if err != nil {
		fmt.Println("Error while setting up:", err)
		return
	}

	err = workflows.ValidateRequirements()
	if err != nil {
		fmt.Println("Error while validating requirements:", err)
		return
	}

	err = workflows.AcceptanceTest(config.GetOCMToken(), config.GetEnvironment(), config.GetNamespace(), config.GetOperatorName(), config.GetPairedSSS(), config.GetImageTag())
	if err != nil {
		fmt.Println("Error while running acceptance test:", err)
		os.Exit(1)
	}

	os.Exit(0)
}
