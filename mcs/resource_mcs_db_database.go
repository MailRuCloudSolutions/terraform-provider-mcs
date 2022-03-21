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
		Update: resourceDatabaseDatabaseUpdate,
		Delete: resourceDatabaseDatabaseDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

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
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      false,
				Deprecated:    "Please, use dmbs_id attribute instead",
				ConflictsWith: []string{"dbms_id"},
			},

			"dbms_id": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      false,
				ConflictsWith: []string{"instance_id"},
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

			"dbms_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceDatabaseDatabaseCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(configer)
	DatabaseV1Client, err := config.DatabaseV1Client(getRegion(d, config))
	if err != nil {
		return fmt.Errorf("error creating OpenStack database client: %s", err)
	}

	databaseName := d.Get("name").(string)
	dbmsIDRaw, dbmsIDOk := d.GetOk("dbms_id")
	instanceIDRaw, instanceIDOk := d.GetOk("instance_id")
	if !dbmsIDOk && !instanceIDOk {
		return fmt.Errorf("only dbms_id must be set")
	}
	var dbmsID string
	if instanceIDOk {
		dbmsID = instanceIDRaw.(string)
	} else {
		dbmsID = dbmsIDRaw.(string)
	}

	dbmsResp, err := getDBMSResource(DatabaseV1Client, dbmsID)
	if err != nil {
		return fmt.Errorf("error while getting instance or cluster: %s", err)
	}
	var dbmsType string
	if instanceResource, ok := dbmsResp.(*instanceResp); ok {
		if isOperationNotSupported(instanceResource.DataStore.Type, Redis, Tarantool) {
			return fmt.Errorf("operation not supported for this datastore")
		}
		if instanceResource.ReplicaOf != nil {
			return fmt.Errorf("operation not supported for replica")
		}
		dbmsType = dbmsTypeInstance
	}
	if clusterResource, ok := dbmsResp.(*dbClusterResp); ok {
		if isOperationNotSupported(clusterResource.DataStore.Type, Redis, Tarantool) {
			return fmt.Errorf("operation not supported for this datastore")
		}
		dbmsType = dbmsTypeCluster
	}
	var databasesList databaseBatchCreateOpts

	db := databaseCreateOpts{
		Name:    databaseName,
		CharSet: d.Get("charset").(string),
		Collate: d.Get("collate").(string),
	}

	databasesList.Databases = append(databasesList.Databases, db)
	err = databaseCreate(DatabaseV1Client, dbmsID, &databasesList, dbmsType).ExtractErr()
	if err != nil {
		return fmt.Errorf("error creating mcs_db_database: %s", err)
	}

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"BUILD"},
		Target:     []string{"ACTIVE"},
		Refresh:    databaseDatabaseStateRefreshFunc(DatabaseV1Client, dbmsID, databaseName, dbmsType),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      dbDatabaseDelay,
		MinTimeout: dbDatabaseMinTimeout,
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("error waiting for mcs_db_database %s to be created: %s", databaseName, err)
	}

	// Store the ID now
	d.SetId(fmt.Sprintf("%s/%s", dbmsID, databaseName))
	// Store dbms type
	d.Set("dbms_type", dbmsType)

	return resourceDatabaseDatabaseRead(d, meta)
}

func resourceDatabaseDatabaseRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(configer)
	DatabaseV1Client, err := config.DatabaseV1Client(getRegion(d, config))
	if err != nil {
		return fmt.Errorf("error creating mcs database client: %s", err)
	}

	databaseID := strings.SplitN(d.Id(), "/", 2)
	if len(databaseID) != 2 {
		return fmt.Errorf("invalid mcs_db_database ID: %s", d.Id())
	}

	dbmsID := databaseID[0]
	databaseName := databaseID[1]

	var dbmsType string
	if dbmsTypeRaw, ok := d.GetOk("dbms_type"); ok {
		dbmsType = dbmsTypeRaw.(string)
	} else {
		dbmsType = dbmsTypeInstance
	}

	_, err = getDBMSResource(DatabaseV1Client, dbmsID)
	if err != nil {
		return checkDeleted(d, err, "Error retrieving mcs_db_database")
	}

	exists, err := databaseDatabaseExists(DatabaseV1Client, dbmsID, databaseName, dbmsType)
	if err != nil {
		return fmt.Errorf("error checking if mcs_db_database %s exists: %s", d.Id(), err)
	}

	if !exists {
		d.SetId("")
		return nil
	}

	d.Set("name", databaseName)
	if _, ok := d.GetOk("instance_id"); ok {
		d.Set("instance_id", dbmsID)
	}
	if _, ok := d.GetOk("dbms_id"); ok {
		d.Set("dbms_id", dbmsID)
	}

	return nil
}

func resourceDatabaseDatabaseUpdate(d *schema.ResourceData, meta interface{}) error {
	_, dbmsIDOk := d.GetOk("dbms_id")
	_, instanceIDOk := d.GetOk("instance_id")
	if !dbmsIDOk && !instanceIDOk {
		return fmt.Errorf("only dbms_id must be set")
	}
	return resourceDatabaseDatabaseRead(d, meta)
}

func resourceDatabaseDatabaseDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(configer)
	DatabaseV1Client, err := config.DatabaseV1Client(getRegion(d, config))
	if err != nil {
		return fmt.Errorf("error creating mcs database client: %s", err)
	}

	databaseID := strings.SplitN(d.Id(), "/", 2)
	if len(databaseID) != 2 {
		return fmt.Errorf("invalid mcs_db_database ID: %s", d.Id())
	}

	dbmsID := databaseID[0]
	databaseName := databaseID[1]
	dbmsType := d.Get("dbms_type").(string)

	exists, err := databaseDatabaseExists(DatabaseV1Client, dbmsID, databaseName, dbmsType)
	if err != nil {
		return fmt.Errorf("error checking if mcs_db_database %s exists: %s", d.Id(), err)
	}

	if !exists {
		return nil
	}

	err = databaseDelete(DatabaseV1Client, dbmsID, databaseName, dbmsType).ExtractErr()
	if err != nil {
		return fmt.Errorf("error deleting mcs_db_database %s: %s", d.Id(), err)
	}

	return nil
}
