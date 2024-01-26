// SPDX-License-Identifier: MIT

package provider

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/nrkno/terraform-provider-windns/internal/dnshelper"
	"golang.org/x/exp/slices"
)

func suppressCaseDiff(key, old, new string, d *schema.ResourceData) bool {
	// k is ignored here, but wee need to include it in the function's
	// signature in order to match the one defined for DiffSuppressFunc
	return strings.EqualFold(old, new)
}

func suppressRecordDiff(key, old, new string, d *schema.ResourceData) bool {
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

	rrType := d.Get("type").(string)
	oldRecords := setToStringSlice(oldData.(*schema.Set))
	newRecords := setToStringSlice(newData.(*schema.Set))

	// prevent (known after apply) to be ignored
	if len(oldRecords) == 0 && len(newRecords) == 0 {
		return false
	}

	return suppressRecordDiffForType(oldRecords, newRecords, rrType)
}

func suppressRecordDiffForType(oldRecords, newRecords []string, rrType string) bool {
	slices.Sort(oldRecords)
	slices.Sort(newRecords)

	if rrType == dnshelper.RecordTypePTR || rrType == dnshelper.RecordTypeCNAME {
		return suppressDotDiff(oldRecords, newRecords)
	}
	return suppressListCaseDiff(oldRecords, newRecords)
}

// Get-DNSResourceRecord always returns AAAA records in lower case.
// To avoid change if a user used uppercase, we ignore case.
func suppressListCaseDiff(oldRecords, newRecords []string) bool {
	return slices.EqualFunc(oldRecords, newRecords, func(old, new string) bool {
		return strings.EqualFold(old, new)
	})
}

// Get-DNSResourceRecord always adds a `.` after the PTR and CNAME record types.
// To avoid change if the user did not add it, we need to add it before we compare.
func suppressDotDiff(oldRecords, newRecords []string) bool {
	var newRecordsWithDot []string

	for _, v := range newRecords {
		if strings.HasSuffix(v, ".") {
			newRecordsWithDot = append(newRecordsWithDot, v)
		} else {
			newRecordsWithDot = append(newRecordsWithDot, fmt.Sprintf("%s.", v))
		}
	}
	return slices.Equal(oldRecords, newRecordsWithDot)
}

func setToStringSlice(d *schema.Set) []string {
	var data []string
	for _, v := range d.List() {
		data = append(data, fmt.Sprintf("%s", v))
	}
	return data
}
