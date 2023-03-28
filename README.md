# Terraform WinDNS Provider

This provider allows Terraform to manage DNS records in a Windows DNS server.

Other Terraform providers have implemented similar functionality, but they either require a local Windows installation
with Powershell or utilize WinRM to execute Powershell remotely. Both these things are not preferable.

This provider supports all Terraform platforms and avoids WinRM limitations by using SSH as the transport.
This provider establishes the SSH transport to a (Windows) jump server running Powershell with the DNSServer module 
installed or directly to the server running the DNS server.

If the DNS server is running on a Domain Controller, you may not want to log in directly to that server.

## Getting started

A minimal terraform configuration:

```
terraform {
  required_providers {
    windns = {
      source  = "nrkno/windns"
      version = "0.0.3"
    }
  }
}

provider "windns" {
  ssh_username = "someuser"      # (environment variable WINDNS_SSH_USERNAME)
  ssh_password = "somepassword"  # (environment variable WINDNS_SSH_PASSWORD)
  ssh_hostname = "somehost"      # (environment variable WINDNS_SSH_HOSTNAME)
  
  # Optional
  dns_server   = "someserver"    # (environment variable WINDNS_DNS_SERVER_HOSTNAME) 
}

resource "windns_record" "r" {
  name      = "nrk"
  zone_name = "example.com"
  type      = "A"
  records   = ["203.0.113.11", "203.0.113.12"]
}
```

