package helpers

import (
	"flag"
	"fmt"
	"os"
	"sync"
)

type runConfig struct {
	OCM_TOKEN     string
	ENVIRONMENT   string
	NAMESPACE     string
	OPERATOR_NAME string
	PAIRED_SSS    string
	IMAGE_TAG     string
	once          sync.Once
}

var RunConfig *runConfig

func (r *runConfig) InitRunConfig() error {
	var err error

	r.once.Do(func() {
		ocmToken := flag.String("token", "", "OCM Token")
		environment := flag.String("environment", "", "Environment")
		namespace := flag.String("namespace", "", "Namespace")
		operatorName := flag.String("operator", "", "Operator Name")
		pairedSSS := flag.String("sss", "", "Paired SSS")
		imageTag := flag.String("imageTag", "", "Image Tag")

		flag.Parse()

		if *ocmToken == "" {
			*ocmToken = os.Getenv("TOKEN")
		}
		if *environment == "" {
			*environment = os.Getenv("ENVIRONMENT")
		}
		if *namespace == "" {
			*namespace = os.Getenv("NAMESPACE")
		}
		if *operatorName == "" {
			*operatorName = os.Getenv("OPERATOR")
		}
		if *pairedSSS == "" {
			*pairedSSS = os.Getenv("SSS")
		}
		if *imageTag == "" {
			*imageTag = os.Getenv("IMAGE_TAG")
		}

		if *ocmToken == "" || *environment == "" || *namespace == "" || *operatorName == "" || *pairedSSS == "" || *imageTag == "" {
			flag.PrintDefaults()
			err = fmt.Errorf("missing required flag(s) to run the acceptance test")
		}

		r.OCM_TOKEN = *ocmToken
		r.ENVIRONMENT = *environment
		r.NAMESPACE = *namespace
		r.OPERATOR_NAME = *operatorName
		r.PAIRED_SSS = *pairedSSS
		r.IMAGE_TAG = *imageTag
	})

	return err
}

func GetRunConfig() *runConfig {
	return RunConfig
}
