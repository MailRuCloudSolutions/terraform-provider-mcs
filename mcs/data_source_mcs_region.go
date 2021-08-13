package mcs

import (
	"fmt"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack/identity/v3/regions"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataSourceMcsRegion() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceMcsRegionRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"parent_region": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceMcsRegionRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(configer)
	client, err := config.IdentityV3Client(config.GetRegion())
	if err != nil {
		return fmt.Errorf("failed to init identity v3 client: %s", err)
	}

	// default region
	regionName := config.GetRegion()
	// or passed from config
	if v, ok := d.GetOk("id"); ok {
		regionName = v.(string)
	}

	region, err := regions.Get(client.(*gophercloud.ServiceClient), regionName).Extract()
	if err != nil {
		return fmt.Errorf("failed to get region for %s: %s", regionName, err)
	}

	d.SetId(region.ID)
	d.Set("parent_region", region.ParentRegionID)
	d.Set("description", region.Description)
	return nil
}
