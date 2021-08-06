package mcs

import (
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"gitlab.corp.mail.ru/infra/paas/terraform-provider-mcs/mcs/internal/valid"
)

// OperationCreate ...
const (
	operationCreate          = 60
	operationUpdate          = 60
	operationDelete          = 30
	createUpdateDelay        = 1
	createUpdatePollInterval = 20
	deleteDelay              = 30
	nodeGroupDeleteDelay     = 10
	deletePollInterval       = 10
)

type clusterStatus string

var (
	clusterStatusDeleting     clusterStatus = "DELETING"
	clusterStatusDeleted      clusterStatus = "DELETED"
	clusterStatusReconciling  clusterStatus = "RECONCILING"
	clusterStatusProvisioning clusterStatus = "PROVISIONING"
	clusterStatusRunning      clusterStatus = "RUNNING"
	clusterStatusError        clusterStatus = "ERROR"
	clusterStatusShutoff      clusterStatus = "SHUTOFF"
)

var stateStatusMap = map[clusterStatus]string{
	clusterStatusRunning: "turn_on_cluster",
	clusterStatusShutoff: "turn_off_cluster",
}

func resourceKubernetesCluster() *schema.Resource {
	return &schema.Resource{
		Create: resourceKubernetesClusterCreate,
		Read:   resourceKubernetesClusterRead,
		Update: resourceKubernetesClusterUpdate,
		Delete: resourceKubernetesClusterDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(operationCreate * time.Minute),
			Update: schema.DefaultTimeout(operationUpdate * time.Minute),
			Delete: schema.DefaultTimeout(operationDelete * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"region": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
					name := val.(string)
					if err := valid.ClusterName(name); err != nil {
						errs = append(errs, err)
					}
					return
				},
			},
			"project_id": {
				Type:     schema.TypeString,
				ForceNew: true,
				Computed: true,
			},
			"user_id": {
				Type:     schema.TypeString,
				ForceNew: true,
				Computed: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				ForceNew: false,
				Computed: true,
			},
			"updated_at": {
				Type:     schema.TypeString,
				ForceNew: false,
				Computed: true,
			},
			"api_address": {
				Type:     schema.TypeString,
				ForceNew: true,
				Computed: true,
			},
			"cluster_template_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: false,
			},
			"master_flavor": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
				Computed: true,
			},
			"keypair": {
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
			},
			"labels": {
				Type:     schema.TypeMap,
				Optional: true,
				Computed: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"master_count": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
			"master_addresses": {
				Type:     schema.TypeString,
				ForceNew: true,
				Computed: true,
			},
			"node_addresses": {
				Type:     schema.TypeString,
				ForceNew: true,
				Computed: true,
			},
			"stack_id": {
				Type:     schema.TypeString,
				ForceNew: true,
				Computed: true,
			},
			"network_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"subnet_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"status": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
				Computed: true,
			},
			"pods_network_cidr": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"floating_ip_enabled": {
				Type:     schema.TypeBool,
				Required: true,
				ForceNew: true,
			},
			"api_lb_vip": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"api_lb_fip": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"ingress_floating_ip": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"registry_auth_password": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"availability_zone": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
					zone := val.(string)
					if err := valid.AvailabilityZone(zone); err != nil {
						errs = append(errs, err)
					}
					return
				},
			},
		},
	}
}

func resourceKubernetesClusterCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(configer)
	containerInfraClient, err := config.ContainerInfraV1Client(getRegion(d, config))
	if err != nil {
		return fmt.Errorf("error creating container infra client: %s", err)
	}

	// Get and check labels map.
	rawLabels := d.Get("labels").(map[string]interface{})
	labels, err := extractKubernetesLabelsMap(rawLabels)
	if err != nil {
		return err
	}

	createOpts := clusterCreateOpts{
		ClusterTemplateID:    d.Get("cluster_template_id").(string),
		MasterFlavorID:       d.Get("master_flavor").(string),
		Keypair:              d.Get("keypair").(string),
		Labels:               labels,
		Name:                 d.Get("name").(string),
		NetworkID:            d.Get("network_id").(string),
		SubnetID:             d.Get("subnet_id").(string),
		PodsNetworkCidr:      d.Get("pods_network_cidr").(string),
		FloatingIPEnabled:    d.Get("floating_ip_enabled").(bool),
		APILBVIP:             d.Get("api_lb_vip").(string),
		APILBFIP:             d.Get("api_lb_fip").(string),
		IngressFloatingIP:    d.Get("ingress_floating_ip").(string),
		RegistryAuthPassword: d.Get("registry_auth_password").(string),
		AvailabilityZone:     d.Get("availability_zone").(string),
	}

	if masterCount, ok := d.GetOk("master_count"); ok {
		mCount := masterCount.(int)
		if mCount < 1 {
			return fmt.Errorf("master_count if set must be greater or equal 1: %s", err)
		}
		createOpts.MasterCount = mCount
	}

	s, err := createCluster(containerInfraClient, &createOpts).Extract()
	if err != nil {
		return fmt.Errorf("error creating mcs_kubernetes_cluster: %s", err)
	}

	// Store the cluster ID.
	d.SetId(s)

	stateConf := &resource.StateChangeConf{
		Pending:      []string{string(clusterStatusProvisioning)},
		Target:       []string{string(clusterStatusRunning)},
		Refresh:      kubernetesStateRefreshFunc(containerInfraClient, s),
		Timeout:      d.Timeout(schema.TimeoutCreate),
		Delay:        createUpdateDelay * time.Minute,
		PollInterval: createUpdatePollInterval * time.Second,
	}
	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf(
			"error waiting for mcs_kubernetes_cluster %s to become ready: %s", s, err)
	}

	log.Printf("[DEBUG] Created mcs_kubernetes_cluster %s", s)
	return resourceKubernetesClusterRead(d, meta)
}

func resourceKubernetesClusterRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(configer)
	containerInfraClient, err := config.ContainerInfraV1Client(getRegion(d, config))
	if err != nil {
		return fmt.Errorf("error creating container infra client: %s", err)
	}

	cluster, err := clusterGet(containerInfraClient, d.Id()).Extract()
	if err != nil {
		return checkDeleted(d, err, "error retrieving mcs_kubernetes_cluster")
	}

	log.Printf("[DEBUG] retrieved mcs_kubernetes_cluster %s: %#v", d.Id(), cluster)

	// Get and check labels map.
	rawLabels := d.Get("labels").(map[string]interface{})
	labels, err := extractKubernetesLabelsMap(rawLabels)
	if err != nil {
		return err
	}

	if err := d.Set("labels", labels); err != nil {
		return fmt.Errorf("unable to set mcs_kubernetes_cluster labels: %s", err)
	}

	d.Set("name", cluster.Name)
	d.Set("api_address", cluster.APIAddress)
	d.Set("cluster_template_id", cluster.ClusterTemplateID)
	d.Set("create_timeout", cluster.CreateTimeout)
	d.Set("discovery_url", cluster.DiscoveryURL)
	d.Set("master_flavor", cluster.MasterFlavorID)
	d.Set("keypair", cluster.KeyPair)
	d.Set("master_count", cluster.MasterCount)
	d.Set("master_addresses", cluster.MasterAddresses)
	d.Set("node_addresses", cluster.NodeAddresses)
	d.Set("stack_id", cluster.StackID)
	d.Set("status", cluster.NewStatus)
	d.Set("pods_network_cidr", cluster.PodsNetworkCidr)
	d.Set("floating_ip_enabled", cluster.FloatingIPEnabled)
	d.Set("api_lb_vip", cluster.APILBVIP)
	d.Set("api_lb_fip", cluster.APILBFIP)
	d.Set("ingress_floating_ip", cluster.IngressFloatingIP)
	d.Set("registry_auth_password", cluster.RegistryAuthPassword)
	d.Set("availability_zone", cluster.AvailabilityZone)

	// Allow to read old api clusters
	if cluster.NetworkID != "" {
		d.Set("network_id", cluster.NetworkID)
	} else {
		d.Set("network_id", cluster.Labels["fixed_network"])
	}
	if cluster.SubnetID != "" {
		d.Set("subnet_id", cluster.SubnetID)
	} else {
		d.Set("subnet_id", cluster.Labels["fixed_subnet"])
	}

	if err := d.Set("created_at", getTimestamp(&cluster.CreatedAt)); err != nil {
		log.Printf("[DEBUG] Unable to set mcs_kubernetes_cluster created_at: %s", err)
	}
	if err := d.Set("updated_at", getTimestamp(&cluster.UpdatedAt)); err != nil {
		log.Printf("[DEBUG] Unable to set mcs_kubernetes_cluster updated_at: %s", err)
	}

	return nil
}

func resourceKubernetesClusterUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(configer)
	containerInfraClient, err := config.ContainerInfraV1Client(getRegion(d, config))
	if err != nil {
		return fmt.Errorf("error creating container infra client: %s", err)
	}

	stateConf := &resource.StateChangeConf{
		Refresh:      kubernetesStateRefreshFunc(containerInfraClient, d.Id()),
		Timeout:      d.Timeout(schema.TimeoutUpdate),
		Delay:        createUpdateDelay * time.Minute,
		PollInterval: createUpdatePollInterval * time.Second,
		Pending:      []string{string(clusterStatusReconciling)},
		Target:       []string{string(clusterStatusRunning)},
	}

	cluster, err := clusterGet(containerInfraClient, d.Id()).Extract()
	if err != nil {
		return fmt.Errorf("error retrieving cluster: %s", err)
	}

	switch cluster.NewStatus {
	case clusterStatusShutoff:
		changed, err := checkForStatus(d, containerInfraClient, cluster)
		if err != nil {
			return err
		}
		if changed {
			err := checkForClusterTemplateID(d, containerInfraClient, stateConf)
			if err != nil {
				return err
			}
			err = checkForMasterFlavor(d, containerInfraClient, stateConf)
			if err != nil {
				return err
			}
		} else {
			return fmt.Errorf("changing cluster attributes is prohibited when cluster has SHUTOFF status")
		}
	case clusterStatusRunning:
		err := checkForClusterTemplateID(d, containerInfraClient, stateConf)
		if err != nil {
			return err
		}
		err = checkForMasterFlavor(d, containerInfraClient, stateConf)
		if err != nil {
			return err
		}
		_, err = checkForStatus(d, containerInfraClient, cluster)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("changes in cluster are prohibited when status is not RUNNING/SHUTOFF; current status: %s", cluster.NewStatus)
	}

	return resourceKubernetesClusterRead(d, meta)
}

func checkForClusterTemplateID(d *schema.ResourceData, containerInfraClient ContainerClient, stateConf *resource.StateChangeConf) error {
	if d.HasChange("cluster_template_id") {
		upgradeOpts := clusterUpgradeOpts{
			ClusterTemplateID: d.Get("cluster_template_id").(string),
			RollingEnabled:    true,
		}

		_, err := clusterUpgrade(containerInfraClient, d.Id(), &upgradeOpts).Extract()
		if err != nil {
			return fmt.Errorf("error upgrade cluster : %s", err)
		}

		_, err = stateConf.WaitForState()
		if err != nil {
			return fmt.Errorf(
				"error waiting for mcs_kubernetes_cluster %s to become upgraded: %s", d.Id(), err)
		}
	}
	return nil
}

func checkForMasterFlavor(d *schema.ResourceData, containerInfraClient ContainerClient, stateConf *resource.StateChangeConf) error {
	if d.HasChange("master_flavor") {
		upgradeOpts := clusterActionsBaseOpts{
			Action: "resize_masters",
			Payload: map[string]string{
				"flavor": d.Get("master_flavor").(string),
			},
		}

		_, err := clusterUpdateMasters(containerInfraClient, d.Id(), &upgradeOpts).Extract()
		if err != nil {
			return fmt.Errorf("error updating cluster's falvor : %s", err)
		}

		_, err = stateConf.WaitForState()
		if err != nil {
			return fmt.Errorf(
				"error waiting for mcs_kubernetes_cluster %s to become updated: %s", d.Id(), err)
		}
	}
	return nil
}

func checkForStatus(d *schema.ResourceData, containerInfraClient ContainerClient, cluster *cluster) (bool, error) {

	turnOffConf := &resource.StateChangeConf{
		Refresh:      kubernetesStateRefreshFunc(containerInfraClient, d.Id()),
		Timeout:      d.Timeout(schema.TimeoutUpdate),
		Delay:        createUpdateDelay * time.Minute,
		PollInterval: createUpdatePollInterval * time.Second,
		Pending:      []string{string(clusterStatusRunning)},
		Target:       []string{string(clusterStatusShutoff)},
	}

	turnOnConf := &resource.StateChangeConf{
		Refresh:      kubernetesStateRefreshFunc(containerInfraClient, d.Id()),
		Timeout:      d.Timeout(schema.TimeoutUpdate),
		Delay:        createUpdateDelay * time.Minute,
		PollInterval: createUpdatePollInterval * time.Second,
		Pending:      []string{string(clusterStatusShutoff)},
		Target:       []string{string(clusterStatusRunning)},
	}

	if d.HasChange("status") {
		currentStatus := d.Get("status").(clusterStatus)
		if cluster.NewStatus != clusterStatusRunning && cluster.NewStatus != clusterStatusShutoff {
			return false, fmt.Errorf("turning on/off is prohibited due to cluster's status %s", cluster.NewStatus)
		}
		switchStateOpts := clusterActionsBaseOpts{
			Action: stateStatusMap[currentStatus],
		}
		_, err := clusterSwitchState(containerInfraClient, d.Id(), &switchStateOpts).Extract()
		if err != nil {
			return false, fmt.Errorf("error during switching state: %s", err)
		}

		var switchStateConf *resource.StateChangeConf
		switch currentStatus {
		case clusterStatusRunning:
			switchStateConf = turnOnConf
		case clusterStatusShutoff:
			switchStateConf = turnOffConf
		default:
			return false, fmt.Errorf("unknown status provided: %s", currentStatus)
		}

		_, err = switchStateConf.WaitForState()
		if err != nil {
			return false, fmt.Errorf(
				"error waiting for mcs_kubernetes_cluster %s to become updated: %s", d.Id(), err)
		}
		return true, nil

	}
	return false, nil
}

func resourceKubernetesClusterDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(configer)
	containerInfraClient, err := config.ContainerInfraV1Client(getRegion(d, config))
	if err != nil {
		return fmt.Errorf("error creating container infra client: %s", err)
	}

	if err := clusterDelete(containerInfraClient, d.Id()).ExtractErr(); err != nil {
		return checkDeleted(d, err, "error deleting mcs_kubernetes_cluster")
	}

	stateConf := &resource.StateChangeConf{
		Pending:      []string{string(clusterStatusDeleting)},
		Target:       []string{string(clusterStatusDeleted)},
		Refresh:      kubernetesStateRefreshFunc(containerInfraClient, d.Id()),
		Timeout:      d.Timeout(schema.TimeoutDelete),
		Delay:        deleteDelay * time.Second,
		PollInterval: deletePollInterval * time.Second,
	}
	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf(
			"error waiting for mcs_kubernetes_cluster %s to become deleted: %s", d.Id(), err)
	}

	return nil
}
