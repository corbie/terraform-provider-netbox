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

func resourceNetboxDcimDevice() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetboxDcimDeviceCreate,
		Read:   resourceNetboxDcimDeviceRead,
		Update: resourceNetboxDcimDeviceUpdate,
		Delete: resourceNetboxDcimDeviceDelete,
		Exists: resourceNetboxDcimDeviceExists,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"asset_tag": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 50),
			},
			"cluster_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"device_role_id": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"device_type_id": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"face": {
				Type:         schema.TypeString,
				Default:      "front",
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"front", "rear"}, false),
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 50),
			},
			"parent_device_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"position": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(1, 32767),
			},
			"platform_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"primary_ip4_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"primary_ip6_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"rack_id": {
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
				ValidateFunc: validation.StringInSlice([]string{"offline", "active", "planned", "staged", "failed", "inventory", "decommissioning"}, false),
			},
			"tenant_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"virtual_chassis_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
		},
	}
}

func resourceNetboxDcimDeviceCreate(d *schema.ResourceData,
	m interface{}) error {
	client := m.(*netboxclient.NetBoxAPI)

	deviceAssetTag := d.Get("asset_tag").(string)
	deviceClusterID := int64(d.Get("cluster_id").(int))
	deviceDeviceRoleID := int64(d.Get("device_role_id").(int))
	deviceDeviceTypeID := int64(d.Get("device_type_id").(int))
	deviceFace := d.Get("face").(string)
	deviceName := d.Get("name").(string)
	deviceParentDeviceID := int64(d.Get("parent_device_id").(int))
	devicePosition := int64(d.Get("position").(int))
	devicePlatformID := int64(d.Get("platform_id").(int))
	devicePrimaryIP4ID := int64(d.Get("primary_ip4_id").(int))
	devicePrimaryIP6ID := int64(d.Get("primary_ip6_id").(int))
	deviceRackID := int64(d.Get("rack_id").(int))
	deviceSerial := d.Get("serial").(string)
	deviceSiteID := int64(d.Get("site_id").(int))
	deviceStatus := d.Get("status").(string)
	deviceTenantID := int64(d.Get("tenant_id").(int))
	deviceVirtualChassisID := int64(d.Get("virtual_chassis_id").(int))

	newResource := &models.WritableDeviceWithConfigContext{
		DeviceRole: &deviceDeviceRoleID,
		DeviceType: &deviceDeviceTypeID,
		Name:       &deviceName,
		Site:       &deviceSiteID,
	}

	if deviceAssetTag != "" {
		newResource.AssetTag = &deviceAssetTag
	}

	if deviceClusterID != 0 {
		newResource.Cluster = &deviceClusterID
	}

	if deviceFace != "" {
		newResource.Face = deviceFace
	}

	if deviceParentDeviceID != 0 {
		newResource.ParentDevice = &models.NestedDevice{ID: deviceParentDeviceID}
	}

	if devicePosition != 0 {
		newResource.Position = &devicePosition
	}

	if devicePlatformID != 0 {
		newResource.Platform = &devicePlatformID
	}

	if devicePrimaryIP4ID != 0 {
		newResource.PrimaryIp4 = &devicePrimaryIP4ID
	}

	if devicePrimaryIP6ID != 0 {
		newResource.PrimaryIp6 = &devicePrimaryIP6ID
	}

	if deviceRackID != 0 {
		newResource.Rack = &deviceRackID
	}

	if deviceSerial != "" {
		newResource.Serial = deviceSerial
	}

	if deviceStatus != "" {
		newResource.Status = deviceStatus
	}

	if deviceTenantID != 0 {
		newResource.Tenant = &deviceTenantID
	}

	if deviceVirtualChassisID != 0 {
		newResource.VirtualChassis = &deviceVirtualChassisID
	}

	resource := dcim.NewDcimDevicesCreateParams().WithData(newResource)

	resourceCreated, err := client.Dcim.DcimDevicesCreate(resource, nil)
	if err != nil {
		return err
	}

	d.SetId(strconv.FormatInt(resourceCreated.Payload.ID, 10))
	return resourceNetboxDcimDeviceRead(d, m)
}

func resourceNetboxDcimDeviceRead(d *schema.ResourceData,
	m interface{}) error {
	client := m.(*netboxclient.NetBoxAPI)

	resourceID := d.Id()
	params := dcim.NewDcimDevicesListParams().WithID(&resourceID)
	resources, err := client.Dcim.DcimDevicesList(params, nil)
	if err != nil {
		return err
	}

	for _, resource := range resources.Payload.Results {
		if strconv.FormatInt(resource.ID, 10) == d.Id() {
			if err = d.Set("asset_tag", resource.AssetTag); err != nil {
				return err
			}

			if resource.Cluster == nil {
				if err = d.Set("cluster_id", 0); err != nil {
					return err
				}
			} else {
				if err = d.Set("cluster_id", resource.Cluster.ID); err != nil {
					return err
				}
			}

			if err = d.Set("name", resource.Name); err != nil {
				return err
			}

			if err = d.Set("position", resource.Position); err != nil {
				return err
			}

			if resource.Platform == nil {
				if err = d.Set("platform_id", 0); err != nil {
					return err
				}
			} else {
				if err = d.Set("platform_id", resource.Platform.ID); err != nil {
					return err
				}
			}

			if resource.PrimaryIp4 == nil {
				if err = d.Set("primary_ip4_id", 0); err != nil {
					return err
				}
			} else {
				if err = d.Set("primary_ip4_id", resource.PrimaryIp4.ID); err != nil {
					return err
				}
			}

			if resource.PrimaryIp6 == nil {
				if err = d.Set("primary_ip6_id", 0); err != nil {
					return err
				}
			} else {
				if err = d.Set("primary_ip6_id", resource.PrimaryIp6.ID); err != nil {
					return err
				}
			}

			if resource.Rack == nil {
				if err = d.Set("rack_id", 0); err != nil {
					return err
				}
			} else {
				if err = d.Set("rack_id", resource.Rack.ID); err != nil {
					return err
				}
			}

			if resource.VirtualChassis == nil {
				if err = d.Set("virtual_chassis_id", 0); err != nil {
					return err
				}
			} else {
				if err = d.Set("virtual_chassis_id", resource.VirtualChassis.ID); err != nil {
					return err
				}
			}

			if resource.DeviceRole == nil {
				if err = d.Set("device_role_id", 0); err != nil {
					return err
				}
			} else {
				if err = d.Set("device_role_id", resource.DeviceRole.ID); err != nil {
					return err
				}
			}

			if resource.DeviceType == nil {
				if err = d.Set("device_type_id", 0); err != nil {
					return err
				}
			} else {
				if err = d.Set("device_type_id", resource.DeviceType.ID); err != nil {
					return err
				}
			}

			if resource.Face == nil {
				if err = d.Set("face", ""); err != nil {
					return err
				}
			} else {
				if err = d.Set("face", resource.Face.Value); err != nil {
					return err
				}
			}

			if resource.ParentDevice == nil {
				if err = d.Set("parent_device_id", 0); err != nil {
					return err
				}
			} else {
				if err = d.Set("parent_device_id", resource.ParentDevice.ID); err != nil {
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

			if resource.Site == nil {
				if err = d.Set("site_id", ""); err != nil {
					return err
				}
			} else {
				if err = d.Set("site_id", resource.Site.ID); err != nil {
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

			return nil
		}
	}

	d.SetId("")
	return nil
}

func resourceNetboxDcimDeviceUpdate(d *schema.ResourceData,
	m interface{}) error {
	client := m.(*netboxclient.NetBoxAPI)
	params := &models.WritableDeviceWithConfigContext{}

	// Required parameters
	deviceRoleID := int64(d.Get("device_role_id").(int))
	params.DeviceRole = &deviceRoleID

	deviceTypeID := int64(d.Get("device_type_id").(int))
	params.DeviceType = &deviceTypeID

	siteID := int64(d.Get("site_id").(int))
	params.Site = &siteID

	name := d.Get("name").(string)
	params.Name = &name

	// Optional parameters
	if d.HasChange("asset_tag") {
		assetTag := d.Get("asset_tag").(string)
		if assetTag != "" {
			params.AssetTag = &assetTag
		}
	}

	if d.HasChange("cluster_id") {
		clusterID := int64(d.Get("cluster_id").(int))
		if clusterID != 0 {
			params.Cluster = &clusterID
		}
	}

	if d.HasChange("face") {
		face := d.Get("face").(string)
		if face != "" {
			params.Face = face
		}
	}

	if d.HasChange("parent_device_id") {
		parentDeviceID := int64(d.Get("parent_device_id").(int))
		if parentDeviceID != 0 {
			params.ParentDevice = &models.NestedDevice{ID: parentDeviceID}
		}
	}

	if d.HasChange("position") {
		position := int64(d.Get("position").(int))
		if position != 0 {
			params.Position = &position
		}
	}

	if d.HasChange("platform_id") {
		platformID := int64(d.Get("platform_id").(int))
		if platformID != 0 {
			params.Platform = &platformID
		}
	}

	if d.HasChange("primary_ip4_id") {
		primaryIP4 := int64(d.Get("primary_ip4_id").(int))
		if primaryIP4 != 0 {
			params.PrimaryIp4 = &primaryIP4
		}
	}

	if d.HasChange("primary_ip6_id") {
		primaryIP6 := int64(d.Get("primary_ip6_id").(int))
		if primaryIP6 != 0 {
			params.PrimaryIp6 = &primaryIP6
		}
	}

	if d.HasChange("rack_id") {
		rackID := int64(d.Get("rack_id").(int))
		if rackID != 0 {
			params.Rack = &rackID
		}
	}

	if d.HasChange("serial") {
		serial := d.Get("serial").(string)
		if serial != "" {
			params.Serial = serial
		}
	}

	if d.HasChange("status") {
		status := d.Get("status").(string)
		if status != "" {
			params.Status = status
		}
	}

	if d.HasChange("tenant_id") {
		tenantID := int64(d.Get("tenant_id").(int))
		if tenantID != 0 {
			params.Tenant = &tenantID
		}
	}

	if d.HasChange("virtual_chassis_id") {
		virtualChassisID := int64(d.Get("virtual_chassis_id").(int))
		if virtualChassisID != 0 {
			params.VirtualChassis = &virtualChassisID
		}
	}

	resource := dcim.NewDcimDevicesPartialUpdateParams().WithData(
		params)

	resourceID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return fmt.Errorf("Unable to convert ID into int64")
	}

	resource.SetID(resourceID)

	_, err = client.Dcim.DcimDevicesPartialUpdate(resource, nil)
	if err != nil {
		return err
	}

	return resourceNetboxDcimDeviceRead(d, m)
}

func resourceNetboxDcimDeviceDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*netboxclient.NetBoxAPI)

	resourceExists, err := resourceNetboxDcimDeviceExists(d, m)
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

	resource := dcim.NewDcimDevicesDeleteParams().WithID(id)
	if _, err := client.Dcim.DcimDevicesDelete(resource, nil); err != nil {
		return err
	}

	return nil
}

func resourceNetboxDcimDeviceExists(d *schema.ResourceData, m interface{}) (b bool,
	e error) {
	client := m.(*netboxclient.NetBoxAPI)
	resourceExist := false

	resourceID := d.Id()
	params := dcim.NewDcimDevicesListParams().WithID(&resourceID)
	resources, err := client.Dcim.DcimDevicesList(params, nil)
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
