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

func resourceNetboxDcimCable() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetboxDcimCableCreate,
		Read:   resourceNetboxDcimCableRead,
		Update: resourceNetboxDcimCableUpdate,
		Delete: resourceNetboxDcimCableDelete,
		Exists: resourceNetboxDcimCableExists,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"status": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "connected",
				ValidateFunc: validation.StringInSlice([]string{"connected", "planned", "decommissioning"}, false),
			},
			"termination_a_id": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"termination_a_type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"dcim.consoleport", "dcim.consoleserverport", "dcim.frontport", "dcim.rearport", "dcim.powerport", "dcim.poweroutlet", "dcim.interface"}, false),
			},
			"termination_b_id": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"termination_b_type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"dcim.consoleport", "dcim.consoleserverport", "dcim.frontport", "dcim.rearport", "dcim.powerport", "dcim.poweroutlet", "dcim.interface"}, false),
			},
		},
	}
}

func resourceNetboxDcimCableCreate(d *schema.ResourceData,
	m interface{}) error {
	client := m.(*netboxclient.NetBoxAPI)

	cableStatus := d.Get("status").(string)
	cableTerminationAId := int64(d.Get("termination_a_id").(int))
	cableTerminationAType := d.Get("termination_a_type").(string)
	cableTerminationBId := int64(d.Get("termination_b_id").(int))
	cableTerminationBType := d.Get("termination_b_type").(string)

	newResource := &models.WritableCable{
		TerminationaID:   &cableTerminationAId,
		TerminationaType: &cableTerminationAType,
		TerminationbID:   &cableTerminationBId,
		TerminationbType: &cableTerminationBType,
	}

	if cableStatus != "" {
		newResource.Status = cableStatus
	}

	resource := dcim.NewDcimCablesCreateParams().WithData(newResource)

	resourceCreated, err := client.Dcim.DcimCablesCreate(resource, nil)
	if err != nil {
		return err
	}

	d.SetId(strconv.FormatInt(resourceCreated.Payload.ID, 10))
	return resourceNetboxDcimCableRead(d, m)
}

func resourceNetboxDcimCableRead(d *schema.ResourceData,
	m interface{}) error {
	client := m.(*netboxclient.NetBoxAPI)

	resourceID := d.Id()
	params := dcim.NewDcimCablesListParams().WithID(&resourceID)
	resources, err := client.Dcim.DcimCablesList(params, nil)
	if err != nil {
		return err
	}

	for _, resource := range resources.Payload.Results {
		if strconv.FormatInt(resource.ID, 10) == d.Id() {
			if resource.Status == nil {
				if err = d.Set("status", ""); err != nil {
					return err
				}
			} else {
				if err = d.Set("status", resource.Status.Value); err != nil {
					return err
				}
			}

			if err = d.Set("termination_a_id", resource.TerminationaID); err != nil {
				return err
			}

			if err = d.Set("termination_a_type", resource.TerminationaType); err != nil {
				return err
			}

			if err = d.Set("termination_b_id", resource.TerminationbID); err != nil {
				return err
			}

			if err = d.Set("termination_b_type", resource.TerminationbType); err != nil {
				return err
			}

			return nil
		}
	}

	d.SetId("")
	return nil
}

func resourceNetboxDcimCableUpdate(d *schema.ResourceData,
	m interface{}) error {
	client := m.(*netboxclient.NetBoxAPI)
	params := &models.WritableCable{}

	// Required parameters
	terminationAId := int64(d.Get("termination_a_id").(int))
	params.TerminationaID = &terminationAId

	terminationAType := d.Get("termination_a_type").(string)
	params.TerminationaType = &terminationAType

	terminationBId := int64(d.Get("termination_b_id").(int))
	params.TerminationbID = &terminationBId

	terminationBType := d.Get("termination_b_type").(string)
	params.TerminationbType = &terminationBType

	// Optional parameters
	if d.HasChange("status") {
		status := d.Get("status").(string)
		params.Status = status
	}

	resource := dcim.NewDcimCablesPartialUpdateParams().WithData(
		params)

	resourceID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return fmt.Errorf("Unable to convert ID into int64")
	}

	resource.SetID(resourceID)

	_, err = client.Dcim.DcimCablesPartialUpdate(resource, nil)
	if err != nil {
		return err
	}

	return resourceNetboxDcimCableRead(d, m)
}

func resourceNetboxDcimCableDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*netboxclient.NetBoxAPI)

	resourceExists, err := resourceNetboxDcimCableExists(d, m)
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

	resource := dcim.NewDcimCablesDeleteParams().WithID(id)
	if _, err := client.Dcim.DcimCablesDelete(resource, nil); err != nil {
		return err
	}

	return nil
}

func resourceNetboxDcimCableExists(d *schema.ResourceData, m interface{}) (b bool,
	e error) {
	client := m.(*netboxclient.NetBoxAPI)
	resourceExist := false

	resourceID := d.Id()
	params := dcim.NewDcimCablesListParams().WithID(&resourceID)
	resources, err := client.Dcim.DcimCablesList(params, nil)
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
