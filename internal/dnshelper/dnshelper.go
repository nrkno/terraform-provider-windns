// SPDX-License-Identifier: MIT

package dnshelper

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var recordInputPattern = regexp.MustCompile(`^[a-zA-Z0-9:.\-_]+$`)

func SanitizeInputString(recordType string, input string) (string, error) {
	if recordType == "TXT" {
		if len(input) > 255 {
			return "", fmt.Errorf("TXT record can only be 255 characters long")
		}
		return escapePowerShellInput(input), nil
	}

	if recordInputPattern.MatchString(input) {
		return input, nil
	}
	return "", fmt.Errorf("invalid characters detected in input: %s", input)
}

func SanitiseTFInput(d *schema.ResourceData, key string) (string, error) {
	return SanitizeInputString(d.Get("type").(string), d.Get(key).(string))
}

func escapePowerShellInput(input string) string {
	replacer := strings.NewReplacer(
		"`", "``",
		";", "`;",
		"&", "`&",
		"(", "`(",
		")", "`)",
	)
	return replacer.Replace(input)
}
