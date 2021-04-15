package netbox

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	netboxclient "github.com/netbox-community/go-netbox/netbox/client"
	"github.com/netbox-community/go-netbox/netbox/client/dcim"
	"github.com/netbox-community/go-netbox/netbox/models"
)

func resourceNetboxDcimInterface() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetboxDcimInterfaceCreate,
		Read:   resourceNetboxDcimInterfaceRead,
		Update: resourceNetboxDcimInterfaceUpdate,
		Delete: resourceNetboxDcimInterfaceDelete,
		Exists: resourceNetboxDcimInterfaceExists,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"device_id": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 64),
			},
			"type": {
				Type:     schema.TypeString,
				Required: true,
			},
			"enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"lag_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"mtu": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(1, 65536),
			},
			"mac_address": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringMatch(
					regexp.MustCompile("^([A-Z0-9]{2}:){5}[A-Z0-9]{2}$"),
					"Must be like AA:AA:AA:AA:AA"),
			},
			"mgmt_only": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      " ",
				ValidateFunc: validation.StringLenBetween(1, 200),
			},
			"connection_status": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"mode": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{"access", "tagged",
					"tagged-all"}, false),
			},
			"untagged_vlan_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"tagged_vlans": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
				Optional: true,
			},
			"tag": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"slug": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
		},
	}
}

func resourceNetboxDcimInterfaceCreate(d *schema.ResourceData,
	m interface{}) error {
	client := m.(*netboxclient.NetBoxAPI)

	deviceID := int64(d.Get("device_id").(int))
	name := d.Get("name").(string)
	infType := d.Get("type").(string)
	enabled := d.Get("enabled").(bool)
	lagID := int64(d.Get("lag_id").(int))
	mtu := int64(d.Get("mtu").(int))
	macAddress := d.Get("mac_address").(string)
	mgmtOnly := d.Get("mgmt_only").(bool)
	description := d.Get("description").(string)
	connectionStatus := d.Get("connection_status").(bool)
	mode := d.Get("mode").(string)
	untaggedVlanID := int64(d.Get("untagged_vlan_id").(int))
	taggedVlans := d.Get("tagged_vlans").(*schema.Set).List()
	tags := d.Get("tag").(*schema.Set).List()

	newResource := &models.WritableInterface{
		Device: &deviceID,
		Name:   &name,
		Type:   &infType,
	}

	if enabled != false {
		newResource.Enabled = enabled
	}

	if lagID != 0 {
		newResource.Lag = &lagID
	}

	if mtu != 0 {
		newResource.Mtu = &mtu
	}

	if macAddress != "" {
		newResource.MacAddress = &macAddress
	}

	newResource.MgmtOnly = mgmtOnly

	if description != "" {
		newResource.Description = description
	}

	if connectionStatus != false {
		newResource.ConnectionStatus = &connectionStatus
	}

	if mode != "" {
		newResource.Mode = mode
	}

	if untaggedVlanID != 0 {
		newResource.UntaggedVlan = &untaggedVlanID
	}

	if taggedVlans != nil {
		newResource.TaggedVlans = expandToInt64Slice(taggedVlans)
	}

	if tags != nil {
		newResource.Tags = convertTagsToNestedTags(tags)
	}

	resource := dcim.NewDcimInterfacesCreateParams().WithData(newResource)

	resourceCreated, err := client.Dcim.DcimInterfacesCreate(
		resource, nil)

	if err != nil {
		return err
	}

	d.SetId(strconv.FormatInt(resourceCreated.Payload.ID, 10))

	return resourceNetboxDcimInterfaceRead(d, m)
}

func resourceNetboxDcimInterfaceRead(d *schema.ResourceData,
	m interface{}) error {
	client := m.(*netboxclient.NetBoxAPI)

	resourceID := d.Id()
	params := dcim.NewDcimInterfacesListParams().WithID(&resourceID)
	resources, err := client.Dcim.DcimInterfacesList(
		params, nil)

	if err != nil {
		return err
	}

	for _, resource := range resources.Payload.Results {
		if strconv.FormatInt(resource.ID, 10) == d.Id() {
			if resource.Device == nil {
				if err = d.Set("device_id", 0); err != nil {
					return err
				}
			} else {
				if err = d.Set("device_id", resource.Device.ID); err != nil {
					return err
				}
			}

			if err = d.Set("name", resource.Name); err != nil {
				return err
			}

			if err = d.Set("type", resource.Type.Value); err != nil {
				return err
			}

			if err = d.Set("enabled", resource.Enabled); err != nil {
				return err
			}

			if resource.Lag == nil {
				if err = d.Set("lag_id", 0); err != nil {
					return err
				}
			} else {
				if err = d.Set("lag", resource.Lag.ID); err != nil {
					return err
				}
			}

			if err = d.Set("mtu", resource.Mtu); err != nil {
				return err
			}

			if err = d.Set("mac_address", resource.MacAddress); err != nil {
				return err
			}

			if err = d.Set("mgmt_only", resource.MgmtOnly); err != nil {
				return err
			}

			var description string

			if resource.Description == "" {
				description = " "
			} else {
				description = resource.Description
			}

			if err = d.Set("description", description); err != nil {
				return err
			}

			if resource.ConnectionStatus == nil {
				if err = d.Set("connection_status", false); err != nil {
					return err
				}
			} else {
				if err = d.Set("connection_status", resource.ConnectionStatus.Value); err != nil {
					return err
				}
			}

			if resource.Mode == nil {
				if err = d.Set("mode", ""); err != nil {
					return err
				}
			} else {
				if err = d.Set("mode", resource.Mode.Value); err != nil {
					return err
				}
			}

			if resource.UntaggedVlan == nil {
				if err = d.Set("untagged_vlan_id", 0); err != nil {
					return err
				}
			} else {
				if err = d.Set("untagged_vlan_id", resource.UntaggedVlan.ID); err != nil {
					return err
				}
			}

			if err = d.Set("tagged_vlans", resource.TaggedVlans); err != nil {
				return err
			}

			if err = d.Set("tag", convertNestedTagsToTags(
				resource.Tags)); err != nil {
				return err
			}

			return nil
		}
	}

	d.SetId("")
	return nil
}

func resourceNetboxDcimInterfaceUpdate(d *schema.ResourceData,
	m interface{}) error {
	client := m.(*netboxclient.NetBoxAPI)
	params := &models.WritableInterface{}

	// Required parameters
	deviceID := int64(d.Get("device_id").(int))
	params.Device = &deviceID

	name := d.Get("name").(string)
	params.Name = &name

	infType := d.Get("type").(string)
	params.Type = &infType

	taggedVlans := d.Get("tagged_vlans").(*schema.Set).List()
	params.TaggedVlans = expandToInt64Slice(taggedVlans)

	// Optional parameters
	if d.HasChange("enabled") {
		enabled := d.Get("enabled").(bool)
		params.Enabled = enabled
	}

	if d.HasChange("lag_id") {
		lagID := int64(d.Get("lag_id").(int))
		params.Lag = &lagID
	}

	if d.HasChange("mtu") {
		mtu := int64(d.Get("mtu").(int))
		params.Mtu = &mtu
	}

	if d.HasChange("mac_address") {
		macAddress := d.Get("mac_address").(string)
		params.MacAddress = &macAddress
	}

	if d.HasChange("mgmt_only") {
		mgmtOnly := d.Get("mgmt_only").(bool)
		params.MgmtOnly = mgmtOnly
	}

	if d.HasChange("description") {
		description := d.Get("description").(string)
		params.Description = description
	}

	if d.HasChange("connection_status") {
		connectionStatus := d.Get("connection_status").(bool)
		params.ConnectionStatus = &connectionStatus
	}

	if d.HasChange("mode") {
		mode := d.Get("mode").(string)
		params.Mode = mode
	}

	if d.HasChange("untagged_vlan_id") {
		untaggedVlanID := int64(d.Get("untagged_vlan_id").(int))
		params.UntaggedVlan = &untaggedVlanID
	}

	tags := d.Get("tag").(*schema.Set).List()
	params.Tags = convertTagsToNestedTags(tags)

	resource := dcim.NewDcimInterfacesPartialUpdateParams().WithData(params)

	resourceID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return fmt.Errorf("Unable to convert ID into int64")
	}

	resource.SetID(resourceID)

	_, err = client.Dcim.DcimInterfacesPartialUpdate(
		resource, nil)
	if err != nil {
		return err
	}

	return resourceNetboxDcimInterfaceRead(d, m)
}

func resourceNetboxDcimInterfaceDelete(d *schema.ResourceData,
	m interface{}) error {
	client := m.(*netboxclient.NetBoxAPI)

	resourceExists, err := resourceNetboxDcimInterfaceExists(d, m)
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

	p := dcim.NewDcimInterfacesDeleteParams().WithID(id)
	if _, err := client.Dcim.DcimInterfacesDelete(
		p, nil); err != nil {
		return err
	}

	return nil
}

func resourceNetboxDcimInterfaceExists(d *schema.ResourceData,
	m interface{}) (b bool,
	e error) {
	client := m.(*netboxclient.NetBoxAPI)
	resourceExist := false

	resourceID := d.Id()
	params := dcim.NewDcimInterfacesListParams().WithID(
		&resourceID)
	resources, err := client.Dcim.DcimInterfacesList(
		params, nil)
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
