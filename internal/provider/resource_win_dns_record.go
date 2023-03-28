package provider

import (
	"context"
	"fmt"
	"strings"

	"golang.org/x/exp/slices"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/nrkno/terraform-provider-windns/internal/config"
	"github.com/nrkno/terraform-provider-windns/internal/dnshelper"
)

func resourceDNSRecord() *schema.Resource {
	return &schema.Resource{
		Description: "`windns_record` manages DNS Records in an Windows DNS Server.",
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
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				MinItems: 1,
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

func suppressCaseDiff(key, old, new string, d *schema.ResourceData) bool {
	// k is ignored here, but wee need to include it in the function's
	// signature in order to match the one defined for DiffSuppressFunc
	return strings.EqualFold(old, new)
}

func suppressRecordDiff(key, old, new string, d *schema.ResourceData) bool {
	if d.Get("type") == dnshelper.RecordTypePTR {
		var (
			oldRecords []string
			newRecords []string
		)

		// For a list, the key is path to the element, rather than the list.
		// E.g. "windns_record.2.records.0"
		lastDotIndex := strings.LastIndex(key, ".")
		if lastDotIndex != -1 {
			key = key[:lastDotIndex]
		}

		oldData, newData := d.GetChange(key)
		if oldData == nil || newData == nil {
			return false
		}

		for _, v := range oldData.(*schema.Set).List() {
			oldRecords = append(oldRecords, fmt.Sprintf("%s", v))
		}
		// Get-DNSResourceRecord always adds a `.` after the PTR record.
		// to avoid change if the user did not add it, we need to add it before we compare.
		for _, v := range newData.(*schema.Set).List() {
			newRecords = append(newRecords, fmt.Sprintf("%s.", v))
		}
		return slices.Equal(oldRecords, newRecords)
	}
	return strings.EqualFold(old, new)
}
