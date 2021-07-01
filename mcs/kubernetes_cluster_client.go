package mcs

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack/containerinfra/v1/clusters"
	"github.com/gophercloud/gophercloud/openstack/containerinfra/v1/clustertemplates"
)

//ContainerClient is interface to work with gopherclod requests
type ContainerClient interface {
	Get(url string, JSONResponse interface{}, opts *gophercloud.RequestOpts) (*http.Response, error)
	Post(url string, JSONBody interface{}, JSONResponse interface{}, opts *gophercloud.RequestOpts) (*http.Response, error)
	Patch(url string, JSONBody interface{}, JSONResponse interface{}, opts *gophercloud.RequestOpts) (*http.Response, error)
	Delete(url string, opts *gophercloud.RequestOpts) (*http.Response, error)
	Head(url string, opts *gophercloud.RequestOpts) (*http.Response, error)
	Put(url string, JSONBody interface{}, JSONResponse interface{}, opts *gophercloud.RequestOpts) (*http.Response, error)
	ServiceURL(parts ...string) string
}

const magnumAPIMicroVersion = "1.16"

var magnumAPIMicroVersionHeader = map[string]string{
	"MCS-API-Version": fmt.Sprintf("container-infra %s", magnumAPIMicroVersion),
}

func addMicroVersionHeader(reqOpts *gophercloud.RequestOpts) {
	reqOpts.MoreHeaders = magnumAPIMicroVersionHeader
}

//Node ...
type Node struct {
	Name        string     `json:"name"`
	UUID        string     `json:"uuid"`
	NodeGroupID string     `json:"node_group_id"`
	CreatedAt   *time.Time `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty"`
}

type nodesFlatSchema []map[string]interface{}

func flattenNodes(nodes []*Node) nodesFlatSchema {
	flatSchema := nodesFlatSchema{}
	for _, node := range nodes {
		flatSchema = append(flatSchema, map[string]interface{}{
			"name":          node.Name,
			"uuid":          node.UUID,
			"node_group_id": node.NodeGroupID,
			"created_at":    GetTimestamp(node.CreatedAt),
			"updated_at":    GetTimestamp(node.UpdatedAt),
		})
	}
	return flatSchema
}

//NodeGroupClusterPatchOpts ...
type NodeGroupClusterPatchOpts []NodeGroupPatchParams

//NodeGroupPatchParams ...
type NodeGroupPatchParams struct {
	Path  string      `json:"path,omitempty"`
	Value interface{} `json:"value,omitempty"`
	Op    string      `json:"op,omitempty"`
}

//NodeGroupBatchDelParams ...
type NodeGroupBatchDelParams struct {
	Action  string   `json:"action,omitempty"`
	Payload []string `json:"payload,omitempty"`
}

//NodeGroupBatchAddParams ...
type NodeGroupBatchAddParams struct {
	Action  string      `json:"action,omitempty"`
	Payload []NodeGroup `json:"payload,omitempty"`
}

//NodeGroups ...
type NodeGroups struct {
	NodeGroups []NodeGroup `json:"node_groups"`
}

//NodeGroup ...
type NodeGroup struct {
	Name        string    `json:"name,omitempty"`
	NodeCount   int       `json:"node_count,omitempty"`
	MaxNodes    int       `json:"max_nodes,omitempty"`
	MinNodes    int       `json:"min_nodes,omitempty"`
	VolumeSize  int       `json:"volume_size,omitempty"`
	VolumeType  string    `json:"volume_type,omitempty"`
	FlavorID    string    `json:"flavor_id,omitempty"`
	ImageID     string    `json:"image_id,omitempty"`
	Autoscaling bool      `json:"autoscaling_enabled,omitempty"`
	ClusterID   string    `json:"cluster_id,omitempty"`
	UUID        string    `json:"uuid,omitempty"`
	CreatedAt   time.Time `json:"created_at,omitempty"`
	UpdatedAt   time.Time `json:"updated_at,omitempty"`
	Nodes       []*Node   `json:"nodes,omitempty"`
	State       string    `json:"state,omitempty"`
}

//NodeGroupLabel ...
type NodeGroupLabel struct {
	Key   string `json:"key"`
	Value string `json:"value,omitempty"`
}

//NodeGroupTaint ...
type NodeGroupTaint struct {
	Key    string `json:"key,omitempty"`
	Value  string `json:"value,omitempty"`
	Effect string `json:"effect,omitempty"`
}

//NodeGroupCreateOpts ...
type NodeGroupCreateOpts struct {
	ClusterID   string           `json:"cluster_id" required:"true"`
	Name        string           `json:"name"`
	Labels      []NodeGroupLabel `json:"labels,omitempty"`
	Taints      []NodeGroupTaint `json:"taints,omitempty"`
	NodeCount   int              `json:"node_count,omitempty"`
	MaxNodes    int              `json:"max_nodes,omitempty"`
	MinNodes    int              `json:"min_nodes,omitempty"`
	VolumeSize  int              `json:"volume_size,omitempty"`
	VolumeType  string           `json:"volume_type,omitempty"`
	FlavorID    string           `json:"flavor_id,omitempty"`
	Autoscaling bool             `json:"autoscaling_enabled,omitempty"`
}

//NodeGroupScaleOpts ...
type NodeGroupScaleOpts struct {
	Delta    int    `json:"delta" required:"true"`
	Rollback string `json:"rollback,omitempty"`
}

//ClusterCreateOpts ...
type ClusterCreateOpts struct {
	ClusterTemplateID    string            `json:"cluster_template_id" required:"true"`
	Keypair              string            `json:"keypair,omitempty"`
	Labels               map[string]string `json:"labels,omitempty"`
	MasterCount          int               `json:"master_count,omitempty"`
	MasterFlavorID       string            `json:"master_flavor_id,omitempty"`
	Name                 string            `json:"name"`
	NetworkID            string            `json:"network_id" required:"true"`
	SubnetID             string            `json:"subnet_id" required:"true"`
	PodsNetworkCidr      string            `json:"pods_network_cidr,omitempty"`
	FloatingIPEnabled    bool              `json:"floating_ip_enabled"`
	APILBVIP             string            `json:"api_lb_vip,omitempty"`
	APILBFIP             string            `json:"api_lb_fip,omitempty"`
	IngressFloatingIP    string            `json:"ingress_floating_ip,omitempty"`
	RegistryAuthPassword string            `json:"registry_auth_password,omitempty"`
}

//ClusterActionsBaseOpts ...
type ClusterActionsBaseOpts struct {
	Action  string      `json:"action" required:"true"`
	Payload interface{} `json:"payload,omitempty"`
}

//ClusterUpgradeOpts ...
type ClusterUpgradeOpts struct {
	ClusterTemplateID string `json:"cluster_template_id" required:"true"`
	RollingEnabled    bool   `json:"rolling_enabled"`
}

//Cluster ...
type Cluster struct {
	APIAddress           string             `json:"api_address"`
	ClusterTemplateID    string             `json:"cluster_template_id"`
	CreateTimeout        int                `json:"create_timeout"`
	CreatedAt            time.Time          `json:"created_at"`
	DiscoveryURL         string             `json:"discovery_url"`
	KeyPair              string             `json:"keypair"`
	Labels               map[string]string  `json:"labels"`
	Links                []gophercloud.Link `json:"links"`
	MasterFlavorID       string             `json:"master_flavor_id"`
	MasterAddresses      []string           `json:"master_addresses"`
	MasterCount          int                `json:"master_count"`
	Name                 string             `json:"name"`
	NodeAddresses        []string           `json:"node_addresses"`
	NodeCount            int                `json:"node_count"`
	ProjectID            string             `json:"project_id"`
	StackID              string             `json:"stack_id"`
	Status               string             `json:"status"`
	NewStatus            string             `json:"new_status"`
	StatusReason         string             `json:"status_reason"`
	UUID                 string             `json:"uuid"`
	UpdatedAt            time.Time          `json:"updated_at"`
	UserID               string             `json:"user_id"`
	NetworkID            string             `json:"network_id"`
	SubnetID             string             `json:"subnet_id"`
	PodsNetworkCidr      string             `json:"pods_network_cidr"`
	FloatingIPEnabled    bool               `json:"floating_ip_enabled"`
	APILBVIP             string             `json:"api_lb_vip"`
	APILBFIP             string             `json:"api_lb_fip"`
	IngressFloatingIP    string             `json:"ingress_floating_ip"`
	RegistryAuthPassword string             `json:"registry_auth_password"`
}

//ClusterTemplate ...
type ClusterTemplate struct {
	clustertemplates.ClusterTemplate
	Version string `json:"version"`
}

//ClusterTemplates ...
type ClusterTemplates struct {
	ClusterTemplates []*ClusterTemplate `json:"clustertemplates"`
}

//OptsBuilder ...
type OptsBuilder interface {
	Map() (map[string]interface{}, error)
}

// PatchOptsBuilder ...
type PatchOptsBuilder interface {
	PatchMap() ([]map[string]interface{}, error)
}

// Map ...
func (opts *ClusterCreateOpts) Map() (map[string]interface{}, error) {
	cluster, err := gophercloud.BuildRequestBody(*opts, "")
	return cluster, err
}

// Map ...
func (opts *ClusterActionsBaseOpts) Map() (map[string]interface{}, error) {
	cluster, err := gophercloud.BuildRequestBody(*opts, "")
	return cluster, err
}

// Map ...
func (opts *NodeGroupCreateOpts) Map() (map[string]interface{}, error) {
	cluster, err := gophercloud.BuildRequestBody(*opts, "")
	return cluster, err
}

// Map ...
func (opts *NodeGroup) Map() (map[string]interface{}, error) {
	body, err := gophercloud.BuildRequestBody(*opts, "")
	return body, err
}

// Map ...
func (opts *NodeGroupScaleOpts) Map() (map[string]interface{}, error) {
	body, err := gophercloud.BuildRequestBody(*opts, "")
	return body, err
}

// Map ...
func (opts *ClusterUpgradeOpts) Map() (map[string]interface{}, error) {
	body, err := gophercloud.BuildRequestBody(*opts, "")
	return body, err
}

// Map ...
func (opts *NodeGroupBatchAddParams) Map() (map[string]interface{}, error) {
	batch, err := gophercloud.BuildRequestBody(*opts, "")
	return batch, err
}

// Map ...
func (opts *NodeGroupBatchDelParams) Map() (map[string]interface{}, error) {
	body, err := gophercloud.BuildRequestBody(*opts, "")
	return body, err
}

// Map ...
func (opts *NodeGroupPatchParams) Map() (map[string]interface{}, error) {
	body, err := gophercloud.BuildRequestBody(*opts, "")
	return body, err
}

// PatchMap ...
func (opts *NodeGroupClusterPatchOpts) PatchMap() ([]map[string]interface{}, error) {
	var lb []map[string]interface{}
	for _, opt := range *opts {
		b, err := opt.Map()
		if err != nil {
			return nil, err
		}
		lb = append(lb, b)
	}
	return lb, nil
}

const (
	clustersAPIPath        = "clusters"
	nodeGroupsAPIPath      = "nodegroups"
	clusterTemplateAPIPath = "clustertemplates"
)

type commonResult struct {
	gophercloud.Result
}

// KubeConfigResult ...
type KubeConfigResult struct {
	commonResult
}

// ClusterResult ...
type ClusterResult struct {
	commonResult
}

// NodeGroupsResult ...
type NodeGroupsResult struct {
	commonResult
}

// NodeGroupResult ...
type NodeGroupResult struct {
	commonResult
}

// NodeGroupBatchResult ...
type NodeGroupBatchResult struct {
	gophercloud.ErrResult
}

// NodeGroupDeleteResult ...
type NodeGroupDeleteResult struct {
	gophercloud.ErrResult
}

// ClusterDeleteResult ...
type ClusterDeleteResult struct {
	gophercloud.ErrResult
}

// NodeGroupScaleResult ...
type NodeGroupScaleResult struct {
	commonResult
}

// ClusterTemplateResult ...
type ClusterTemplateResult struct {
	commonResult
}

// ClusterTemplateListResult ...
type ClusterTemplateListResult struct {
	commonResult
}

// Extract ...
func (r NodeGroupScaleResult) Extract() (string, error) {
	var s struct {
		UUID string
	}
	err := r.ExtractInto(&s)
	return s.UUID, err
}

// Extract ...
func (r ClusterResult) Extract() (*Cluster, error) {
	var s *Cluster
	err := r.ExtractInto(&s)
	return s, err
}

// Extract ...
func (r KubeConfigResult) Extract() (*string, error) {
	var s *string
	err := r.ExtractInto(&s)
	return s, err
}

// Extract ...
func (r ClusterTemplateResult) Extract() (*ClusterTemplate, error) {
	var s *ClusterTemplate
	err := r.ExtractInto(&s)
	return s, err
}

// Extract ...
func (r ClusterTemplateListResult) Extract() (*ClusterTemplates, error) {
	var s *ClusterTemplates
	err := r.ExtractInto(&s)
	return s, err
}

// Extract ...
func (r NodeGroupsResult) Extract() (*NodeGroups, error) {
	var s *NodeGroups
	err := r.ExtractInto(&s)
	return s, err
}

// Extract ...
func (r NodeGroupResult) Extract() (*NodeGroup, error) {
	var s *NodeGroup
	err := r.ExtractInto(&s)
	return s, err
}

// ClusterTemplateGet ...
func ClusterTemplateGet(client ContainerClient, id string) (r ClusterTemplateResult) {
	var result *http.Response
	reqOpts := getRequestOpts(200)
	result, r.Err = client.Get(getURL(client, clusterTemplateAPIPath, id), &r.Body, reqOpts)
	if r.Err == nil {
		r.Header = result.Header
	}
	return
}

// CreateCluster ...
func CreateCluster(client ContainerClient, opts OptsBuilder) (r clusters.CreateResult) {
	log.Printf("CREATE cluster")
	b, err := opts.Map()
	if err != nil {
		r.Err = err
		return
	}
	var result *http.Response
	reqOpts := getRequestOpts(202)
	result, r.Err = client.Post(baseURL(client, clustersAPIPath), b, &r.Body, reqOpts)
	if r.Err == nil {
		r.Header = result.Header
	}
	return
}

// ClusterUpgrade ...
func ClusterUpgrade(client ContainerClient, id string, opts OptsBuilder) (r clusters.UpdateResult) {
	b, err := opts.Map()
	if err != nil {
		r.Err = err
		return
	}
	reqOpts := getRequestOpts(200, 202)
	var result *http.Response
	result, r.Err = client.Patch(upgradeURL(client, clustersAPIPath, id), b, &r.Body, reqOpts)
	if r.Err == nil {
		r.Header = result.Header
	}
	return
}

// ClusterUpdateMasters ...
func ClusterUpdateMasters(client ContainerClient, id string, opts OptsBuilder) (r clusters.UpdateResult) {
	log.Printf("UPDATE masters for cluster %s", id)
	b, err := opts.Map()
	if err != nil {
		r.Err = err
		return
	}
	reqOpts := getRequestOpts(200, 202)
	var result *http.Response
	result, r.Err = client.Post(actionsURL(client, clustersAPIPath, id), b, &r.Body, reqOpts)
	if r.Err == nil {
		r.Header = result.Header
	}
	return
}

// ClusterSwitchState ...
func ClusterSwitchState(client ContainerClient, id string, opts OptsBuilder) (r clusters.UpdateResult) {
	reqBody, err := opts.Map()
	if err != nil {
		r.Err = err
		return
	}
	reqOpts := getRequestOpts(202)
	var result *http.Response
	result, r.Err = client.Post(actionsURL(client, clustersAPIPath, id), reqBody, &r.Body, reqOpts)
	if r.Err == nil {
		r.Header = result.Header
	}
	return
}

// NodeGroupGet ...
func NodeGroupGet(client ContainerClient, id string) (r NodeGroupResult) {
	var result *http.Response
	reqOpts := getRequestOpts(200)
	result, r.Err = client.Get(getURL(client, nodeGroupsAPIPath, id), &r.Body, reqOpts)
	if r.Err == nil {
		r.Header = result.Header
	}
	return
}

// NodeGroupScale ...
func NodeGroupScale(client ContainerClient, id string, opts OptsBuilder) (r NodeGroupResult) {
	b, err := opts.Map()
	if err != nil {
		r.Err = err
		return
	}
	reqOpts := getRequestOpts(202)
	var result *http.Response
	result, r.Err = client.Patch(scaleURL(client, nodeGroupsAPIPath, id), b, &r.Body, reqOpts)
	if r.Err == nil {
		r.Header = result.Header
	}
	return
}

// NodeGroupCreate ...
func NodeGroupCreate(client ContainerClient, opts OptsBuilder) (r NodeGroupResult) {
	b, err := opts.Map()
	if err != nil {
		r.Err = err
		return
	}
	var result *http.Response
	reqOpts := getRequestOpts(202)
	result, r.Err = client.Post(baseURL(client, nodeGroupsAPIPath), b, &r.Body, reqOpts)
	if r.Err == nil {
		r.Header = result.Header
	}
	return
}

// ClusterGet ...
func ClusterGet(client ContainerClient, id string) (r ClusterResult) {
	log.Printf("GET cluster %s", id)
	var result *http.Response
	reqOpts := getRequestOpts(200)
	result, r.Err = client.Get(getURL(client, clustersAPIPath, id), &r.Body, reqOpts)
	if r.Err == nil {
		r.Header = result.Header
	}
	return
}

// K8sConfigGet ...
func K8sConfigGet(client ContainerClient, id string) (string, error) {
	var result *http.Response
	reqOpts := getRequestOpts(200)
	result, err := client.Get(kubeConfigURL(client, clustersAPIPath, id), nil, reqOpts)
	if err != nil {
		return "", err
	}
	buf := bytes.NewBuffer(make([]byte, 0, result.ContentLength))
	_, err = io.Copy(buf, result.Body)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

// NodeGroupPatch ...
func NodeGroupPatch(client ContainerClient, id string, opts PatchOptsBuilder) (r NodeGroupScaleResult) {
	b, err := opts.PatchMap()
	if err != nil {
		r.Err = err
	}
	var result *http.Response
	reqOpts := getRequestOpts(200)
	result, r.Err = client.Patch(getURL(client, nodeGroupsAPIPath, id), b, &r.Body, reqOpts)
	if r.Err == nil {
		r.Header = result.Header
	}
	return
}

// NodeGroupDelete ...
func NodeGroupDelete(client ContainerClient, id string) (r NodeGroupDeleteResult) {
	var result *http.Response
	reqOpts := getRequestOpts(204)
	result, r.Err = client.Delete(getURL(client, nodeGroupsAPIPath, id), reqOpts)
	if r.Err == nil {
		r.Header = result.Header
	}
	return
}

// ClusterDelete ...
func ClusterDelete(client ContainerClient, id string) (r ClusterDeleteResult) {
	log.Printf("DELETE clsuter %s", id)
	var result *http.Response
	reqOpts := getRequestOpts()
	result, r.Err = client.Delete(deleteURL(client, clustersAPIPath, id), reqOpts)
	r.Header = result.Header
	return
}

func getRequestOpts(codes ...int) *gophercloud.RequestOpts {
	reqOpts := &gophercloud.RequestOpts{
		OkCodes: codes,
	}
	if len(codes) != 0 {
		reqOpts.OkCodes = codes
	}
	addMicroVersionHeader(reqOpts)
	return reqOpts
}
