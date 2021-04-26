package netbox

import (
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	netboxclient "github.com/netbox-community/go-netbox/netbox/client"
	"github.com/netbox-community/go-netbox/netbox/client/virtualization"
	"github.com/netbox-community/go-netbox/netbox/models"
)

func resourceNetboxVirtualizationCluster() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetboxVirtualizationClusterCreate,
		Read:   resourceNetboxVirtualizationClusterRead,
		Update: resourceNetboxVirtualizationClusterUpdate,
		Delete: resourceNetboxVirtualizationClusterDelete,
		Exists: resourceNetboxVirtualizationClusterExists,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 50),
			},
			"type_id": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"site_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"tenant_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
		},
	}
}

func resourceNetboxVirtualizationClusterCreate(d *schema.ResourceData,
	m interface{}) error {
	client := m.(*netboxclient.NetBoxAPI)

	clusterName := d.Get("name").(string)
	clusterTypeID := int64(d.Get("type_id").(int))
	clusterSiteID := int64(d.Get("site_id").(int))
	clusterTenantID := int64(d.Get("tenant_id").(int))

	newResource := &models.WritableCluster{
		Name: &clusterName,
		Type: &clusterTypeID,
	}

	if clusterTenantID != 0 {
		newResource.Tenant = &clusterTenantID
	}

	if clusterSiteID != 0 {
		newResource.Site = &clusterSiteID
	}

	resource := virtualization.NewVirtualizationClustersCreateParams().WithData(newResource)

	resourceCreated, err := client.Virtualization.VirtualizationClustersCreate(resource, nil)
	if err != nil {
		return err
	}

	d.SetId(strconv.FormatInt(resourceCreated.Payload.ID, 10))
	return resourceNetboxVirtualizationClusterRead(d, m)
}

func resourceNetboxVirtualizationClusterRead(d *schema.ResourceData,
	m interface{}) error {
	client := m.(*netboxclient.NetBoxAPI)

	resourceID := d.Id()
	params := virtualization.NewVirtualizationClustersListParams().WithID(&resourceID)
	resources, err := client.Virtualization.VirtualizationClustersList(params, nil)
	if err != nil {
		return err
	}

	for _, resource := range resources.Payload.Results {
		if strconv.FormatInt(resource.ID, 10) == d.Id() {
			if err = d.Set("name", resource.Name); err != nil {
				return err
			}

			if resource.Type == nil {
				if err = d.Set("type_id", 0); err != nil {
					return err
				}
			} else {
				if err = d.Set("type_id", resource.Type.ID); err != nil {
					return err
				}
			}

			if resource.Site == nil {
				if err = d.Set("site_id", 0); err != nil {
					return err
				}
			} else {
				if err = d.Set("site_id", resource.Site.ID); err != nil {
					return err
				}
			}

			if resource.Tenant == nil {
				if err = d.Set("tenant_id", 0); err != nil {
					return err
				}
			} else {
				if err = d.Set("tenant_id", resource.Tenant.ID); err != nil {
					return err
				}
			}

			return nil
		}
	}

	d.SetId("")
	return nil
}

func resourceNetboxVirtualizationClusterUpdate(d *schema.ResourceData,
	m interface{}) error {
	client := m.(*netboxclient.NetBoxAPI)
	params := &models.WritableCluster{}

	// Required parameters
	name := d.Get("name").(string)
	params.Name = &name

	typeID := int64(d.Get("type_id").(int))
	params.Type = &typeID

	//Optional parameters
	if d.HasChange("site_id") {
		siteID := int64(d.Get("site_id").(int))
		params.Site = &siteID
	}

	if d.HasChange("tenant_id") {
		tenantID := int64(d.Get("tenant_id").(int))
		params.Tenant = &tenantID
	}

	resource := virtualization.NewVirtualizationClustersPartialUpdateParams().WithData(
		params)

	resourceID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return fmt.Errorf("Unable to convert ID into int64")
	}

	resource.SetID(resourceID)

	_, err = client.Virtualization.VirtualizationClustersPartialUpdate(resource, nil)
	if err != nil {
		return err
	}

	return resourceNetboxVirtualizationClusterRead(d, m)
}

func resourceNetboxVirtualizationClusterDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*netboxclient.NetBoxAPI)

	resourceExists, err := resourceNetboxVirtualizationClusterExists(d, m)
	if err != nil {
		return err
	}

	if !resourceExists {
		return nil
	}

	id, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return fmt.Errorf("Unable to convert ID into int64")
	}

	resource := virtualization.NewVirtualizationClustersDeleteParams().WithID(id)
	if _, err := client.Virtualization.VirtualizationClustersDelete(resource, nil); err != nil {
		return err
	}

	return nil
}

func resourceNetboxVirtualizationClusterExists(d *schema.ResourceData, m interface{}) (b bool,
	e error) {
	client := m.(*netboxclient.NetBoxAPI)
	resourceExist := false

	resourceID := d.Id()
	params := virtualization.NewVirtualizationClustersListParams().WithID(&resourceID)
	resources, err := client.Virtualization.VirtualizationClustersList(params, nil)
	if err != nil {
		return resourceExist, err
	}

	for _, resource := range resources.Payload.Results {
		if strconv.FormatInt(resource.ID, 10) == d.Id() {
			resourceExist = true
		}
	}

	return resourceExist, nil
}
