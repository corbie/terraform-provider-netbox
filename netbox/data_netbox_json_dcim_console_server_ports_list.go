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

func dataNetboxJSONDcimConsoleServerPortsList() *schema.Resource {
	return &schema.Resource{
		Read: dataNetboxJSONDcimConsoleServerPortsListRead,

		Schema: map[string]*schema.Schema{
			"json": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataNetboxJSONDcimConsoleServerPortsListRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*netboxclient.NetBoxAPI)
	params := dcim.NewDcimConsoleServerPortsListParams()
	acc := []*models.ConsoleServerPort{}
	var offset int64

	// Paginate API results
	// TODO make this generic for all client.*.*List()
	for {
		params.Offset = &offset
		list, err := client.Dcim.DcimConsoleServerPortsList(params, nil)
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
	d.SetId("NetboxJSONDcimConsoleServerPortsList")

	return nil
}
