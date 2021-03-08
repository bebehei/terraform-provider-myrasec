package myrasec

import (
	"fmt"
	"strconv"
	"time"

	myrasec "github.com/Myra-Security-GmbH/myrasec-go"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

//
// dataSourceWAFActions ...
//
func dataSourceWAFActions() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceWAFActionsRead,
		Schema: map[string]*schema.Schema{
			"filter": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"waf_actions": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"modified": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"created": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"custom_key": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"value": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"available_phases": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"force_custom_values": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

//
// dataSourceWAFActionsRead ...
//
func dataSourceWAFActionsRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*myrasec.API)

	actions, err := client.ListWAFActions()
	if err != nil {
		return fmt.Errorf("Error fetching WAF actions: %s", err)
	}

	var filter string
	cfg := d.Get("filter").([]interface{})
	m := cfg[0].(map[string]interface{})
	actionType, ok := m["type"]
	if ok {
		filter = actionType.(string)
	}

	wafActionData := make([]interface{}, 0)
	for _, r := range actions {
		if len(filter) > 0 && r.Type != filter {
			continue
		}

		wafActionData = append(wafActionData, map[string]interface{}{
			"id":                  r.ID,
			"created":             r.Created.Format(time.RFC3339),
			"modified":            r.Modified.Format(time.RFC3339),
			"name":                r.Name,
			"available_phases":    r.AvailablePhases,
			"custom_key":          r.CustomKey,
			"force_custom_values": r.ForceCustomValues,
			"type":                r.Type,
			"value":               r.Value,
		})
	}

	if err := d.Set("waf_actions", wafActionData); err != nil {
		return err
	}

	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return nil
}
