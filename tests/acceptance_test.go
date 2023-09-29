package tests

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	ocmclient "github.com/openshift/osde2e-common/pkg/clients/ocm"
	osdProvider "github.com/openshift/osde2e-common/pkg/openshift/osd"
)

var (
	ctx                                 = context.Background()
	testSuiteName                       = "Progressive Delivery"
	logger                              = GinkgoLogr
	managementClusters, serviceClusters []string
)

var _ = BeforeSuite(func() {
	ctx = context.WithValue(ctx, "ocmToken", "")
	// Create a new OSD Provider
	osdProvider, err := osdProvider.New(
		ctx,
		ctx.Value("ocmToken").(string),
		ocmclient.Integration,
		logger,
	)
	Expect(err).NotTo(HaveOccurred())

	// Get the client for the resources that manages the collection of clusters
	collection := osdProvider.ClustersMgmt().V1().Clusters()

	// Retrieve All managementClusters
	response, err := collection.List().Search("name like '%hs-mc%'").SendContext(ctx)
	Expect(err).NotTo(HaveOccurred(), "Unable to retrieve Management Clusters")
	response.Items().Range(func(index int, cluster *cmv1.Cluster) bool {
		managementClusters = append(managementClusters, fmt.Sprintf("%s - %s - %s - %s\n", cluster.ID(), cluster.Name(), cluster.Region().ID(), cluster.State()))
		//fmt.Printf("%s - %s - %s - %s\n", cluster.ID(), cluster.Name(), cluster.Region().ID(), cluster.State())
		return true
	})
	Expect(managementClusters).NotTo(BeEmpty(), "No Management Clusters found")

	// Retrieve All ServiceClusters
	response, err = collection.List().Search("name like '%hs-sc%'").SendContext(ctx)
	Expect(err).NotTo(HaveOccurred(), "Unable to retrieve Service Clusters")
	response.Items().Range(func(index int, cluster *cmv1.Cluster) bool {
		serviceClusters = append(serviceClusters, fmt.Sprintf("%s - %s - %s - %s\n", cluster.ID(), cluster.Name(), cluster.Region().ID(), cluster.State()))
		//fmt.Printf("%s - %s - %s - %s\n", cluster.ID(), cluster.Name(), cluster.Region().ID(), cluster.State())
		return true
	})
	Expect(managementClusters).NotTo(BeEmpty(), "No Service Clusters found")
})

var _ = DescribeTable(testSuiteName+" - SSS Applied Check", func(clusterID []string) {
	for _, cluster := range clusterID {
		Describe("SSS Aplied Check", func() {
			It("Should be applied", func() {
				fmt.Printf("Cluster Row is %s\n", cluster)
				Expect(cluster).NotTo(BeEmpty(), "no clusters were found")
			})
		})
	}
},
	Entry("Management Clusters", managementClusters),
	Entry("Service Clusters", serviceClusters),
)
