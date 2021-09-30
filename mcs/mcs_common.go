package mcs

import (
	"fmt"
	"strings"
	"time"

	"github.com/gophercloud/gophercloud"
)

// mcsErrorHandler overrides error messages
type mcsErrorHandler struct {
	gophercloud.ErrUnexpectedResponseCode
}

// mcsError404 is needed to customize http 404 error message
type mcsError404 struct {
	gophercloud.ErrUnexpectedResponseCode
}

// Error404 overrides gophercloud http 404 error message
func (e mcsErrorHandler) Error404(res gophercloud.ErrUnexpectedResponseCode) error {
	return mcsError404{res}
}

func (e mcsError404) Error() string {
	return fmt.Sprintf("resource not found with: [%s %s], error message: %s",
		e.Method, e.URL, e.Body)
}

func getRequestOpts(codes ...int) *gophercloud.RequestOpts {
	reqOpts := &gophercloud.RequestOpts{
		OkCodes:      codes,
		ErrorContext: mcsErrorHandler{},
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
