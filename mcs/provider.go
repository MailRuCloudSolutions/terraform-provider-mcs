package mcs

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/utils/terraform/auth"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/meta"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

const (
	maxRetriesCount         = 3
	defaultIdentityEndpoint = "https://infra.mail.ru/identity/v3/"
	defaultUsersDomainName  = "users"
	requestsMaxRetriesCount = 3
	requestsRetryDelay      = 30 * time.Millisecond
)

// configer is interface to work with gophercloud.Config calls
type configer interface {
	LoadAndValidate() error
	IdentityV3Client(region string) (ContainerClient, error)
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

// IdentityV3Client is implementation of ContainerInfraV1Client method
func (c *config) IdentityV3Client(region string) (ContainerClient, error) {
	return c.Config.IdentityV3Client(region)
}

// ContainerInfraV1Client is implementation of ContainerInfraV1Client method
func (c *config) ContainerInfraV1Client(region string) (ContainerClient, error) {
	return c.Config.ContainerInfraV1Client(region)
}

// DatabaseV1Client is implementation of DatabaseV1Client method
func (c *config) DatabaseV1Client(region string) (ContainerClient, error) {
	client, clientErr := c.Config.DatabaseV1Client(region)
	client.ProviderClient.RetryFunc = func(context context.Context, method, url string, options *gophercloud.RequestOpts, err error, failCount uint) error {
		if failCount >= requestsMaxRetriesCount {
			return err
		}
		switch errType := err.(type) {
		case gophercloud.ErrDefault500, gophercloud.ErrDefault503:
			time.Sleep(requestsRetryDelay)
			return nil
		case gophercloud.ErrUnexpectedResponseCode:
			if errType.Actual == http.StatusGatewayTimeout {
				time.Sleep(requestsRetryDelay)
				return nil
			}
			return err
		default:
			return err
		}
	}
	return client, clientErr
}

func newConfig(d *schema.ResourceData, terraformVersion string) (configer, error) {
	if os.Getenv("TF_ACC") == "" {
		return &dummyConfig{}, nil
	}

	config := &config{
		auth.Config{
			CACertFile:       d.Get("cacert_file").(string),
			ClientCertFile:   d.Get("cert").(string),
			ClientKeyFile:    d.Get("key").(string),
			Password:         d.Get("password").(string),
			TenantID:         d.Get("project_id").(string),
			Region:           d.Get("region").(string),
			AllowReauth:      true,
			MaxRetries:       maxRetriesCount,
			TerraformVersion: terraformVersion,
			SDKVersion:       meta.SDKVersionString(),
		},
	}

	if config.TenantID == "" {
		config.TenantID = os.Getenv("OS_PROJECT_ID")
	}
	if config.UserDomainID == "" {
		config.UserDomainID = os.Getenv("OS_USER_DOMAIN_ID")
	}
	if config.Password == "" {
		config.Password = os.Getenv("OS_PASSWORD")
	}
	if config.Username == "" {
		config.Username = os.Getenv("OS_USERNAME")
	}
	if config.Region == "" {
		config.Region = os.Getenv("OS_REGION")
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

	config.Username = os.Getenv("OS_USERNAME")
	if v, ok := d.GetOk("username"); ok {
		config.Username = v.(string)
	}
	if config.Username == "" {
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
				Description: "The Identity authentication URL.",
			},
			"project_id": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("PROJECT_ID", ""),
				Description: "The ID of Project to login with.",
			},
			"password": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("PASSWORD", ""),
				Description: "Password to login with.",
			},
			"username": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("USER_NAME", ""),
				Description: "User name to login with.",
			},
			"region": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("REGION", "RegionOne"),
				Description: "A region to use.",
			},
			"insecure": {
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("INSECURE", nil),
				Description: "Trust self-signed certificates.",
			},
			"cacert_file": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("CACERT", ""),
				Description: "A Custom CA certificate.",
			},
			"cert": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("CERT", ""),
				Description: "A client certificate to authenticate with.",
			},
			"key": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("KEY", ""),
				Description: "A client private key to authenticate with.",
			},
		},

		DataSourcesMap: map[string]*schema.Resource{
			"mcs_kubernetes_clustertemplate":  dataSourceKubernetesClusterTemplate(),
			"mcs_kubernetes_clustertemplates": dataSourceKubernetesClusterTemplates(),
			"mcs_kubernetes_cluster":          dataSourceKubernetesCluster(),
			"mcs_kubernetes_node_group":       dataSourceKubernetesNodeGroup(),
			"mcs_db_instance":                 dataSourceDatabaseInstance(),
			"mcs_db_user":                     dataSourceDatabaseUser(),
			"mcs_db_database":                 dataSourceDatabaseDatabase(),
			"mcs_region":                      dataSourceMcsRegion(),
			"mcs_regions":                     dataSourceMcsRegions(),
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
