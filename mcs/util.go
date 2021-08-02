package mcs

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/gophercloud/gophercloud"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/mitchellh/mapstructure"
)

var decoderConfig = &mapstructure.DecoderConfig{
	TagName: "json",
}

// mapStructureDecoder ...
func mapStructureDecoder(strct interface{}, v *map[string]interface{}, config *mapstructure.DecoderConfig) error {
	config.Result = strct
	decoder, _ := mapstructure.NewDecoder(config)
	return decoder.Decode(*v)
}

// getTimestamp ...
func getTimestamp(t *time.Time) string {
	if t != nil {
		return t.Format(time.RFC3339)
	}
	return ""
}

// checkDeleted checks the error to see if it's a 404 (Not Found) and, if so,
// sets the resource ID to the empty string instead of throwing an error.
func checkDeleted(d *schema.ResourceData, err error, msg string) error {
	if _, ok := err.(gophercloud.ErrDefault404); ok {
		d.SetId("")
		return nil
	}

	return fmt.Errorf("%s %s: %s", msg, d.Id(), err)
}

// getRegion returns the region that was specified in the resource. If a
// region was not set, the provider-level region is checked. The provider-level
// region can either be set by the region argument or by OS_REGION_NAME.
func getRegion(d *schema.ResourceData, config configer) string {
	if v, ok := d.GetOk("region"); ok {
		return v.(string)
	}

	return config.GetRegion()
}

func ensureOnlyOnePresented(d *schema.ResourceData, keys ...string) (string, error) {
	var isPresented bool
	var keyPresented string
	for _, key := range keys {
		_, ok := d.GetOk(key)

		if ok {
			if isPresented {
				return "", fmt.Errorf("only one of %v keys can be presented", keys)
			}

			isPresented = true
			keyPresented = key
		}
	}

	if !isPresented {
		return "", fmt.Errorf("no one of %v keys are presented", keys)
	}

	return keyPresented, nil
}

func randomName(n int) string {
	charSet := []byte("abcdefghijklmnopqrstuvwxyz012346789")
	result := make([]byte, 0, n)
	for i := 0; i < n; i++ {
		result = append(result, charSet[rand.Intn(len(charSet))])
	}
	return string(result)
}
