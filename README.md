# Terraform WinDNS Provider

This provider allows Terraform to manage DNS records in a Windows DNS server.

This provider utilizes SSH as the transport from the machine running Terraform to jump host, or the DNS server itself
where Powershell commands are issues towards the DNS Server.

Other providers offer WinDNS provisioning but rely on WinRM as transport. SSH is used instead in this provider to avoid WinRM double-hop issues when you don't want to expose WinRM directly
on your domain controller, for instance.



## Getting started

A minimal terraform configuration:

```
terraform {
  required_providers {
    windns = {
      source  = "nrkno/windns"
      version = "0.0.1"
    }
  }
}

provider "windns" {
  # configured by environment
  # WINDNS_SSH_USERNAME
  # WINDNS_SSH_PASSWORD
  # WINDNS_SSH_HOSTNAME
  # WINDNS_DNS_SERVER_HOSTNAME
}

resource "windns_record" "r" {
  name      = "nrk"
  zone_name = "example.com"
  type      = "A"
  records   = ["203.0.113.11", "203.0.113.12"]
}
```

