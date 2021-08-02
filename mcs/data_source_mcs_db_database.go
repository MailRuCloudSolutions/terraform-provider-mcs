package mcs

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataSourceDatabaseDatabase() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceDatabaseDatabaseRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Required: true,
			},

			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"instance_id": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"charset": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},

			"collate": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
		},
	}
}

func dataSourceDatabaseDatabaseRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(configer)
	DatabaseV1Client, err := config.DatabaseV1Client(getRegion(d, config))
	if err != nil {
		return fmt.Errorf("error creating mcs database client: %s", err)
	}

	id := d.Get("id").(string)
	databaseID := strings.SplitN(id, "/", 2)
	if len(databaseID) != 2 {
		return fmt.Errorf("invalid mcs_db_database id: %s", id)
	}

	instanceID := databaseID[0]
	databaseName := databaseID[1]

	exists, err := databaseDatabaseExists(DatabaseV1Client, instanceID, databaseName)
	if err != nil {
		return fmt.Errorf("error checking if mcs_db_database %s exists: %s", d.Id(), err)
	}

	if !exists {
		d.SetId("")
		return nil
	}

	d.SetId(id)
	d.Set("name", databaseName)

	return nil
}
