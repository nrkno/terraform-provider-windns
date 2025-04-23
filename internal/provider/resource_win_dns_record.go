// SPDX-License-Identifier: MIT

package provider

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/nrkno/terraform-provider-windns/internal/config"
	"github.com/nrkno/terraform-provider-windns/internal/dnshelper"
)

func resourceDNSRecord() *schema.Resource {
	return &schema.Resource{
		Description: "`windns_record` manages DNS Records in a Windows DNS Server.",
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		ReadContext:   resourceDNSRecordRead,
		CreateContext: resourceDNSRecordCreate,
		UpdateContext: resourceDNSRecordUpdate,
		DeleteContext: resourceDNSRecordDelete,
		Schema: map[string]*schema.Schema{
			"zone_name": {
				Type:             schema.TypeString,
				Required:         true,
				DiffSuppressFunc: suppressCaseDiff,
				Description:      "The zone name for the dns records.",
			},
			"name": {
				Type:             schema.TypeString,
				Required:         true,
				DiffSuppressFunc: suppressCaseDiff,
				Description:      "The name of the dns records.",
			},
			"type": {
				Type:             schema.TypeString,
				Required:         true,
				DiffSuppressFunc: suppressCaseDiff,
				Description:      "The type of the dns records.",
			},
			"records": {
				Type:             schema.TypeSet,
				Required:         true,
				Description:      "A list of records.",
				DiffSuppressFunc: suppressRecordDiff,
				Set:              schema.HashString,
				Elem:             &schema.Schema{Type: schema.TypeString},
				MinItems:         1,
			},
			"create_ptr": {
				Type:        schema.TypeBool,
				Required:    false,
				Optional:    true,
				Description: "Create PTR records for requested (A or AAAA) records.",
			},
		},
		CustomizeDiff: customdiff.All(
			customdiff.ForceNewIfChange("zone_name", func(ctx context.Context, old, new, meta any) bool {
				return new.(string) != old.(string)
			}),
			customdiff.ForceNewIfChange("name", func(ctx context.Context, old, new, meta any) bool {
				return new.(string) != old.(string)
			}),
			customdiff.ForceNewIfChange("type", func(ctx context.Context, old, new, meta any) bool {
				return new.(string) != old.(string)
			}),
			customdiff.ForceNewIfChange("type", func(ctx context.Context, old, new, meta any) bool {
				return new.(string) != old.(string)
			}),
		),
	}
}

func resourceDNSRecordCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	record := dnshelper.NewDNSRecordFromResource(d)

	id, err := record.Create(meta.(*config.ProviderConf))
	if err != nil {
		return diag.Errorf("error while creating new record object: %s", err)
	}
	d.SetId(id)

	return resourceDNSRecordRead(ctx, d, meta)
}

func resourceDNSRecordRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if d.Id() == "" {
		return nil
	}

	record, err := dnshelper.GetDNSRecordFromId(ctx, meta.(*config.ProviderConf), d.Id())
	if err != nil {
		if strings.Contains(err.Error(), "ObjectNotFound") {
			// Resource no longer exists
			d.SetId("")
			return nil
		}
		return diag.Errorf("error while reading record with id %q: %s", d.Id(), err)
	}

	_ = d.Set("zone_name", record.ZoneName)
	_ = d.Set("name", record.HostName)
	_ = d.Set("type", record.RecordType)
	_ = d.Set("records", record.Records)
	_ = d.Set("create_ptr", record.CreatePtr)

	return nil
}

func resourceDNSRecordUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	record := dnshelper.NewDNSRecordFromResource(d)
	keys := []string{"records"}
	changes := make(map[string]interface{})
	for _, key := range keys {
		if d.HasChange(key) {
			changes[key] = d.Get(key)
		}
	}

	err := record.Update(ctx, meta.(*config.ProviderConf), changes)
	if err != nil {
		return diag.Errorf("error while updating record with id %q: %s", d.Id(), err)
	}
	return resourceDNSRecordRead(ctx, d, meta)
}

func resourceDNSRecordDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if d.Id() == "" {
		return nil
	}
	record := dnshelper.NewDNSRecordFromResource(d)
	err := record.Delete(meta.(*config.ProviderConf))
	if err != nil {
		return diag.Errorf("error while deleting a record object with id %q: %s", d.Id(), err)
	}

	return nil
}
