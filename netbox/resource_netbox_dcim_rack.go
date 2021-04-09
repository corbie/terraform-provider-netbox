package netbox

import (
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	netboxclient "github.com/netbox-community/go-netbox/netbox/client"
	"github.com/netbox-community/go-netbox/netbox/client/dcim"
	"github.com/netbox-community/go-netbox/netbox/models"
)

func resourceNetboxDcimRack() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetboxDcimRackCreate,
		Read:   resourceNetboxDcimRackRead,
		Update: resourceNetboxDcimRackUpdate,
		Delete: resourceNetboxDcimRackDelete,
		Exists: resourceNetboxDcimRackExists,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"asset_tag": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 50),
			},
			"desc_units": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"facility_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 50),
			},
			"group_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 50),
			},
			"role_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"serial": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 50),
			},
			"site_id": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"status": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "active",
				ValidateFunc: validation.StringInSlice([]string{"reserved", "available", "planned", "active", "deprecated"}, false),
			},
			"tenant_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"u_height": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(1, 100),
			},
		},
	}
}

func resourceNetboxDcimRackCreate(d *schema.ResourceData,
	m interface{}) error {
	client := m.(*netboxclient.NetBoxAPI)

	rackName := d.Get("name").(string)
	rackSiteID := int64(d.Get("site_id").(int))
	rackStatus := d.Get("status").(string)
	rackTenantID := int64(d.Get("tenant_id").(int))
	rackAssetTag := d.Get("asset_tag").(string)
	rackFacilityID := d.Get("facility_id").(string)
	rackGroupID := int64(d.Get("group_id").(int))
	rackRoleID := int64(d.Get("role_id").(int))
	rackSerial := d.Get("serial").(string)
	rackDescUnits := d.Get("desc_units").(bool)
	rackUHeight := int64(d.Get("u_height").(int))

	newResource := &models.WritableRack{
		Name: &rackName,
		Site: &rackSiteID,
	}

	if rackAssetTag != "" {
		newResource.AssetTag = &rackAssetTag
	}

	if rackDescUnits != false {
		newResource.DescUnits = rackDescUnits
	}

	if rackFacilityID != "" {
		newResource.FacilityID = &rackFacilityID
	}

	if rackGroupID != 0 {
		newResource.Group = &rackGroupID
	}

	if rackRoleID != 0 {
		newResource.Role = &rackRoleID
	}

	if rackSerial != "" {
		newResource.Serial = rackSerial
	}

	if rackStatus != "" {
		newResource.Status = rackStatus
	}

	if rackTenantID != 0 {
		newResource.Tenant = &rackTenantID
	}

	if rackUHeight != 0 {
		newResource.UHeight = rackUHeight
	}

	resource := dcim.NewDcimRacksCreateParams().WithData(newResource)

	resourceCreated, err := client.Dcim.DcimRacksCreate(resource, nil)
	if err != nil {
		return err
	}

	d.SetId(strconv.FormatInt(resourceCreated.Payload.ID, 10))
	return resourceNetboxDcimRackRead(d, m)
}

func resourceNetboxDcimRackRead(d *schema.ResourceData,
	m interface{}) error {
	client := m.(*netboxclient.NetBoxAPI)

	resourceID := d.Id()
	params := dcim.NewDcimRacksListParams().WithID(&resourceID)
	resources, err := client.Dcim.DcimRacksList(params, nil)
	if err != nil {
		return err
	}

	for _, resource := range resources.Payload.Results {
		if strconv.FormatInt(resource.ID, 10) == d.Id() {
			if err = d.Set("name", resource.Name); err != nil {
				return err
			}

			if err = d.Set("asset_tag", resource.AssetTag); err != nil {
				return err
			}

			if err = d.Set("facility_id", resource.FacilityID); err != nil {
				return err
			}

			if err = d.Set("serial", resource.Serial); err != nil {
				return err
			}

			if err = d.Set("u_height", resource.UHeight); err != nil {
				return err
			}

			if err = d.Set("desc_units", resource.DescUnits); err != nil {
				return err
			}

			if resource.Group == nil {
				if err = d.Set("group_id", 0); err != nil {
					return err
				}
			} else {
				if err = d.Set("group_id", resource.Group.ID); err != nil {
					return err
				}
			}

			if resource.Role == nil {
				if err = d.Set("role_id", 0); err != nil {
					return err
				}
			} else {
				if err = d.Set("role_id", resource.Role.ID); err != nil {
					return err
				}
			}

			if resource.Status == nil {
				if err = d.Set("status", ""); err != nil {
					return err
				}
			} else {
				if err = d.Set("status", resource.Status.Value); err != nil {
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

func resourceNetboxDcimRackUpdate(d *schema.ResourceData,
	m interface{}) error {
	client := m.(*netboxclient.NetBoxAPI)
	params := &models.WritableRack{}

	// Required parameters
	name := d.Get("name").(string)
	params.Name = &name

	siteID := int64(d.Get("site_id").(int))
	params.Site = &siteID

	//Optional parameters
	if d.HasChange("asset_tag") {
		assetTag := d.Get("asset_tag").(string)
		params.AssetTag = &assetTag
	}

	if d.HasChange("facility_id") {
		facilityID := d.Get("facility_id").(string)
		params.FacilityID = &facilityID
	}

	if d.HasChange("group_id") {
		groupID := int64(d.Get("group_id").(int))
		if groupID != 0 {
			params.Group = &groupID
		}
	}

	if d.HasChange("role_id") {
		roleID := int64(d.Get("role_id").(int))
		if roleID != 0 {
			params.Role = &roleID
		}
	}

	if d.HasChange("serial") {
		serial := d.Get("serial").(string)
		params.Serial = serial
	}

	if d.HasChange("status") {
		status := d.Get("status").(string)
		params.Status = status
	}

	if d.HasChange("tenant_id") {
		tenantID := int64(d.Get("tenant_id").(int))
		params.Tenant = &tenantID
	}

	if d.HasChange("u_height") {
		uHeight := int64(d.Get("u_height").(int))
		params.UHeight = uHeight
	}

	if d.HasChange("desc_units") {
		descUnits := d.Get("u_height").(bool)
		params.DescUnits = descUnits
	}

	resource := dcim.NewDcimRacksPartialUpdateParams().WithData(
		params)

	resourceID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return fmt.Errorf("Unable to convert ID into int64")
	}

	resource.SetID(resourceID)

	_, err = client.Dcim.DcimRacksPartialUpdate(resource, nil)
	if err != nil {
		return err
	}

	return resourceNetboxDcimRackRead(d, m)
}

func resourceNetboxDcimRackDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*netboxclient.NetBoxAPI)

	resourceExists, err := resourceNetboxDcimRackExists(d, m)
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

	resource := dcim.NewDcimRacksDeleteParams().WithID(id)
	if _, err := client.Dcim.DcimRacksDelete(resource, nil); err != nil {
		return err
	}

	return nil
}

func resourceNetboxDcimRackExists(d *schema.ResourceData, m interface{}) (b bool,
	e error) {
	client := m.(*netboxclient.NetBoxAPI)
	resourceExist := false

	resourceID := d.Id()
	params := dcim.NewDcimRacksListParams().WithID(&resourceID)
	resources, err := client.Dcim.DcimRacksList(params, nil)
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
