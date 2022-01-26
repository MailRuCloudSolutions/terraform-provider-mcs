package mcs

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceDatabaseUser() *schema.Resource {
	return &schema.Resource{
		Create: resourceDatabaseUserCreate,
		Read:   resourceDatabaseUserRead,
		Delete: resourceDatabaseUserDelete,
		Update: resourceDatabaseUserUpdate,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(dbUserCreateTimeout),
			Delete: schema.DefaultTimeout(dbUserDeleteTimeout),
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: false,
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

			"password": {
				Type:      schema.TypeString,
				Required:  true,
				ForceNew:  false,
				Sensitive: true,
			},

			"host": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
			},

			"databases": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				ForceNew: false,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},

			"dbms_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceDatabaseUserCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(configer)
	DatabaseV1Client, err := config.DatabaseV1Client(getRegion(d, config))
	if err != nil {
		return fmt.Errorf("error creating OpenStack database client: %s", err)
	}

	userName := d.Get("name").(string)
	rawDatabases := d.Get("databases").([]interface{})
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
		return fmt.Errorf("error while getting resource: %s", err)
	}
	var dbmsType string
	if instanceResource, ok := dbmsResp.(instanceResp); ok {
		if instanceResource.DataStore.Type == Redis {
			return fmt.Errorf("operation not supported for this datastore")
		}
		if instanceResource.ReplicaOf != nil {
			return fmt.Errorf("operation not supported for replica")
		}
		dbmsType = dbmsTypeInstance
	}
	if clusterResource, ok := dbmsResp.(dbClusterResp); ok {
		if clusterResource.DataStore.Type == Redis {
			return fmt.Errorf("operation not supported for this datastore")
		}
		dbmsType = dbmsTypeCluster
	}

	var usersList userBatchCreateOpts

	u := userCreateOpts{
		Name:     userName,
		Password: d.Get("password").(string),
		Host:     d.Get("host").(string),
	}
	u.Databases, err = extractDatabaseUserDatabases(rawDatabases)
	if err != nil {
		return fmt.Errorf("unable to determine user`s databases")
	}
	usersList.Users = append(usersList.Users, u)

	err = userCreate(DatabaseV1Client, dbmsID, &usersList, dbmsType).ExtractErr()
	if err != nil {
		return fmt.Errorf("error creating mcs_db_user: %s", err)
	}

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"BUILD"},
		Target:     []string{"ACTIVE"},
		Refresh:    databaseUserStateRefreshFunc(DatabaseV1Client, dbmsID, userName, dbmsType),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      dbUserDelay,
		MinTimeout: dbUserMinTimeout,
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("error waiting for mcs_db_user %s to be created: %s", userName, err)
	}

	// Store the ID now
	d.SetId(fmt.Sprintf("%s/%s", dbmsID, userName))
	// Store dbms type
	d.Set("dbms_type", dbmsType)

	return resourceDatabaseUserRead(d, meta)
}

func resourceDatabaseUserRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(configer)
	DatabaseV1Client, err := config.DatabaseV1Client(getRegion(d, config))
	if err != nil {
		return fmt.Errorf("error creating mcs database client: %s", err)
	}

	userID := strings.SplitN(d.Id(), "/", 2)
	if len(userID) != 2 {
		return fmt.Errorf("invalid mcs_db_user ID: %s", d.Id())
	}

	dbmsID := userID[0]
	userName := userID[1]
	var dbmsType string
	if dbmsTypeRaw, ok := d.GetOk("dbms_type"); ok {
		dbmsType = dbmsTypeRaw.(string)
	} else {
		dbmsType = dbmsTypeInstance
	}

	exists, userObj, err := databaseUserExists(DatabaseV1Client, dbmsID, userName, dbmsType)
	if err != nil {
		return fmt.Errorf("error checking if mcs_db_user %s exists: %s", d.Id(), err)
	}

	if !exists {
		d.SetId("")
		return nil
	}

	d.Set("name", userName)

	databases := flattenDatabaseUserDatabases(userObj.Databases)
	if err := d.Set("databases", databases); err != nil {
		return fmt.Errorf("unable to set databases: %s", err)
	}

	if _, ok := d.GetOk("instance_id"); ok {
		d.Set("instance_id", dbmsID)
	}
	if _, ok := d.GetOk("dbms_id"); ok {
		d.Set("dbms_id", dbmsID)
	}

	return nil
}

func resourceDatabaseUserUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(configer)
	DatabaseV1Client, err := config.DatabaseV1Client(getRegion(d, config))
	if err != nil {
		return fmt.Errorf("error creating OpenStack database client: %s", err)
	}

	userID := strings.SplitN(d.Id(), "/", 2)
	if len(userID) != 2 {
		return fmt.Errorf("invalid mcs_db_user ID: %s", d.Id())
	}

	_, dbmsIDOk := d.GetOk("dbms_id")
	_, instanceIDOk := d.GetOk("instance_id")
	if !dbmsIDOk && !instanceIDOk {
		return fmt.Errorf("only dbms_id must be set")
	}

	dbmsID := userID[0]
	userName := userID[1]
	dbmsType := d.Get("dbms_type").(string)

	if d.HasChange("databases") {
		stateConf := &resource.StateChangeConf{
			Pending:    []string{"BUILD"},
			Target:     []string{"ACTIVE"},
			Refresh:    databaseUserStateRefreshFunc(DatabaseV1Client, dbmsID, userName, dbmsType),
			Timeout:    d.Timeout(schema.TimeoutCreate),
			Delay:      dbUserDelay,
			MinTimeout: dbUserMinTimeout,
		}

		oldDatabases, newDatabases := d.GetChange("databases")
		databasesForDeletion := make([]interface{}, 0)
		var exists bool
		for _, oldDatabase := range oldDatabases.([]interface{}) {
			exists = false
			for _, newDatabase := range newDatabases.([]interface{}) {
				if oldDatabase.(string) == newDatabase.(string) {
					exists = true
					break
				}
			}
			if !exists {
				databasesForDeletion = append(databasesForDeletion, oldDatabase)
			}
		}

		for _, databaseForDeletion := range databasesForDeletion {
			databaseName := databaseForDeletion.(string)
			err = userDeleteDatabase(DatabaseV1Client, dbmsID, userName, databaseName, dbmsType).ExtractErr()
			if err != nil {
				return fmt.Errorf("error deleting database from mcs_db_user: %s", err)
			}
		}
		newDatabasesOpts := make([]map[string]string, len(newDatabases.([]interface{})))
		for i, newDatabase := range newDatabases.([]interface{}) {
			newDatabasesOpts[i] = map[string]string{"name": newDatabase.(string)}
		}
		userUpdateDatabasesOpts := userUpdateDatabasesOpts{
			Databases: newDatabasesOpts,
		}
		err = userUpdateDatabases(DatabaseV1Client, dbmsID, userName, &userUpdateDatabasesOpts, dbmsType).ExtractErr()
		if err != nil {
			return fmt.Errorf("error adding databases to mcs_db_user: %s", err)
		}

		_, err = stateConf.WaitForState()
		if err != nil {
			return fmt.Errorf("error waiting for mcs_db_user %s to be updated: %s", userName, err)
		}
	}
	var userUpdateParams userUpdateOpts
	userUpdateParams.User.Name = userName

	if d.HasChange("name") {
		_, new := d.GetChange("name")
		userUpdateParams.User.Name = new.(string)
	}

	if d.HasChange("password") {
		_, new := d.GetChange("password")
		userUpdateParams.User.Password = new.(string)
	}

	if d.HasChange("host") {
		_, new := d.GetChange("host")
		userUpdateParams.User.Host = new.(string)
	}
	if d.HasChange("name") || d.HasChange("password") || d.HasChange("host") {
		stateConf := &resource.StateChangeConf{
			Pending:    []string{"BUILD"},
			Target:     []string{"ACTIVE"},
			Refresh:    databaseUserStateRefreshFunc(DatabaseV1Client, dbmsID, userUpdateParams.User.Name, dbmsType),
			Timeout:    d.Timeout(schema.TimeoutCreate),
			Delay:      dbUserDelay,
			MinTimeout: dbUserMinTimeout,
		}
		err = userUpdate(DatabaseV1Client, dbmsID, userName, &userUpdateParams, dbmsType).ExtractErr()
		if err != nil {
			return fmt.Errorf("error updating mcs_db_user: %s", err)
		}
		d.SetId(fmt.Sprintf("%s/%s", dbmsID, userUpdateParams.User.Name))
		_, err = stateConf.WaitForState()
		if err != nil {
			return fmt.Errorf("error waiting for mcs_db_user %s to be updated: %s", userName, err)
		}
	}

	return resourceDatabaseUserRead(d, meta)
}

func resourceDatabaseUserDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(configer)
	DatabaseV1Client, err := config.DatabaseV1Client(getRegion(d, config))
	if err != nil {
		return fmt.Errorf("error creating mcs database client: %s", err)
	}

	userID := strings.SplitN(d.Id(), "/", 2)
	if len(userID) != 2 {
		return fmt.Errorf("invalid mcs_db_user ID: %s", d.Id())
	}

	dbmsID := userID[0]
	userName := userID[1]
	dbmsType := d.Get("dbms_type").(string)

	exists, _, err := databaseUserExists(DatabaseV1Client, dbmsID, userName, dbmsType)
	if err != nil {
		return fmt.Errorf("error checking if mcs_db_user %s exists: %s", d.Id(), err)
	}

	if !exists {
		return nil
	}

	err = userDelete(DatabaseV1Client, dbmsID, userName, dbmsType).ExtractErr()
	if err != nil {
		return fmt.Errorf("error deleting mcs_db_user %s: %s", d.Id(), err)
	}

	return nil
}
