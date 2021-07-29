package mcs

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataSourceDatabaseUser() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceDatabaseUserRead,
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

			"password": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"host": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"databases": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func dataSourceDatabaseUserRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(Config)
	DatabaseV1Client, err := config.DatabaseV1Client(getRegion(d, config))
	if err != nil {
		return fmt.Errorf("error creating mcs database client: %s", err)
	}

	id := d.Get("id").(string)
	userID := strings.SplitN(id, "/", 2)
	if len(userID) != 2 {
		return fmt.Errorf("invalid mcs_db_user id: %s", id)
	}

	instanceID := userID[0]
	userName := userID[1]

	exists, userObj, err := databaseUserExists(DatabaseV1Client, instanceID, userName)
	if err != nil {
		return fmt.Errorf("error checking if mcs_db_user %s exists: %s", d.Id(), err)
	}

	if !exists {
		d.SetId("")
		return nil
	}

	d.SetId(id)
	d.Set("name", userName)

	databases := flattenDatabaseUserDatabases(userObj.Databases)
	if err := d.Set("databases", databases); err != nil {
		return fmt.Errorf("unable to set databases: %s", err)
	}

	return nil
}
