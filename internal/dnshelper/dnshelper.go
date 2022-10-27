package dnshelper

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"strings"
)

func SanitiseTFInput(d *schema.ResourceData, key string) string {
	return SanitiseString(d.Get(key).(string))
}

func SanitiseString(key string) string {
	cleanupReplacer := strings.NewReplacer(
		"`", "``",
		`"`, "`\"",
		"$", "`$",
		"\x00", "`0",
		"\x07", "`a",
		"\x08", "`b",
		"\x1f", "`e",
		"\x0c", "`f",
		"\n", "`n",
		"\r", "`r",
		"\t", "`t",
		"\v", "`v",
	)
	out := cleanupReplacer.Replace(key)
	return out
}
