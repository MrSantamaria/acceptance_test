package helpers

import (
	"flag"
	"fmt"
	"os"
	"sync"
)

type RunConfig struct {
	OCM_TOKEN     string
	ENVIRONMENT   string
	NAMESPACE     string
	OPERATOR_NAME string
	PAIRED_SSS    string
	IMAGE_TAG     string
	once          sync.Once
}

func (r *RunConfig) InitRunConfig() error {
	var err error

	r.once.Do(func() {
		ocmToken := flag.String("token", "", "OCM Token")
		environment := flag.String("environment", "", "Environment")
		namespace := flag.String("namespace", "", "Namespace")
		operatorName := flag.String("operator", "", "Operator Name")
		pairedSSS := flag.String("pairedSss", "", "Paired SSS")
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
			*pairedSSS = os.Getenv("PAIRED_SSS")
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

func (r *RunConfig) GetOCMToken() string {
	return r.OCM_TOKEN
}

func (r *RunConfig) GetEnvironment() string {
	return r.ENVIRONMENT
}

func (r *RunConfig) GetNamespace() string {
	return r.NAMESPACE
}

func (r *RunConfig) GetOperatorName() string {
	return r.OPERATOR_NAME
}

func (r *RunConfig) GetPairedSSS() string {
	return r.PAIRED_SSS
}

func (r *RunConfig) GetImageTag() string {
	return r.IMAGE_TAG
}
