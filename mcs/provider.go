package mcs

import (
	"fmt"
	"os"

	"github.com/gophercloud/utils/terraform/auth"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/meta"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

const (
	maxRetriesCount         = 3
	defaultIdentityEndpoint = "https://infra.mail.ru/identity/v3/"
	defaultUsersDomainName  = "users"
)

// configer is interface to work with gophercloud.Config calls
type configer interface {
	LoadAndValidate() error
	ContainerInfraV1Client(region string) (ContainerClient, error)
	DatabaseV1Client(region string) (ContainerClient, error)
	GetRegion() string
}

// config uses openstackbase.Config as the base/foundation of this provider's
type config struct {
	auth.Config
}

var _ configer = &config{}

// GetRegion is implementation of getRegion method
func (c *config) GetRegion() string {
	return c.Region
}

// ContainerInfraV1Client is implementation of ContainerInfraV1Client method
func (c *config) ContainerInfraV1Client(region string) (ContainerClient, error) {
	return c.Config.ContainerInfraV1Client(region)
}

// DatabaseV1Client is implementation of DatabaseV1Client method
func (c *config) DatabaseV1Client(region string) (ContainerClient, error) {
	return c.Config.DatabaseV1Client(region)
}

func newConfig(d *schema.ResourceData, terraformVersion string) (configer, error) {
	if os.Getenv("TF_ACC") != "" && os.Getenv("TF_ACC_DBAAS") == "" {
		return &dummyConfig{}, nil
	}
	config := &config{
		auth.Config{
			CACertFile:       d.Get("cacert_file").(string),
			ClientCertFile:   d.Get("cert").(string),
			ClientKeyFile:    d.Get("key").(string),
			Password:         d.Get("password").(string),
			TenantID:         d.Get("project_id").(string),
			AllowReauth:      true,
			MaxRetries:       maxRetriesCount,
			TerraformVersion: terraformVersion,
			SDKVersion:       meta.SDKVersionString(),
		},
	}

	v, ok := d.GetOk("insecure")
	if ok {
		insecure := v.(bool)
		config.Insecure = &insecure
	}
	v, ok = d.GetOk("auth_url")
	if ok {
		config.IdentityEndpoint = v.(string)
	} else {
		config.IdentityEndpoint = defaultIdentityEndpoint
	}
	if err := initWithUsername(d, config); err != nil {
		return nil, err
	}

	if err := config.LoadAndValidate(); err != nil {
		return nil, err
	}
	return config, nil
}

func initWithUsername(d *schema.ResourceData, config *config) error {
	config.UserDomainName = defaultUsersDomainName

	v, ok := d.GetOk("username")
	if ok {
		config.Username = v.(string)
	} else {
		return fmt.Errorf("username must be specified")
	}
	return nil
}

// Provider returns a schema.Provider for MCS.
func Provider() terraform.ResourceProvider {
	provider := &schema.Provider{
		Schema: map[string]*schema.Schema{
			"auth_url": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("AUTH_URL", ""),
				Description: descriptions["auth_url"],
			},

			"project_id": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("PROJECT_ID", ""),
				Description: descriptions["project_id"],
			},

			"password": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("PASSWORD", ""),
				Description: descriptions["password"],
			},

			"username": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("USER_NAME", ""),
				Description: descriptions["username"],
			},

			"insecure": {
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("INSECURE", nil),
				Description: descriptions["insecure"],
			},

			"cacert_file": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("CACERT", ""),
				Description: descriptions["cacert_file"],
			},

			"cert": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("CERT", ""),
				Description: descriptions["cert"],
			},

			"key": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("KEY", ""),
				Description: descriptions["key"],
			},
		},

		DataSourcesMap: map[string]*schema.Resource{
			"mcs_kubernetes_clustertemplate": dataSourceKubernetesClusterTemplate(),
			"mcs_kubernetes_cluster":         dataSourceKubernetesCluster(),
			"mcs_kubernetes_node_group":      dataSourceKubernetesNodeGroup(),
			"mcs_db_instance":                dataSourceDatabaseInstance(),
			"mcs_db_user":                    dataSourceDatabaseUser(),
			"mcs_db_database":                dataSourceDatabaseDatabase(),
		},

		ResourcesMap: map[string]*schema.Resource{
			"mcs_kubernetes_cluster":     resourceKubernetesCluster(),
			"mcs_kubernetes_node_group":  resourceKubernetesNodeGroup(),
			"mcs_db_instance":            resourceDatabaseInstance(),
			"mcs_db_user":                resourceDatabaseUser(),
			"mcs_db_database":            resourceDatabaseDatabase(),
			"mcs_db_cluster":             resourceDatabaseCluster(),
			"mcs_db_cluster_with_shards": resourceDatabaseClusterWithShards(),
		},
	}

	provider.ConfigureFunc = func(d *schema.ResourceData) (interface{}, error) {
		terraformVersion := provider.TerraformVersion
		if terraformVersion == "" {
			// Terraform 0.12 introduced this field to the protocol
			// We can therefore assume that if it's missing it's 0.10 or 0.11
			terraformVersion = "0.11+compatible"
		}
		return newConfig(d, terraformVersion)
	}

	return provider
}

var descriptions map[string]string

func init() {
	descriptions = map[string]string{
		"auth_url": "The Identity authentication URL.",

		"project_id": "The ID of Project to login with.",

		"username": "User name to login with.",

		"password": "Password to login with.",

		"insecure": "Trust self-signed certificates.",

		"cacert_file": "A Custom CA certificate.",

		"cert": "A client certificate to authenticate with.",

		"key": "A client private key to authenticate with.",
	}
}
