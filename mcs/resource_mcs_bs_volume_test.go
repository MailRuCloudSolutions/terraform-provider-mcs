package mcs

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack/blockstorage/v3/volumes"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

const volumeResourceFixture = `
		resource "mcs_blockstorage_volume" "%[1]s" {
			name = "%[1]s"
			size = %d
			volume_type = "%s"
			availability_zone = "%s"
			description = "%s"
		}`

func volumeFixture(name string, size int, volumeType, availabilityZone, description string) *volumes.CreateOpts {
	return &volumes.CreateOpts{
		Name:             name,
		Size:             size,
		VolumeType:       volumeType,
		AvailabilityZone: availabilityZone,
		Description:      description,
	}
}

func TestAccBlockStorageVolume_basic(t *testing.T) {
	var vol volumes.Volume

	volumeName := "testvolume" + acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)
	createVolumeFixture := volumeFixture(volumeName, 8, osBSVolumeType, "DP1", "Test description")
	volumeResourceName := "mcs_blockstorage_volume." + volumeName

	updateVolumeFixture := volumeFixture(volumeName, 8, osBSVolumeType, "DP1", "Test description updated")

	resizeVolumeFixture := volumeFixture(volumeName, 10, osBSVolumeType, "DP1", "Test description updated")
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheckBlockStorage(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBlockStorageVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckBlockStorageVolumeBasic(createVolumeFixture),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBlockStorageVolumeExists(volumeResourceName, &vol),
					checkVolumeAttrs(volumeResourceName, createVolumeFixture),
				),
			},
			{
				Config: testAccCheckBlockStorageVolumeBasic(updateVolumeFixture),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(volumeResourceName, "description", updateVolumeFixture.Description),
				),
			},
			{
				Config: testAccCheckBlockStorageVolumeBasic(resizeVolumeFixture),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(volumeResourceName, "size", strconv.Itoa(resizeVolumeFixture.Size)),
				),
			},
		},
	})
}

func checkVolumeAttrs(resourceName string, vol *volumes.CreateOpts) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if s.Empty() == true {
			return fmt.Errorf("state not updated")
		}

		checksStore := []resource.TestCheckFunc{
			resource.TestCheckResourceAttr(resourceName, "name", vol.Name),
			resource.TestCheckResourceAttr(resourceName, "size", strconv.Itoa(vol.Size)),
			resource.TestCheckResourceAttr(resourceName, "volume_type", vol.VolumeType),
			resource.TestCheckResourceAttr(resourceName, "availability_zone", vol.AvailabilityZone),
			resource.TestCheckResourceAttr(resourceName, "description", vol.Description),
		}

		return resource.ComposeTestCheckFunc(checksStore...)(s)
	}
}

func testAccCheckBlockStorageVolumeExists(n string, vol *volumes.Volume) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("volume not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no id is set")
		}

		config := testAccProvider.Meta().(configer)
		BlockStorageV3Client, err := config.BlockStorageV3Client(osRegionName)
		if err != nil {
			return fmt.Errorf("error creating block storage client: %s", err)
		}

		found, err := volumeGet(BlockStorageV3Client.(*gophercloud.ServiceClient), rs.Primary.ID).Extract()
		if err != nil {
			return err
		}

		if found.ID != rs.Primary.ID {
			return fmt.Errorf("volume not found")
		}

		*vol = *found
		return nil
	}
}

func testAccCheckBlockStorageVolumeDestroy(s *terraform.State) error {
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

func testAccCheckBlockStorageVolumeBasic(fixture *volumes.CreateOpts) string {
	return fmt.Sprintf(
		volumeResourceFixture,
		fixture.Name,
		fixture.Size,
		fixture.VolumeType,
		fixture.AvailabilityZone,
		fixture.Description,
	)
}
