package mcs

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceDatabaseDatabase() *schema.Resource {
	return &schema.Resource{
		Create: resourceDatabaseDatabaseCreate,
		Read:   resourceDatabaseDatabaseRead,
		Delete: resourceDatabaseDatabaseDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(dbDatabaseCreateTimeout),
			Delete: schema.DefaultTimeout(dbDatabaseDeleteTimeout),
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"instance_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
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

func resourceDatabaseDatabaseCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(Config)
	DatabaseV1Client, err := config.DatabaseV1Client(getRegion(d, config))
	if err != nil {
		return fmt.Errorf("error creating OpenStack database client: %s", err)
	}

	databaseName := d.Get("name").(string)
	instanceID := d.Get("instance_id").(string)

	instance, err := instanceGet(DatabaseV1Client, instanceID).extract()
	if err != nil {
		return fmt.Errorf("error while getting mcs_db_instance: %s", err)
	}
	if instance.DataStore.Type == Redis {
		return fmt.Errorf("operation not supported for this datastore")
	}
	if instance.ReplicaOf != nil {
		return fmt.Errorf("operation not supported for replica")
	}

	var databasesList databaseBatchCreateOpts

	db := databaseCreateOpts{
		Name:    databaseName,
		CharSet: d.Get("charset").(string),
		Collate: d.Get("collate").(string),
	}

	databasesList.Databases = append(databasesList.Databases, db)
	err = databaseCreate(DatabaseV1Client, instanceID, &databasesList).ExtractErr()
	if err != nil {
		return fmt.Errorf("error creating mcs_db_database: %s", err)
	}

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"BUILD"},
		Target:     []string{"ACTIVE"},
		Refresh:    databaseDatabaseStateRefreshFunc(DatabaseV1Client, instanceID, databaseName),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      dbDatabaseDelay,
		MinTimeout: dbDatabaseMinTimeout,
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("error waiting for mcs_db_database %s to be created: %s", databaseName, err)
	}

	// Store the ID now
	d.SetId(fmt.Sprintf("%s/%s", instanceID, databaseName))

	return resourceDatabaseDatabaseRead(d, meta)
}

func resourceDatabaseDatabaseRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(Config)
	DatabaseV1Client, err := config.DatabaseV1Client(getRegion(d, config))
	if err != nil {
		return fmt.Errorf("error creating mcs database client: %s", err)
	}

	databaseID := strings.SplitN(d.Id(), "/", 2)
	if len(databaseID) != 2 {
		return fmt.Errorf("invalid mcs_db_database ID: %s", d.Id())
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

	d.Set("name", databaseName)

	return nil
}

func resourceDatabaseDatabaseDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(Config)
	DatabaseV1Client, err := config.DatabaseV1Client(getRegion(d, config))
	if err != nil {
		return fmt.Errorf("error creating mcs database client: %s", err)
	}

	databaseID := strings.SplitN(d.Id(), "/", 2)
	if len(databaseID) != 2 {
		return fmt.Errorf("invalid mcs_db_database ID: %s", d.Id())
	}

	instanceID := databaseID[0]
	databaseName := databaseID[1]

	exists, err := databaseDatabaseExists(DatabaseV1Client, instanceID, databaseName)
	if err != nil {
		return fmt.Errorf("error checking if mcs_db_database %s exists: %s", d.Id(), err)
	}

	if !exists {
		return nil
	}

	err = databaseDelete(DatabaseV1Client, instanceID, databaseName).ExtractErr()
	if err != nil {
		return fmt.Errorf("error deleting mcs_db_database %s: %s", d.Id(), err)
	}

	return nil
}
