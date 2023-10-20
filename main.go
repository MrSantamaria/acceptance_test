package main

import (
	"fmt"
	"os"

	"github.com/MrSantamaria/acceptance_test/pkg/helpers"
	"github.com/MrSantamaria/acceptance_test/workflows"
)

func main() {

	args := helpers.RunConfig

	err := workflows.SetUp()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	err = workflows.ValidateRequirements()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	err = workflows.AcceptanceTest(args.OCM_TOKEN, args.ENVIRONMENT, args.NAMESPACE, args.OPERATOR_NAME, args.PAIRED_SSS, args.IMAGE_TAG)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	os.Exit(0)
}
