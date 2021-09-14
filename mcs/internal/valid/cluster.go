package valid

import (
	"errors"
	"strings"

	"github.com/MailRuCloudSolutions/terraform-provider-mcs/mcs/internal/util/textutil"
)

var (
	ErrInvalidClusterName      = errors.New("invalid cluster name")
	ErrInvalidAvailabilityZone = errors.New("invalid availability zone")
)

// ClusterName validates name of cluster.
// Value should match the pattern ^[a-zA-Z][a-zA-Z0-9_.-]*$
func ClusterName(name string) error {
	if len(name) == 0 {
		return ErrInvalidClusterName
	}

	if !textutil.IsLetter(rune(name[0])) {
		return ErrInvalidClusterName
	}

	for _, r := range name[1:] {
		if !textutil.IsLetterDigitSymbol(r, '_', '.', '-') {
			return ErrInvalidClusterName
		}
	}

	return nil
}

var availabilityAvailabilityZones = map[string]struct{}{
	"dp1": {},
	"ms1": {},
}

// AvailabilityZone validates provided availability zone.
func AvailabilityZone(name string) error {
	if _, ok := availabilityAvailabilityZones[strings.ToLower(name)]; !ok {
		return ErrInvalidAvailabilityZone
	}
	return nil
}
