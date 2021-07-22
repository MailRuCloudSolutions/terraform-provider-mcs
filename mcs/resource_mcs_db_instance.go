package mcs

import (
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

// Dbaas timeouts
const (
	DBInstanceDelay         = 10 * time.Second
	DBInstanceMinTimeout    = 3 * time.Second
	DBDatabaseDelay         = 10 * time.Second
	DBDatabaseMinTimeout    = 3 * time.Second
	DBUserDelay             = 10 * time.Second
	DBUserMinTimeout        = 3 * time.Second
	DBCreateTimeout         = 30 * time.Minute
	DBDeleteTimeout         = 30 * time.Minute
	DBUserCreateTimeout     = 10 * time.Minute
	DBUserDeleteTimeout     = 10 * time.Minute
	DBDatabaseCreateTimeout = 10 * time.Minute
	DBDatabaseDeleteTimeout = 10 * time.Minute
)

var dbstatus = Status{
	DELETED:  "DELETED",
	BUILD:    "BUILD",
	ACTIVE:   "ACTIVE",
	SHUTDOWN: "SHUTDOWN",
	RESIZE:   "RESIZE",
	DETACH:   "DETACH",
}

func resourceDatabaseInstance() *schema.Resource {
	return &schema.Resource{
		Create: resourceDatabaseInstanceCreate,
		Read:   resourceDatabaseInstanceRead,
		Delete: resourceDatabaseInstanceDelete,
		Update: resourceDatabaseInstanceUpdate,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(DBCreateTimeout),
			Delete: schema.DefaultTimeout(DBDeleteTimeout),
		},

		Schema: map[string]*schema.Schema{
			"region": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"flavor_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: false,
				Computed: false,
			},

			"size": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: false,
				Computed: false,
			},

			"volume_type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				Computed: false,
			},

			"wal_volume": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"size": {
							Type:     schema.TypeInt,
							Required: true,
							ForceNew: false,
						},
						"volume_type": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"autoexpand": {
							Type:     schema.TypeBool,
							Optional: true,
							ForceNew: false,
						},
						"max_disk_size": {
							Type:     schema.TypeInt,
							Optional: true,
							ForceNew: false,
						},
					},
				},
			},

			"datastore": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"version": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"type": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
					},
				},
			},

			"network": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"uuid": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"port": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"fixed_ip_v4": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
					},
				},
			},

			"configuration_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: false,
				ForceNew: false,
			},

			"replica_of": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: false,
				ForceNew: false,
			},

			"root_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: false,
			},

			"root_password": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
				Computed:  true,
				ForceNew:  false,
			},

			"availability_zone": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: false,
				ForceNew: true,
			},

			"floating_ip_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: false,
				ForceNew: true,
			},

			"keypair": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: false,
				ForceNew: true,
			},

			"disk_autoexpand": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: false,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"autoexpand": {
							Type:     schema.TypeBool,
							Optional: true,
							ForceNew: false,
						},
						"max_disk_size": {
							Type:     schema.TypeInt,
							Optional: true,
							ForceNew: false,
						},
					},
				},
			},

			"capabilities": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: false,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"settings": {
							Type:     schema.TypeMap,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      schema.HashString,
						},
					},
				},
			},
		},
		CustomizeDiff: customdiff.All(
			customdiff.ValidateChange("size", func(old, new, meta interface{}) error {
				if new.(int) < old.(int) {
					return fmt.Errorf("the new volume size %d must be larger than the current volume size of %d", new.(int), old.(int))
				}
				return nil
			}),
			customdiff.ValidateChange("wal_volume", func(old, new, meta interface{}) error {
				if len(old.([]interface{})) == 0 {
					return nil
				}

				walVolumeOptsNew, err := extractDatabaseInstanceWalVolume(new.([]interface{}))
				if err != nil {
					return fmt.Errorf("unable to determine mcs_db_instance wal_volume")
				}

				walVolumeOptsOld, err := extractDatabaseInstanceWalVolume(old.([]interface{}))
				if err != nil {
					return fmt.Errorf("unable to determine mcs_db_instance wal_volume")
				}

				if walVolumeOptsNew.Size < walVolumeOptsOld.Size {
					return fmt.Errorf("the new wal volume size %d must be larger than the current volume size of %d", walVolumeOptsNew.Size, walVolumeOptsOld.Size)
				}
				return nil
			}),
		),
	}
}

func resourceDatabaseInstanceCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(Config)
	DatabaseV1Client, err := config.DatabaseV1Client(GetRegion(d, config))
	if err != nil {
		return fmt.Errorf("error creating OpenStack database client: %s", err)
	}

	size := d.Get("size").(int)
	createOpts := &dbInstanceCreateOpts{
		FlavorRef:         d.Get("flavor_id").(string),
		Name:              d.Get("name").(string),
		Volume:            &volume{Size: &size, VolumeType: d.Get("volume_type").(string)},
		ReplicaOf:         d.Get("replica_of").(string),
		AvailabilityZone:  d.Get("availability_zone").(string),
		FloatingIPEnabled: d.Get("floating_ip_enabled").(bool),
		Keypair:           d.Get("keypair").(string),
	}

	message := "unable to determine mcs_db_instance"
	if v, ok := d.GetOk("datastore"); ok {
		datastore, err := extractDatabaseInstanceDatastore(v.([]interface{}))
		if err != nil {
			return fmt.Errorf("%s datastore", message)
		}
		createOpts.Datastore = &datastore
	}

	if replicaOf, ok := d.GetOk("replica_of"); ok {
		if createOpts.Datastore.Type == PostgresPro {
			return fmt.Errorf("replica_of field is forbidden for PostgresPro")
		}
		createOpts.ReplicaOf = replicaOf.(string)
	}

	if v, ok := d.GetOk("network"); ok {
		createOpts.Nics, err = extractDatabaseInstanceNetworks(v.([]interface{}))
		if err != nil {
			return fmt.Errorf("%s network", message)
		}
	}

	if v, ok := d.GetOk("disk_autoexpand"); ok {
		autoExpandOpts, err := extractDatabaseInstanceAutoExpand(v.([]interface{}))
		if err != nil {
			return fmt.Errorf("%s disk_autoexpand", message)
		}
		if autoExpandOpts.AutoExpand {
			createOpts.AutoExpand = 1
		} else {
			createOpts.AutoExpand = 0
		}
		createOpts.MaxDiskSize = autoExpandOpts.MaxDiskSize
	}

	if v, ok := d.GetOk("wal_volume"); ok {
		walVolumeOpts, err := extractDatabaseInstanceWalVolume(v.([]interface{}))
		if err != nil {
			return fmt.Errorf("%s wal_volume", message)
		}
		createOpts.Walvolume = &walVolume{
			Size:        &walVolumeOpts.Size,
			VolumeType:  walVolumeOpts.VolumeType,
			MaxDiskSize: walVolumeOpts.MaxDiskSize,
		}
		if walVolumeOpts.AutoExpand {
			createOpts.Walvolume.AutoExpand = 1
		} else {
			createOpts.Walvolume.AutoExpand = 0
		}
	}

	if capabilities, ok := d.GetOk("capabilities"); ok {
		capabilitiesOpts, err := extractDatabaseInstanceCapabilities(capabilities.([]interface{}))
		if err != nil {
			return fmt.Errorf("%s capability", message)
		}
		createOpts.Capabilities = capabilitiesOpts
	}

	log.Printf("[DEBUG] mcs_db_instance create options: %#v", createOpts)

	inst := dbInstance{}
	inst.Instance = createOpts

	instance, err := instanceCreate(DatabaseV1Client, inst).extract()
	if err != nil {
		return fmt.Errorf("error creating mcs_db_instance: %s", err)
	}

	// Wait for the instance to become available.
	log.Printf("[DEBUG] Waiting for mcs_db_instance %s to become available", instance.ID)

	stateConf := &resource.StateChangeConf{
		Pending:    []string{dbstatus.BUILD},
		Target:     []string{dbstatus.ACTIVE},
		Refresh:    databaseInstanceStateRefreshFunc(DatabaseV1Client, instance.ID),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      DBInstanceDelay,
		MinTimeout: DBInstanceMinTimeout,
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("error waiting for mcs_db_instance %s to become ready: %s", instance.ID, err)
	}

	if configuration, ok := d.GetOk("configuration_id"); ok {
		log.Printf("[DEBUG] Attaching configuration %s to mcs_db_instance %s", configuration, instance.ID)
		var attachConfigurationOpts instanceAttachConfigurationGroupOpts
		attachConfigurationOpts.Instance.Configuration = configuration.(string)
		err := instanceAttachConfigurationGroup(DatabaseV1Client, instance.ID, &attachConfigurationOpts).ExtractErr()
		if err != nil {
			return fmt.Errorf("error attaching configuration group %s to mcs_db_instance %s: %s",
				configuration, instance.ID, err)
		}
	}

	if rootEnabled, ok := d.GetOk("root_enabled"); ok {
		if rootEnabled.(bool) {
			rootPassword := d.Get("root_password")
			var rootUserEnableOpts instanceRootUserEnableOpts
			if rootPassword != "" {
				rootUserEnableOpts.Password = rootPassword.(string)
			}
			rootUser, err := instanceRootUserEnable(DatabaseV1Client, instance.ID, &rootUserEnableOpts).extract()
			if err != nil {
				return fmt.Errorf("error creating root user for instance: %s: %s", instance.ID, err)
			}
			d.Set("root_password", rootUser.Password)
		}
	}

	// Store the ID now
	d.SetId(instance.ID)

	return resourceDatabaseInstanceRead(d, meta)
}

func resourceDatabaseInstanceRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(Config)
	DatabaseV1Client, err := config.DatabaseV1Client(GetRegion(d, config))
	if err != nil {
		return fmt.Errorf("error creating OpenStack database client: %s", err)
	}

	instance, err := instanceGet(DatabaseV1Client, d.Id()).extract()
	if err != nil {
		return CheckDeleted(d, err, "Error retrieving mcs_db_instance")
	}

	log.Printf("[DEBUG] Retrieved mcs_db_instance %s: %#v", d.Id(), instance)

	d.Set("name", instance.Name)
	d.Set("flavor_id", instance.Flavor)
	d.Set("datastore", instance.DataStore)
	d.Set("region", GetRegion(d, config))
	if instance.ReplicaOf != nil {
		d.Set("replica_of", instance.ReplicaOf.ID)
	}
	isRootEnabledResult := instanceRootUserGet(DatabaseV1Client, d.Id())
	isRootEnabled, err := isRootEnabledResult.extract()
	if err != nil {
		return fmt.Errorf("error checking if root user is enabled for instance: %s: %s", d.Id(), err)
	}
	if isRootEnabled {
		d.Set("root_enabled", true)
	}

	return nil
}

func resourceDatabaseInstanceUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(Config)
	DatabaseV1Client, err := config.DatabaseV1Client(GetRegion(d, config))
	if err != nil {
		return fmt.Errorf("error creating OpenStack database client: %s", err)
	}

	stateConf := &resource.StateChangeConf{
		Pending:    []string{dbstatus.BUILD},
		Target:     []string{dbstatus.ACTIVE},
		Refresh:    databaseInstanceStateRefreshFunc(DatabaseV1Client, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      DBInstanceDelay,
		MinTimeout: DBInstanceMinTimeout,
	}

	if d.HasChange("configuration_id") {
		old, new := d.GetChange("configuration_id")

		err := instanceDetachConfigurationGroup(DatabaseV1Client, d.Id()).ExtractErr()
		if err != nil {
			return err
		}
		log.Printf("Detaching configuration %s from mcs_db_instance %s", old, d.Id())

		_, err = stateConf.WaitForState()
		if err != nil {
			return fmt.Errorf("error waiting for mcs_db_instance %s to become ready: %s", d.Id(), err)
		}

		if new != "" {
			var attachConfigurationOpts instanceAttachConfigurationGroupOpts
			attachConfigurationOpts.Instance.Configuration = new.(string)
			err := instanceAttachConfigurationGroup(DatabaseV1Client, d.Id(), &attachConfigurationOpts).ExtractErr()
			if err != nil {
				return err
			}
			log.Printf("Attaching configuration %s to mcs_db_instance %s", new, d.Id())

			_, err = stateConf.WaitForState()
			if err != nil {
				return fmt.Errorf("error waiting for mcs_db_instance %s to become ready: %s", d.Id(), err)
			}
		}
	}

	if d.HasChange("size") {
		_, new := d.GetChange("size")
		var resizeVolumeOpts instanceResizeVolumeOpts
		resizeVolumeOpts.Resize.Volume.Size = new.(int)
		err := instanceAction(DatabaseV1Client, d.Id(), &resizeVolumeOpts).ExtractErr()
		if err != nil {
			return err
		}
		log.Printf("Resizing volume from mcs_db_instance %s", d.Id())

		stateConf.Pending = []string{dbstatus.RESIZE}
		stateConf.Target = []string{dbstatus.ACTIVE}

		_, err = stateConf.WaitForState()
		if err != nil {
			return fmt.Errorf("error waiting for mcs_db_instance %s to become ready: %s", d.Id(), err)
		}
	}

	if d.HasChange("flavor_id") {
		var resizeOpts instanceResizeOpts
		resizeOpts.Resize.FlavorRef = d.Get("flavor_id").(string)
		err := instanceAction(DatabaseV1Client, d.Id(), &resizeOpts).ExtractErr()
		if err != nil {
			return err
		}
		log.Printf("Resizing flavor from mcs_db_instance %s", d.Id())

		stateConf.Pending = []string{dbstatus.RESIZE}
		stateConf.Target = []string{dbstatus.ACTIVE}

		_, err = stateConf.WaitForState()
		if err != nil {
			return fmt.Errorf("error waiting for mcs_db_instance %s to become ready: %s", d.Id(), err)
		}
	}

	if d.HasChange("replica_of") {
		old, new := d.GetChange("replica_of")
		if old != "" && new == "" {
			detachReplicaOpts := &instanceDetachReplicaOpts{}
			detachReplicaOpts.Instance.ReplicaOf = old.(string)
			err := instanceDetachReplica(DatabaseV1Client, d.Id(), detachReplicaOpts).ExtractErr()
			if err != nil {
				return err
			}
			log.Printf("Detach replica from mcs_db_instance %s", d.Id())

			stateConf.Pending = []string{dbstatus.DETACH}
			stateConf.Target = []string{dbstatus.ACTIVE}

			_, err = stateConf.WaitForState()
			if err != nil {
				return fmt.Errorf("error waiting for mcs_db_instance %s to become ready: %s", d.Id(), err)
			}
		}
	}

	if d.HasChange("root_enabled") {
		_, new := d.GetChange("root_enabled")
		if new == true {
			rootPassword := d.Get("root_password")
			var rootUserEnableOpts instanceRootUserEnableOpts
			if rootPassword != "" {
				rootUserEnableOpts.Password = rootPassword.(string)
			}

			rootUser, err := instanceRootUserEnable(DatabaseV1Client, d.Id(), &rootUserEnableOpts).extract()
			if err != nil {
				return fmt.Errorf("error creating root user for instance: %s: %s", d.Id(), err)
			}
			d.Set("root_password", rootUser.Password)
		} else {
			err = instanceRootUserDisable(DatabaseV1Client, d.Id()).ExtractErr()
			if err != nil {
				return fmt.Errorf("error deleting root_user for instance %s: %s", d.Id(), err)
			}
		}
	}

	if d.HasChange("disk_autoexpand") {
		_, new := d.GetChange("disk_autoexpand")
		autoExpandProperties, err := extractDatabaseInstanceAutoExpand(new.([]interface{}))
		if err != nil {
			return fmt.Errorf("unable to determine mcs_db_instance disk_autoexpand")
		}
		var autoExpandOpts instanceUpdateAutoExpandOpts
		if autoExpandProperties.AutoExpand {
			autoExpandOpts.Instance.VolumeAutoresizeEnabled = 1
		} else {
			autoExpandOpts.Instance.VolumeAutoresizeEnabled = 0
		}
		autoExpandOpts.Instance.VolumeAutoresizeMaxSize = autoExpandProperties.MaxDiskSize
		err = instanceUpdateAutoExpand(DatabaseV1Client, d.Id(), &autoExpandOpts).ExtractErr()
		if err != nil {
			return err
		}

		stateConf.Pending = []string{dbstatus.BUILD}
		stateConf.Target = []string{dbstatus.ACTIVE}

		_, err = stateConf.WaitForState()
		if err != nil {
			return fmt.Errorf("error waiting for mcs_db_instance %s to become ready: %s", d.Id(), err)
		}
	}

	if d.HasChange("wal_volume") {
		old, new := d.GetChange("wal_volume")
		walVolumeOptsNew, err := extractDatabaseInstanceWalVolume(new.([]interface{}))
		if err != nil {
			return fmt.Errorf("unable to determine mcs_db_instance wal_volume")
		}

		walVolumeOptsOld, err := extractDatabaseInstanceWalVolume(old.([]interface{}))
		if err != nil {
			return fmt.Errorf("unable to determine mcs_db_instance wal_volume")
		}

		if walVolumeOptsNew.Size != walVolumeOptsOld.Size {
			var resizeWalVolumeOpts instanceResizeWalVolumeOpts
			resizeWalVolumeOpts.Resize.Volume.Size = walVolumeOptsNew.Size
			resizeWalVolumeOpts.Resize.Volume.Kind = "wal"
			err = instanceAction(DatabaseV1Client, d.Id(), &resizeWalVolumeOpts).ExtractErr()
			if err != nil {
				return err
			}

			stateConf.Pending = []string{dbstatus.RESIZE}
			stateConf.Target = []string{dbstatus.ACTIVE}

			_, err = stateConf.WaitForState()
			if err != nil {
				return fmt.Errorf("error waiting for mcs_db_instance %s to become ready: %s", d.Id(), err)
			}
		}

		// Wal volume autoresize params update
		var autoExpandWalOpts instanceUpdateAutoExpandWalOpts
		if walVolumeOptsNew.AutoExpand {
			autoExpandWalOpts.Instance.WalVolume.VolumeAutoresizeEnabled = 1
		} else {
			autoExpandWalOpts.Instance.WalVolume.VolumeAutoresizeEnabled = 0
		}
		autoExpandWalOpts.Instance.WalVolume.VolumeAutoresizeMaxSize = walVolumeOptsNew.MaxDiskSize
		err = instanceUpdateAutoExpand(DatabaseV1Client, d.Id(), &autoExpandWalOpts).ExtractErr()
		if err != nil {
			return err
		}
	}

	if d.HasChange("capabilities") {
		_, newCapabilities := d.GetChange("capabilities")
		newCapabilitiesOpts, err := extractDatabaseInstanceCapabilities(newCapabilities.([]interface{}))
		if err != nil {
			return fmt.Errorf("unable to determine mcs_db_instance capability")
		}
		var applyCapabilityOpts instanceApplyCapabilityOpts
		applyCapabilityOpts.ApplyCapability.Capabilities = newCapabilitiesOpts

		err = instanceAction(DatabaseV1Client, d.Id(), &applyCapabilityOpts).ExtractErr()

		if err != nil {
			return fmt.Errorf("error applying capability to mcs_db_instance %s: %s", d.Id(), err)
		}
	}

	return resourceDatabaseInstanceRead(d, meta)
}

func resourceDatabaseInstanceDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(Config)
	DatabaseV1Client, err := config.DatabaseV1Client(GetRegion(d, config))
	if err != nil {
		return fmt.Errorf("error creating OpenStack database client: %s", err)
	}

	err = instanceDelete(DatabaseV1Client, d.Id()).ExtractErr()
	if err != nil {
		return CheckDeleted(d, err, "Error deleting mcs_db_instance")
	}

	stateConf := &resource.StateChangeConf{
		Pending:    []string{dbstatus.ACTIVE, dbstatus.SHUTDOWN},
		Target:     []string{dbstatus.DELETED},
		Refresh:    databaseInstanceStateRefreshFunc(DatabaseV1Client, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      DBInstanceDelay,
		MinTimeout: DBInstanceMinTimeout,
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("error waiting for mcs_db_instance %s to delete: %s", d.Id(), err)
	}

	return nil
}
