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
				Type:       schema.TypeString,
				Optional:   true,
				Deprecated: "Please, use dmbs_id attribute instead",
			},

			"dbms_id": {
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

	dbmsID := databaseID[0]
	databaseName := databaseID[1]
	dbmsResp, err := getDBMSResource(DatabaseV1Client, dbmsID)
	if err != nil {
		return fmt.Errorf("error while getting resource: %s", err)
	}
	var dbmsType string
	if _, ok := dbmsResp.(instanceResp); ok {
		dbmsType = dbmsTypeInstance
	}
	if _, ok := dbmsResp.(dbClusterResp); ok {
		dbmsType = dbmsTypeCluster
	}
	exists, err := databaseDatabaseExists(DatabaseV1Client, dbmsID, databaseName, dbmsType)
	if err != nil {
		return fmt.Errorf("error checking if mcs_db_database %s exists: %s", d.Id(), err)
	}

	if !exists {
		d.SetId("")
		return nil
	}

	d.SetId(id)
	d.Set("name", databaseName)
	d.Set("instance_id", dbmsID)
	d.Set("dbms_id", dbmsID)
	return nil
}
