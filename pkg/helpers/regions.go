package helpers

// awsRegionMap is a package-level variable for quick lookups of AWS regions
var awsRegionMap = map[string]bool{
	"us-east-1":      true,
	"us-east-2":      true,
	"us-west-1":      true,
	"us-west-2":      true,
	"ca-central-1":   true,
	"eu-central-1":   true,
	"eu-west-1":      true,
	"eu-west-2":      true,
	"eu-west-3":      true,
	"eu-north-1":     true,
	"ap-east-1":      true,
	"ap-south-1":     true,
	"ap-southeast-1": true,
	"ap-southeast-2": true,
	"ap-northeast-1": true,
	"ap-northeast-2": true,
	"sa-east-1":      true,
	"me-south-1":     true,
}

func IsAWSRegion(region string) bool {
	return awsRegionMap[region]
}
