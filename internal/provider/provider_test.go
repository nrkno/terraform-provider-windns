package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var testAccProviderFactories map[string]func() (*schema.Provider, error)
var testAccProvider *schema.Provider

func init() {
	testAccProvider = Provider("dev")()
	testAccProviderFactories = map[string]func() (*schema.Provider, error){
		"windns": func() (*schema.Provider, error) {
			return testAccProvider, nil
		},
	}

}

func TestProvider(t *testing.T) {
	if err := Provider("dev")().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func testAccPreCheck(t *testing.T, envVars []string) {
	for _, envVar := range envVars {
		if val := os.Getenv(envVar); val == "" {
			t.Fatalf("%s must be set for acceptance tests to work", envVar)
		}
	}
}
