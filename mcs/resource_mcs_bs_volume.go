package mcs

import (
	"fmt"
	"log"
	"time"

	"github.com/gophercloud/gophercloud"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

const (
	bsVolumeCreateTimeout = 30 * time.Minute
	bsVolumeDelay         = 10 * time.Second
	bsVolumeMinTimeout    = 3 * time.Second
)

var (
	bsVolumeStatusBuild     = "creating"
	bsVolumeStatusActive    = "available"
	bsVolumeStatusShutdown  = "deleting"
	bsVolumeStatusDeleted   = "deleted"
	bsVolumeMigrationPolicy = "on-demand"
)

func resourceBlockStorageVolume() *schema.Resource {
	return &schema.Resource{
		Create: resourceBlockStorageVolumeCreate,
		Read:   resourceBlockStorageVolumeRead,
		Update: resourceBlockStorageVolumeUpdate,
		Delete: resourceBlockStorageVolumeDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(bsVolumeCreateTimeout),
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
			},

			"size": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: false,
			},

			"volume_type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: false,
			},

			"availability_zone": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: false,
			},

			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
			},

			"metadata": {
				Type: schema.TypeMap,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional: true,
				ForceNew: false,
			},

			"snapshot_id": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"source_volume_id", "image_id"},
			},

			"source_volume_id": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"snapshot_id", "image_id"},
			},

			"image_id": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"snapshot_id", "source_volume_id"},
			},
		},
	}
}

func resourceBlockStorageVolumeCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(configer)
	BlockStorageV3Client, err := config.BlockStorageV3Client(getRegion(d, config))
	if err != nil {
		return fmt.Errorf("error creating OpenStack block storage client: %s", err)
	}

	createOpts := &volumeCreateOpts{}
	createOpts.AvailabilityZone = d.Get("availability_zone").(string)
	createOpts.VolumeType = d.Get("volume_type").(string)
	createOpts.Size = d.Get("size").(int)

	if v, ok := d.GetOk("name"); ok {
		createOpts.Name = v.(string)
	}

	if v, ok := d.GetOk("description"); ok {
		createOpts.Description = v.(string)
	}

	rawMetadata := d.Get("metadata").(map[string]interface{})
	metadata, err := extractBlockStorageMetadataMap(rawMetadata)
	if err != nil {
		return err
	}
	createOpts.Metadata = metadata

	if v, ok := d.GetOk("snapshot_id"); ok {
		createOpts.SnapshotID = v.(string)
	}

	if v, ok := d.GetOk("source_volume_id"); ok {
		createOpts.SourceVolID = v.(string)
	}

	if v, ok := d.GetOk("image_id"); ok {
		createOpts.ImageID = v.(string)
	}

	log.Printf("[DEBUG] mcs_blockstorage_volume create options: %#v", createOpts)

	vol, err := volumeCreate(BlockStorageV3Client.(*gophercloud.ServiceClient), *createOpts).Extract()
	if err != nil {
		return fmt.Errorf("error creating mcs_blockstorage_volume: %s", err)
	}

	// Wait for the volume to become available.
	log.Printf("[DEBUG] Waiting for mcs_blockstorage_volume %s to become available", vol.ID)

	stateConf := &resource.StateChangeConf{
		Pending:    []string{bsVolumeStatusBuild},
		Target:     []string{bsVolumeStatusActive},
		Refresh:    blockStorageVolumeStateRefreshFunc(BlockStorageV3Client, vol.ID),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      bsVolumeDelay,
		MinTimeout: bsVolumeMinTimeout,
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("error waiting for mcs_blockstorage_volume %s to become ready: %s", vol.ID, err)
	}

	// Store the ID now
	d.SetId(vol.ID)

	return resourceBlockStorageVolumeRead(d, meta)
}

func resourceBlockStorageVolumeRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(configer)
	BlockStorageV3Client, err := config.BlockStorageV3Client(getRegion(d, config))
	if err != nil {
		return fmt.Errorf("error creating OpenStack block storage client: %s", err)
	}

	vol, err := volumeGet(BlockStorageV3Client.(*gophercloud.ServiceClient), d.Id()).Extract()
	if err != nil {
		return checkDeleted(d, err, "error retrieving mcs_blockstorage_volume")
	}

	log.Printf("[DEBUG] Retrieved mcs_blockstorage_volume %s: %#v", d.Id(), vol)

	rawMetadata := d.Get("metadata").(map[string]interface{})
	metadata, err := extractBlockStorageMetadataMap(rawMetadata)
	if err != nil {
		return err
	}

	if err := d.Set("metadata", metadata); err != nil {
		return fmt.Errorf("unable to set mcs_blockstorage_volume metadata: %s", err)
	}

	d.Set("name", vol.Name)
	d.Set("size", vol.Size)
	d.Set("volume_type", vol.VolumeType)
	d.Set("availability_zone", vol.AvailabilityZone)
	d.Set("description", vol.Description)
	d.Set("snapshot_id", vol.SnapshotID)
	d.Set("source_volume_id", vol.SourceVolID)

	return nil
}

func resourceBlockStorageVolumeUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(configer)
	BlockStorageV3Client, err := config.BlockStorageV3Client(getRegion(d, config))
	if err != nil {
		return fmt.Errorf("error creating OpenStack block storage client: %s", err)
	}

	stateConf := &resource.StateChangeConf{
		Pending:    []string{bsVolumeStatusBuild},
		Target:     []string{bsVolumeStatusActive},
		Refresh:    blockStorageVolumeStateRefreshFunc(BlockStorageV3Client, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      bsVolumeDelay,
		MinTimeout: bsVolumeMinTimeout,
	}

	updateOpts := volumeUpdateOpts{}

	if d.HasChange("name") {
		_, new := d.GetChange("name")
		newName := new.(string)
		updateOpts.Name = &newName
	}
	if d.HasChange("description") {
		_, new := d.GetChange("description")
		newDescription := new.(string)
		updateOpts.Description = &newDescription
	}

	if d.HasChange("metadata") {
		rawMetadata := d.Get("metadata").(map[string]interface{})
		metadata, err := extractBlockStorageMetadataMap(rawMetadata)
		if err != nil {
			return err
		}
		updateOpts.Metadata = metadata
	}
	_, err = volumeUpdate(BlockStorageV3Client.(*gophercloud.ServiceClient), d.Id(), updateOpts).Extract()
	if err != nil {
		return fmt.Errorf("error updating mcs_blockstorage_volume")
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("error waiting for mcs_blockstorage_volume %s to become ready: %s", d.Id(), err)
	}

	if d.HasChange("size") {
		_, new := d.GetChange("size")
		extendSizeOpts := volumeExtendSizeOpts{}
		extendSizeOpts.NewSize = new.(int)
		err = volumeExtendSize(BlockStorageV3Client.(*gophercloud.ServiceClient), d.Id(), extendSizeOpts).ExtractErr()
		if err != nil {
			return fmt.Errorf("error resizing mcs_blockstorage_volume")
		}
		_, err = stateConf.WaitForState()
		if err != nil {
			return fmt.Errorf("error waiting for mcs_blockstorage_volume %s to become ready: %s", d.Id(), err)
		}
	}

	var newVolumeType string
	var newAvailabilityZone string
	if d.HasChange("volume_type") {
		_, new := d.GetChange("volume_type")
		newVolumeType = new.(string)
	}
	if d.HasChange("availability_zone") {
		_, new := d.GetChange("availability_zone")
		newAvailabilityZone = new.(string)
	}

	changeTypeOpts := volumeChangeTypeOpts{}
	if newAvailabilityZone != "" {
		changeTypeOpts.AvailabilityZone = newAvailabilityZone
	}
	if newVolumeType != "" {
		changeTypeOpts.NewType = newVolumeType
	} else {
		changeTypeOpts.NewType = d.Get("volume_type").(string)
	}

	if newVolumeType != "" || newAvailabilityZone != "" {
		changeTypeOpts.MigrationPolicy = bsVolumeMigrationPolicy
		err = volumeChangeType(BlockStorageV3Client.(*gophercloud.ServiceClient), d.Id(), changeTypeOpts).ExtractErr()
		if err != nil {
			return fmt.Errorf("error changing type of mcs_blockstorage_volume")
		}
		_, err = stateConf.WaitForState()
		if err != nil {
			return fmt.Errorf("error waiting for mcs_blockstorage_volume %s to become ready: %s", d.Id(), err)
		}
	}

	return resourceBlockStorageVolumeRead(d, meta)
}

func resourceBlockStorageVolumeDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(configer)
	BlockStorageV3Client, err := config.BlockStorageV3Client(getRegion(d, config))
	if err != nil {
		return fmt.Errorf("error creating OpenStack block storage client: %s", err)
	}

	err = volumeDelete(BlockStorageV3Client.(*gophercloud.ServiceClient), d.Id()).ExtractErr()
	if err != nil {
		return checkDeleted(d, err, "Error deleting mcs_blockstorage_volume")
	}

	stateConf := &resource.StateChangeConf{
		Pending:    []string{bsVolumeStatusActive, bsVolumeStatusShutdown},
		Target:     []string{bsVolumeStatusDeleted},
		Refresh:    blockStorageVolumeStateRefreshFunc(BlockStorageV3Client, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      bsVolumeDelay,
		MinTimeout: bsVolumeMinTimeout,
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("error waiting for mcs_blockstorage_volume %s to delete : %s", d.Id(), err)
	}

	return nil
}
