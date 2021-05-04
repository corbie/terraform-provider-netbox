package netbox

import (
	"encoding/json"
	url "net/url"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	netboxclient "github.com/netbox-community/go-netbox/netbox/client"
	"github.com/netbox-community/go-netbox/netbox/client/dcim"
	"github.com/netbox-community/go-netbox/netbox/models"
)

func dataNetboxJSONDcimConsolePortsList() *schema.Resource {
	return &schema.Resource{
		Read: dataNetboxJSONDcimConsolePortsListRead,

		Schema: map[string]*schema.Schema{
			"json": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataNetboxJSONDcimConsolePortsListRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*netboxclient.NetBoxAPI)
	params := dcim.NewDcimConsolePortsListParams()
	acc := []*models.ConsolePort{}
	var offset int64

	// Paginate API results
	// TODO make this generic for all client.*.*List()
	for {
		params.Offset = &offset
		list, err := client.Dcim.DcimConsolePortsList(params, nil)
		if err != nil {
			return err
		}

		acc = append(acc, list.Payload.Results...)

		if list.Payload.Next == nil {
			break
		} else {
			urlStr, err := url.Parse(list.Payload.Next.String())
			if err != nil {
				return err
			}

			newOffset, ok := urlStr.Query()["offset"]
			if !ok {
				break
			}

			if len(newOffset) > 0 {
				tmp, err := strconv.ParseInt(newOffset[0], 10, 64)
				if err != nil {
					return err
				}

				offset = tmp
			}
		}
	}

	j, _ := json.Marshal(acc)

	d.Set("json", string(j))
	d.SetId("NetboxJSONDcimConsolePortsList")

	return nil
}
