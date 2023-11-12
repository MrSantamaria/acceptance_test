package helpers

var openshiftSectorMap = map[string]bool{
	"canary":    true,
	"ibm-infra": true,
	"main":      true,
}

func IsOpenshiftSector(sector string) bool {
	return openshiftSectorMap[sector]
}
