package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/nrkno/terraform-provider-windns/internal/config"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Provider exports the provider schema
func Provider(version string) func() *schema.Provider {
	return func() *schema.Provider {
		p := &schema.Provider{
			Schema: map[string]*schema.Schema{
				"ssh_username": {
					Type:        schema.TypeString,
					Required:    true,
					DefaultFunc: schema.EnvDefaultFunc("WINDNS_SSH_USERNAME", nil),
					Description: "The username used to authenticate to the server's ssh service. (Environment variable: WINDNS_SSH_USERNAME)",
				},
				"ssh_password": {
					Type:        schema.TypeString,
					Required:    true,
					DefaultFunc: schema.EnvDefaultFunc("WINDNS_SSH_PASSWORD", nil),
					Description: "The password used to authenticate to the server's SSH service. (Environment variable: WINDNS_SSH_PASSWORD)",
				},
				"ssh_hostname": {
					Type:        schema.TypeString,
					Required:    true,
					DefaultFunc: schema.EnvDefaultFunc("WINDNS_SSH_HOSTNAME", nil),
					Description: "The hostname of the server we will use to run powershell scripts over SSH. (Environment variable: WINDNS_SSH_HOSTNAME)",
				},
				"dns_server": {
					Type:        schema.TypeString,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("WINDNS_DNS_SERVER_HOSTNAME", nil),
					Default:     "",
					Description: "The hostname of the DNS server. (Environment variable: WINDNS_DNS_SERVER_HOSTNAME)",
				},
			},
			DataSourcesMap: map[string]*schema.Resource{},
			ResourcesMap: map[string]*schema.Resource{
				"windns_record": resourceDNSRecord(),
			},
			ConfigureContextFunc: providerConfigure,
		}
		return p
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (any, diag.Diagnostics) {
	cfg, err := config.NewConfig(d)
	if err != nil {
		return nil, diag.FromErr(err)
	}
	pcfg := config.NewProviderConf(cfg)
	return pcfg, nil
}
