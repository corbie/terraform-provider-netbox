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

func resourceNetboxDcimSite() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetboxDcimSiteCreate,
		Read:   resourceNetboxDcimSiteRead,
		Update: resourceNetboxDcimSiteUpdate,
		Delete: resourceNetboxDcimSiteDelete,
		Exists: resourceNetboxDcimSiteExists,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 50),
			},
			"slug": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringMatch(
					regexp.MustCompile("^[-a-zA-Z0-9_]{1,50}$"),
					"Must be like ^[-a-zA-Z0-9_]{1,50}$"),
			},
			"status": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "active",
				ValidateFunc: validation.StringInSlice([]string{"active", "planned", "retired"}, false),
			},
			"tenant_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
		},
	}
}

func resourceNetboxDcimSiteCreate(d *schema.ResourceData,
	m interface{}) error {
	client := m.(*netboxclient.NetBoxAPI)

	siteName := d.Get("name").(string)
	siteSlug := d.Get("slug").(string)
	siteStatus := d.Get("status").(string)
	siteTenantID := int64(d.Get("tenant_id").(int))

	newResource := &models.WritableSite{
		Name: &siteName,
		Slug: &siteSlug,
	}

	if siteStatus != "" {
		newResource.Status = siteStatus
	}

	if siteTenantID != 0 {
		newResource.Tenant = &siteTenantID
	}

	resource := dcim.NewDcimSitesCreateParams().WithData(newResource)

	resourceCreated, err := client.Dcim.DcimSitesCreate(resource, nil)
	if err != nil {
		return err
	}

	d.SetId(strconv.FormatInt(resourceCreated.Payload.ID, 10))
	return resourceNetboxDcimSiteRead(d, m)
}

func resourceNetboxDcimSiteRead(d *schema.ResourceData,
	m interface{}) error {
	client := m.(*netboxclient.NetBoxAPI)

	resourceID := d.Id()
	params := dcim.NewDcimSitesListParams().WithID(&resourceID)
	resources, err := client.Dcim.DcimSitesList(params, nil)
	if err != nil {
		return err
	}

	for _, resource := range resources.Payload.Results {
		if strconv.FormatInt(resource.ID, 10) == d.Id() {
			if err = d.Set("name", resource.Name); err != nil {
				return err
			}

			if err = d.Set("slug", resource.Slug); err != nil {
				return err
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

func resourceNetboxDcimSiteUpdate(d *schema.ResourceData,
	m interface{}) error {
	client := m.(*netboxclient.NetBoxAPI)
	params := &models.WritableSite{}

	// Required parameters
	name := d.Get("name").(string)
	params.Name = &name

	slug := d.Get("slug").(string)
	params.Slug = &slug

	// Optional parameters
	if d.HasChange("status") {
		status := d.Get("status").(string)
		params.Status = status
	}

	if d.HasChange("tenant_id") {
		tenantID := d.Get("tenant_id").(int64)
		params.Tenant = &tenantID
	}

	resource := dcim.NewDcimSitesPartialUpdateParams().WithData(
		params)

	resourceID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return fmt.Errorf("Unable to convert ID into int64")
	}

	resource.SetID(resourceID)

	_, err = client.Dcim.DcimSitesPartialUpdate(resource, nil)
	if err != nil {
		return err
	}

	return resourceNetboxDcimSiteRead(d, m)
}

func resourceNetboxDcimSiteDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*netboxclient.NetBoxAPI)

	resourceExists, err := resourceNetboxDcimSiteExists(d, m)
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

	resource := dcim.NewDcimSitesDeleteParams().WithID(id)
	if _, err := client.Dcim.DcimSitesDelete(resource, nil); err != nil {
		return err
	}

	return nil
}

func resourceNetboxDcimSiteExists(d *schema.ResourceData, m interface{}) (b bool,
	e error) {
	client := m.(*netboxclient.NetBoxAPI)
	resourceExist := false

	resourceID := d.Id()
	params := dcim.NewDcimSitesListParams().WithID(&resourceID)
	resources, err := client.Dcim.DcimSitesList(params, nil)
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
