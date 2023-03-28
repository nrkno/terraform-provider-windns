package provider

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/nrkno/terraform-provider-windns/internal/config"
	"github.com/nrkno/terraform-provider-windns/internal/dnshelper"
)

/*
Prerequisites for acceptance tests

- A Windows DNS server with the following zones configured:
	- example.com
	- 10.10.in-addr.arpa
- A Windows server with SSH enabled and the Powershell DnsServer module installed.
	- This could be the same as running the DNS server, or another to jump through.
*/

const testAccResourceDNSRecordConfigBasicPTR = `
resource "windns_record" "r1" {
  name      = "12.113"
  zone_name = "10.10.in-addr.arpa"
  type      = "PTR"
  records   = ["example-host.example.com."]
}
`

const testAccResourceDSRRecordConfigPTRWithoutDot = `
resource "windns_record" "r1" {
  name      = "12.113"
  zone_name = "10.10.in-addr.arpa"
  type      = "PTR"
  records   = ["example-host.example.com"]
}
`

const testAccResourceDNSRecordConfigBasicA = `
variable "windns_record_name" {}

resource "windns_record" "r1" {
  name      = var.windns_record_name
  zone_name = "example.com"
  type      = "A"
  records   = ["203.0.113.11", "203.0.113.12"]
}
`

const testAccResourceDNSRecordConfigBasicAAAA = `
variable "windns_record_name" {}

resource "windns_record" "r1" {
  name      = var.windns_record_name
  zone_name = "example.com"
  type      = "AAAA"
  records   = ["2001:db8::1", "2001:db8::2"]
}
`

const testAccResourceDNSRecordConfigUpperAAAA = `
variable "windns_record_name" {}

resource "windns_record" "r1" {
  name      = var.windns_record_name
  zone_name = "example.com"
  type      = "AAAA"
  records   = ["2001:DB8::1", "2001:DB8::2"]
}
`

const testAccResourceDNSRecordConfigBasicTXT = `
variable "windns_record_name" {}

resource "windns_record" "r1" {
  name      = var.windns_record_name
  zone_name = "example.com"
  type      = "TXT"
  records   = ["TxTdAtA"]
}
`

const testAccResourceDNSRecordConfigMultiple = `
variable "windns_record_name" {}

resource "windns_record" "r1" {
  name      = var.windns_record_name
  zone_name = "example.com"
  type      = "A"
  records   = ["203.0.113.11", "203.0.113.12"]
}

resource "windns_record" "r2" {
  name      = var.windns_record_name
  zone_name = "example.com"
  type      = "AAAA"
  records   = ["2001:db8::1"]
}

resource "windns_record" "r3" {
  name      = var.windns_record_name
  zone_name = "example.com"
  type      = "TXT"
  records   = ["TXTDATA"]
}
`

const testAccResourceDNSRecordConfigMultipleUpdated = `
variable "windns_record_name" {}

resource "windns_record" "r1" {
  name      = var.windns_record_name
  zone_name = "example.com"
  type      = "A"
  records   = ["203.0.113.21", "203.0.113.22"]
}

resource "windns_record" "r2" {
  name      = var.windns_record_name
  zone_name = "example.com"
  type      = "AAAA"
  records   = ["2001:db8::2"]
}

resource "windns_record" "r3" {
  name      = var.windns_record_name
  zone_name = "example.com"
  type      = "TXT"
  records   = ["UPDATED_DATA"]
}
`

const testAccResourceDNSRecordConfigCNAME = `
variable "windns_record_name" {}

resource "windns_record" "r1" {
  name      = var.windns_record_name
  zone_name = "example.com"
  type      = "CNAME"
  records   = ["cname.example.com"]
}
`

func TestAccResourceDNSRecord_BasicPTR(t *testing.T) {
	envVars := []string{"TF_VAR_windns_record_name"}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t, envVars) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccResourceDNSRecordExists("windns_record.r1", "example-host.example.com.", dnshelper.RecordTypePTR, false),
		),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceDNSRecordConfigBasicPTR,
				Check: resource.ComposeTestCheckFunc(
					testAccResourceDNSRecordExists("windns_record.r1", "example-host.example.com.", dnshelper.RecordTypePTR, true),
				),
			},
			{
				ResourceName:      "windns_record.r1",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceDSRRecord_BasicPTRWithoutDot(t *testing.T) {
	envVars := []string{"TF_VAR_windns_record_name"}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t, envVars) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccResourceDNSRecordExists("windns_record.r1", "example-host.example.com.", dnshelper.RecordTypePTR, false),
		),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceDSRRecordConfigPTRWithoutDot,
				Check: resource.ComposeTestCheckFunc(
					testAccResourceDNSRecordExists("windns_record.r1", "example-host.example.com.", dnshelper.RecordTypePTR, true),
				),
			},
			{
				ResourceName:      "windns_record.r1",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceDNSRecord_BasicA(t *testing.T) {
	envVars := []string{"TF_VAR_windns_record_name"}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t, envVars) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccResourceDNSRecordExists("windns_record.r1", "203.0.113.11", dnshelper.RecordTypeA, false),
		),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceDNSRecordConfigBasicA,
				Check: resource.ComposeTestCheckFunc(
					testAccResourceDNSRecordExists("windns_record.r1", "203.0.113.11", dnshelper.RecordTypeA, true),
				),
			},
			{
				ResourceName:      "windns_record.r1",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceDNSRecord_BasicAAAA(t *testing.T) {
	envVars := []string{"TF_VAR_windns_record_name"}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t, envVars) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccResourceDNSRecordExists("windns_record.r1", "2001:db8::1", dnshelper.RecordTypeAAAA, false),
		),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceDNSRecordConfigBasicAAAA,
				Check: resource.ComposeTestCheckFunc(
					testAccResourceDNSRecordExists("windns_record.r1", "2001:db8::1", dnshelper.RecordTypeAAAA, true),
				),
			},
			{
				ResourceName:      "windns_record.r1",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceDNSRecord_UpperAAAA(t *testing.T) {
	envVars := []string{"TF_VAR_windns_record_name"}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t, envVars) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccResourceDNSRecordExists("windns_record.r1", "2001:db8::1", dnshelper.RecordTypeAAAA, false),
		),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceDNSRecordConfigUpperAAAA,
				Check: resource.ComposeTestCheckFunc(
					testAccResourceDNSRecordExists("windns_record.r1", "2001:db8::1", dnshelper.RecordTypeAAAA, true),
				),
			},
			{
				ResourceName:      "windns_record.r1",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceDNSRecord_BasicTXT(t *testing.T) {
	envVars := []string{"TF_VAR_windns_record_name"}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t, envVars) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccResourceDNSRecordExists("windns_record.r1", "TxTdAtA", dnshelper.RecordTypeTXT, false),
		),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceDNSRecordConfigBasicTXT,
				Check: resource.ComposeTestCheckFunc(
					testAccResourceDNSRecordExists("windns_record.r1", "TxTdAtA", dnshelper.RecordTypeTXT, true),
				),
			},
			{
				ResourceName:      "windns_record.r1",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceDNSRecord_Multiple(t *testing.T) {
	envVars := []string{"TF_VAR_windns_record_name"}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t, envVars) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccResourceDNSRecordExists("windns_record.r1", "203.0.113.11", dnshelper.RecordTypeA, false),
			testAccResourceDNSRecordExists("windns_record.r2", "2001:db8::1", dnshelper.RecordTypeAAAA, false),
			testAccResourceDNSRecordExists("windns_record.r3", "TXTDATA", dnshelper.RecordTypeTXT, false),
		),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceDNSRecordConfigMultiple,
				Check: resource.ComposeTestCheckFunc(
					testAccResourceDNSRecordExists("windns_record.r1", "203.0.113.11", dnshelper.RecordTypeA, true),
					testAccResourceDNSRecordExists("windns_record.r2", "2001:db8::1", dnshelper.RecordTypeAAAA, true),
					testAccResourceDNSRecordExists("windns_record.r3", "TXTDATA", dnshelper.RecordTypeTXT, true),
				),
			},
			{
				ResourceName:      "windns_record.r1",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      "windns_record.r2",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      "windns_record.r3",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceDNSRecord_MultipleUpdated(t *testing.T) {
	envVars := []string{"TF_VAR_windns_record_name"}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t, envVars) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccResourceDNSRecordExists("windns_record.r1", "203.0.113.11", dnshelper.RecordTypeA, false),
			testAccResourceDNSRecordExists("windns_record.r2", "2001:db8::1", dnshelper.RecordTypeAAAA, false),
			testAccResourceDNSRecordExists("windns_record.r3", "TXTDATA", dnshelper.RecordTypeTXT, false),
		),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceDNSRecordConfigMultiple,
				Check: resource.ComposeTestCheckFunc(
					testAccResourceDNSRecordExists("windns_record.r1", "203.0.113.11", dnshelper.RecordTypeA, true),
					testAccResourceDNSRecordExists("windns_record.r2", "2001:db8::1", dnshelper.RecordTypeAAAA, true),
					testAccResourceDNSRecordExists("windns_record.r3", "TXTDATA", dnshelper.RecordTypeTXT, true),
				),
			},
			{
				Config: testAccResourceDNSRecordConfigMultipleUpdated,
				Check: resource.ComposeTestCheckFunc(
					testAccResourceDNSRecordExists("windns_record.r1", "203.0.113.21", dnshelper.RecordTypeA, true),
					testAccResourceDNSRecordExists("windns_record.r2", "2001:db8::2", dnshelper.RecordTypeAAAA, true),
					testAccResourceDNSRecordExists("windns_record.r3", "UPDATED_DATA", dnshelper.RecordTypeTXT, true),
				),
			},
		},
	})
}

func TestAccResourceDNSRecord_CNAME(t *testing.T) {
	envVars := []string{"TF_VAR_windns_record_name"}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t, envVars) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccResourceDNSRecordExists("windns_record.r1", "cname.example.com", dnshelper.RecordTypeCNAME, false),
		),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceDNSRecordConfigCNAME,
				Check: resource.ComposeTestCheckFunc(
					testAccResourceDNSRecordExists("windns_record.r1", "cname.example.com.", dnshelper.RecordTypeCNAME, true),
				),
			},
			{
				ResourceName:      "windns_record.r1",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccResourceDNSRecordExists(resource, expectedRecord string, expectedRecordType string, expected bool) resource.TestCheckFunc {
	ctx := context.Background()
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("%s key not found in state", resource)
		}

		r, err := dnshelper.GetDNSRecordFromId(ctx, testAccProvider.Meta().(*config.ProviderConf), rs.Primary.ID)
		if err != nil {
			if strings.Contains(err.Error(), "ObjectNotFound") && !expected {
				return nil
			}
			return err
		}

		found := false
		for _, record := range r.Records {
			if expectedRecordType == dnshelper.RecordTypeAAAA {
				if strings.ToLower(expectedRecord) == record {
					found = true
				}
			}

			if expectedRecord == record {
				found = true
			}
		}

		if !found {
			return fmt.Errorf("record %s did not contain expected record data %q", r.Id(), expectedRecord)
		}
		return nil
	}
}
