package mcs

import (
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack/blockstorage/extensions/volumeactions"
	"github.com/gophercloud/gophercloud/openstack/blockstorage/v3/snapshots"
	"github.com/gophercloud/gophercloud/openstack/blockstorage/v3/volumes"
)

type volumeCreateOpts struct {
	volumes.CreateOpts
}

type volumeUpdateOpts struct {
	volumes.UpdateOpts
}

type volumeExtendSizeOpts struct {
	volumeactions.ExtendSizeOpts
}

type volumeChangeTypeOpts struct {
	NewType          string `json:"new_type,omitempty"`
	AvailabilityZone string `json:"availability_zone,omitempty"`
	MigrationPolicy  string `json:"migration_policy,omitempty"`
}

type snapshotCreateOpts struct {
	snapshots.CreateOpts
}

type snapshotUpdateMetadataOpts struct {
	snapshots.UpdateMetadataOpts
}

type snapshotUpdateOpts struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}

func (opts volumeChangeTypeOpts) ToVolumeChangeTypeMap() (map[string]interface{}, error) {
	return gophercloud.BuildRequestBody(opts, "os-retype")
}

func (opts snapshotUpdateOpts) ToSnapshotUpdateMap() (map[string]interface{}, error) {
	return gophercloud.BuildRequestBody(opts, "snapshot")
}

type snapshotUpdateResult struct {
	commonResult
}

func (r snapshotUpdateResult) Extract() (*snapshots.Snapshot, error) {
	var s snapshots.Snapshot
	err := r.ExtractInto(&s)
	return &s, err
}

func snapshotUpdateURL(c *gophercloud.ServiceClient, id string) string {
	return c.ServiceURL("snapshots", id)
}

func volumeCreate(client *gophercloud.ServiceClient, opts volumeCreateOpts) (r volumes.CreateResult) {
	return volumes.Create(client, opts)
}

func volumeGet(client *gophercloud.ServiceClient, id string) (r volumes.GetResult) {
	return volumes.Get(client, id)
}

func volumeUpdate(client *gophercloud.ServiceClient, id string, opts volumeUpdateOpts) (r volumes.UpdateResult) {
	return volumes.Update(client, id, opts)
}

func volumeDelete(client *gophercloud.ServiceClient, id string) (r volumes.DeleteResult) {
	return volumes.Delete(client, id, nil)
}

func volumeExtendSize(client *gophercloud.ServiceClient, id string, opts volumeExtendSizeOpts) (r volumeactions.ExtendSizeResult) {
	return volumeactions.ExtendSize(client, id, opts)
}

func volumeChangeType(client *gophercloud.ServiceClient, id string, opts volumeChangeTypeOpts) (r volumeactions.ChangeTypeResult) {
	return volumeactions.ChangeType(client, id, opts)
}

func snapshotCreate(client *gophercloud.ServiceClient, opts snapshotCreateOpts) (r snapshots.CreateResult) {
	return snapshots.Create(client, opts)
}

func snapshotGet(client *gophercloud.ServiceClient, id string) (r snapshots.GetResult) {
	return snapshots.Get(client, id)
}

func snapshotDelete(client *gophercloud.ServiceClient, id string) (r snapshots.DeleteResult) {
	return snapshots.Delete(client, id)
}

func snapshotUpdate(client *gophercloud.ServiceClient, id string, opts snapshotUpdateOpts) (r snapshotUpdateResult) {
	b, err := opts.ToSnapshotUpdateMap()
	if err != nil {
		r.Err = err
		return
	}
	resp, err := client.Put(snapshotUpdateURL(client, id), b, &r.Body, &gophercloud.RequestOpts{
		OkCodes: []int{200},
	})
	_, r.Header, r.Err = gophercloud.ParseResponse(resp, err)
	return
}

func snapshotUpdateMetadata(client *gophercloud.ServiceClient, id string, opts snapshotUpdateMetadataOpts) (r snapshots.UpdateMetadataResult) {
	return snapshots.UpdateMetadata(client, id, opts)
}
