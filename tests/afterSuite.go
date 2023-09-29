package tests

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	//. "github.com/onsi/gomega"
)

var _ = AfterSuite(func() {
	fmt.Println("AfterSuite")

})
