package main

import (
	"context"
	"fmt"

	"github.com/openshift/hive/contrib/pkg/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func main() {
	// Create a new context.
	ctx := context.TODO()

	client, err := utils.GetClient()
	if err != nil {
		fmt.Printf("Error creating dynamic client: %v\n", err)
		return
	}

	// Use this client to list all namespaces.
	nsList, err := client.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		fmt.Printf("Error listing namespaces: %v\n", err)
		return
	}

	// Print the list of namespaces.
	fmt.Printf("Namespaces:\n")
	for _, ns := range nsList.Items {
		fmt.Printf("  %s\n", ns.Name)
	}

}
