package mcs

import (
	"strings"
	"time"

	"github.com/gophercloud/gophercloud"
)

func getRequestOpts(codes ...int) *gophercloud.RequestOpts {
	reqOpts := &gophercloud.RequestOpts{
		OkCodes: codes,
	}
	if len(codes) != 0 {
		reqOpts.OkCodes = codes
	}
	addMicroVersionHeader(reqOpts)
	return reqOpts
}

// dateTimeWithoutTZFormat represents format of time used in dbaas
type dateTimeWithoutTZFormat struct {
	time.Time
}

// UnmarshalJSON is used to correctly unmarshal datetime fields
func (t *dateTimeWithoutTZFormat) UnmarshalJSON(b []byte) (err error) {
	layout := "2006-01-02T15:04:05"
	s := strings.Trim(string(b), "\"")
	if s == "null" {
		return
	}
	t.Time, err = time.Parse(layout, s)
	return
}
