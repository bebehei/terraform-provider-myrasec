package myrasec

import (
	"context"
	"log"
	"regexp"
	"strconv"
	"time"

	myrasec "github.com/Myra-Security-GmbH/myrasec-go"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

//
// dataSourceMyrasecDomains ...
//
func dataSourceMyrasecDomains() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceMyrasecDomainsRead,
		Schema: map[string]*schema.Schema{
			"filter": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"match": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"domains": {
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
						"auto_update": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"paused": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"paused_until": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Second),
			Update: schema.DefaultTimeout(30 * time.Second),
		},
	}
}

//
// dataSourceMyrasecDomainsRead ...
//
func dataSourceMyrasecDomainsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*myrasec.API)

	var diags diag.Diagnostics

	f := prepareDomainFilter(d.Get("filter"))
	if f == nil {
		f = &domainFilter{}
	}

	params := map[string]string{}

	if f.regex == nil && len(f.name) > 0 {
		params["search"] = f.name
	}

	domains, err := client.ListDomains(params)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Error fetching domains",
			Detail:   err.Error(),
		})
		return diags
	}

	domainData := make([]interface{}, 0)
	for _, r := range domains {

		if f.regex != nil && !f.regex.MatchString(r.Name) {
			continue
		}

		if f.id != 0 && r.ID != f.id {
			continue
		}

		var created string
		if r.Created != nil {
			created = r.Created.Format(time.RFC3339)
		}

		var modified string
		if r.Created != nil {
			modified = r.Modified.Format(time.RFC3339)
		}

		domainData = append(domainData, map[string]interface{}{
			"id":           r.ID,
			"created":      created,
			"modified":     modified,
			"name":         r.Name,
			"auto_update":  r.AutoUpdate,
			"paused":       r.Paused,
			"paused_until": r.PausedUntil,
		})
	}

	if err := d.Set("domains", domainData); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return nil
}

//
// prepareDomainFilter fetches the panic that can happen in parseDomainFilter
//
func prepareDomainFilter(d interface{}) *domainFilter {
	defer func() {
		if r := recover(); r != nil {
			log.Println("[DEBUG] recovered in prepareDomainFilter", r)
		}
	}()

	return parseDomainFilter(d)
}

//
// parseDomainFilter converts the filter data to a domainFilter struct
//
func parseDomainFilter(d interface{}) *domainFilter {

	cfg := d.([]interface{})
	f := &domainFilter{}

	m := cfg[0].(map[string]interface{})

	id, ok := m["id"]
	if ok {
		f.id = id.(int)
	}

	name, ok := m["name"]
	if ok {
		f.name = name.(string)
	}

	match, ok := m["match"]
	if ok {
		regex, err := regexp.Compile(match.(string))
		if err != nil {
			log.Println("[WARN] The passed regex is not valid", err.Error())

			return f
		}
		f.regex = regex
	}

	return f
}

//
// domainFilter struct ...
//
type domainFilter struct {
	id    int
	name  string
	regex *regexp.Regexp
}
