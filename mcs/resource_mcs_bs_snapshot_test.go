package mcs

import (
	"fmt"
	"testing"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack/blockstorage/v3/snapshots"
	"github.com/gophercloud/gophercloud/openstack/blockstorage/v3/volumes"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

const snapshotResourceFixture = `
		%s

		resource "mcs_blockstorage_snapshot" "%[2]s" {
			volume_id = mcs_blockstorage_volume.%s.id
			name = "%[2]s"
			description = "%[4]s"
			metadata = {
				"key1" : "value1"
				"key2" : "value2"
			}
		}`

func snapshotFixture(name, description string) *snapshots.CreateOpts {
	return &snapshots.CreateOpts{
		Name:        name,
		Description: description,
	}
}

func TestAccBlockStorageSnapshot_basic(t *testing.T) {
	var vol volumes.Volume
	var snapshot snapshots.Snapshot

	volumeName := "testvolume" + acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)
	createVolumeFixture := volumeFixture(volumeName, 8, osBSVolumeType, "DP1", "Test volume description")
	volumeResourceName := "mcs_blockstorage_volume." + volumeName

	snapshotName := "testsnapshot" + acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)
	createSnapshotFixture := snapshotFixture(snapshotName, "Test snapshot description")
	snapshotResourceName := "mcs_blockstorage_snapshot." + snapshotName

	updateSnapshotFixture := snapshotFixture(snapshotName, "Test snapshot description updated")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheckBlockStorage(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBlockStorageSnapshotDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckBlockStorageSnapshotBasic(volumeName, testAccCheckBlockStorageVolumeBasic(createVolumeFixture), createSnapshotFixture),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBlockStorageVolumeExists(volumeResourceName, &vol),
					testAccCheckBlockStorageSnapshotExists(snapshotResourceName, &snapshot),
					checkSnapshotAttrs(snapshotResourceName, createSnapshotFixture),
				),
			},
			{
				Config: testAccCheckBlockStorageSnapshotBasic(volumeName, testAccCheckBlockStorageVolumeBasic(createVolumeFixture), updateSnapshotFixture),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(snapshotResourceName, "description", updateSnapshotFixture.Description),
				),
			},
		},
	})
}

func checkSnapshotAttrs(resourceName string, vol *snapshots.CreateOpts) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if s.Empty() == true {
			return fmt.Errorf("state not updated")
		}

		checksStore := []resource.TestCheckFunc{
			resource.TestCheckResourceAttr(resourceName, "name", vol.Name),
			resource.TestCheckResourceAttr(resourceName, "description", vol.Description),
		}

		return resource.ComposeTestCheckFunc(checksStore...)(s)
	}
}

func testAccCheckBlockStorageSnapshotExists(n string, snapshot *snapshots.Snapshot) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("volume snapshot not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no id is set")
		}

		config := testAccProvider.Meta().(configer)
		BlockStorageV3Client, err := config.BlockStorageV3Client(osRegionName)
		if err != nil {
			return fmt.Errorf("error creating block storage client: %s", err)
		}

		found, err := snapshotGet(BlockStorageV3Client.(*gophercloud.ServiceClient), rs.Primary.ID).Extract()
		if err != nil {
			return err
		}

		if found.ID != rs.Primary.ID {
			return fmt.Errorf("volume not found")
		}

		*snapshot = *found
		return nil
	}
}

func testAccCheckBlockStorageSnapshotDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(configer)
	BlockStorageV3Client, err := config.BlockStorageV3Client(osRegionName)
	if err != nil {
		return fmt.Errorf("error creating block storage client: %s", err)
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "mcs_blockstorage_volume" {
			continue
		}

		_, err := volumeGet(BlockStorageV3Client.(*gophercloud.ServiceClient), rs.Primary.ID).Extract()
		if err == nil {
			return fmt.Errorf("volume still exists")
		}
	}

	return nil
}

func testAccCheckBlockStorageSnapshotBasic(volumeName, volumeResource string, fixture *snapshots.CreateOpts) string {
	return fmt.Sprintf(
		snapshotResourceFixture,
		volumeResource,
		fixture.Name,
		volumeName,
		fixture.Description,
	)
}
