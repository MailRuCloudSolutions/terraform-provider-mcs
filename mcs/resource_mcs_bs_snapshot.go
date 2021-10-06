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
	bsSnapshotCreateTimeout = 30 * time.Minute
	bsSnapshotDelay         = 10 * time.Second
	bsSnapshotMinTimeout    = 3 * time.Second
)

var (
	bsSnapshotStatusBuild    = "creating"
	bsSnapshotStatusActive   = "available"
	bsSnapshotStatusShutdown = "deleting"
	bsSnapshotStatusDeleted  = "deleted"
)

func resourceBlockStorageSnapshot() *schema.Resource {
	return &schema.Resource{
		Create: resourceBlockStorageSnapshotCreate,
		Read:   resourceBlockStorageSnapshotRead,
		Update: resourceBlockStorageSnapshotUpdate,
		Delete: resourceBlockStorageSnapshotDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(bsSnapshotCreateTimeout),
		},

		Schema: map[string]*schema.Schema{
			"volume_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"name": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
			},

			"force": {
				Type:     schema.TypeBool,
				Optional: true,
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
		},
	}
}

func resourceBlockStorageSnapshotCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(configer)
	BlockStorageV3Client, err := config.BlockStorageV3Client(getRegion(d, config))
	if err != nil {
		return fmt.Errorf("error creating OpenStack block storage client: %s", err)
	}

	createOpts := &snapshotCreateOpts{}
	createOpts.VolumeID = d.Get("volume_id").(string)

	if v, ok := d.GetOk("name"); ok {
		createOpts.Name = v.(string)
	}

	if v, ok := d.GetOk("force"); ok {
		createOpts.Force = v.(bool)
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

	log.Printf("[DEBUG] mcs_blockstorage_snapshot create options: %#v", createOpts)

	snapshot, err := snapshotCreate(BlockStorageV3Client.(*gophercloud.ServiceClient), *createOpts).Extract()
	if err != nil {
		return fmt.Errorf("error creating mcs_blockstorage_snapshot: %s", err)
	}

	// Wait for the volume snapshot to become available.
	log.Printf("[DEBUG] Waiting for mcs_blockstorage_volume_snapshot %s to become available", snapshot.ID)

	stateConf := &resource.StateChangeConf{
		Pending:    []string{bsSnapshotStatusBuild},
		Target:     []string{bsSnapshotStatusActive},
		Refresh:    blockStorageSnapshotStateRefreshFunc(BlockStorageV3Client, snapshot.ID),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      bsSnapshotDelay,
		MinTimeout: bsSnapshotMinTimeout,
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("error waiting for mcs_blockstorage_volume_snapshot %s to become ready: %s", snapshot.ID, err)
	}

	// Store the ID now
	d.SetId(snapshot.ID)

	return resourceBlockStorageSnapshotRead(d, meta)
}

func resourceBlockStorageSnapshotRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(configer)
	BlockStorageV3Client, err := config.BlockStorageV3Client(getRegion(d, config))
	if err != nil {
		return fmt.Errorf("error creating OpenStack block storage client: %s", err)
	}

	snapshot, err := snapshotGet(BlockStorageV3Client.(*gophercloud.ServiceClient), d.Id()).Extract()
	if err != nil {
		return checkDeleted(d, err, "error retrieving mcs_blockstorage_snapshot")
	}

	log.Printf("[DEBUG] Retrieved mcs_blockstorage_snapshot %s: %#v", d.Id(), snapshot)

	rawMetadata := d.Get("metadata").(map[string]interface{})
	metadata, err := extractBlockStorageMetadataMap(rawMetadata)
	if err != nil {
		return err
	}

	if err := d.Set("metadata", metadata); err != nil {
		return fmt.Errorf("unable to set mcs_blockstorage_snapshot metadata: %s", err)
	}

	d.Set("name", snapshot.Name)
	d.Set("description", snapshot.Description)
	d.Set("volume_id", snapshot.VolumeID)

	return nil
}

func resourceBlockStorageSnapshotUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(configer)
	BlockStorageV3Client, err := config.BlockStorageV3Client(getRegion(d, config))
	if err != nil {
		return fmt.Errorf("error creating OpenStack block storage client: %s", err)
	}

	stateConf := &resource.StateChangeConf{
		Pending:    []string{bsSnapshotStatusBuild},
		Target:     []string{bsSnapshotStatusActive},
		Refresh:    blockStorageSnapshotStateRefreshFunc(BlockStorageV3Client, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      bsSnapshotDelay,
		MinTimeout: bsSnapshotMinTimeout,
	}

	updateOpts := snapshotUpdateOpts{}

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

	_, err = snapshotUpdate(BlockStorageV3Client.(*gophercloud.ServiceClient), d.Id(), updateOpts).Extract()
	if err != nil {
		return fmt.Errorf("error updating mcs_blockstorage_snapshot")
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("error waiting for mcs_blockstorage_snapshot %s to become ready: %s", d.Id(), err)
	}

	if d.HasChange("metadata") {
		updateMetadataOpts := snapshotUpdateMetadataOpts{}
		rawMetadata := d.Get("metadata").(map[string]interface{})
		updateMetadataOpts.Metadata = rawMetadata
		_, err = snapshotUpdateMetadata(BlockStorageV3Client.(*gophercloud.ServiceClient), d.Id(), updateMetadataOpts).Extract()
		if err != nil {
			return fmt.Errorf("error updating mcs_blockstorage_snapshot metadata")
		}

		_, err = stateConf.WaitForState()
		if err != nil {
			return fmt.Errorf("error waiting for mcs_blockstorage_snapshot %s to become ready: %s", d.Id(), err)
		}
	}

	return resourceBlockStorageSnapshotRead(d, meta)
}

func resourceBlockStorageSnapshotDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(configer)
	BlockStorageV3Client, err := config.BlockStorageV3Client(getRegion(d, config))
	if err != nil {
		return fmt.Errorf("error creating OpenStack block storage client: %s", err)
	}
	err = snapshotDelete(BlockStorageV3Client.(*gophercloud.ServiceClient), d.Id()).ExtractErr()
	if err != nil {
		return checkDeleted(d, err, "Error deleting mcs_blockstorage_snapshot")
	}

	stateConf := &resource.StateChangeConf{
		Pending:    []string{bsSnapshotStatusActive, bsSnapshotStatusShutdown},
		Target:     []string{bsSnapshotStatusDeleted},
		Refresh:    blockStorageVolumeStateRefreshFunc(BlockStorageV3Client, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      bsSnapshotDelay,
		MinTimeout: bsSnapshotMinTimeout,
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("error waiting for mcs_blockstorage_snapshot %s to delete : %s", d.Id(), err)
	}

	return nil
}
