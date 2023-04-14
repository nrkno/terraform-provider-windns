# Terraform WinDNS Provider


This Terraform provider allows you to manage your Windows DNS server resources through Terraform. Currently, it supports
managing records of type `AAAA`, `A`, `CNAME`, `TXT` and `PTR`.

## Prerequisites
This provider requires a remote Windows server exposed with SSH and with the
[DnsService](https://learn.microsoft.com/en-us/powershell/module/dnsserver/?view=windowsserver2022-ps)
PowerShell module installed. This server could be the DNS server itself.

## Why use this provider?
Other Terraform providers have implemented similar functionality, but they either require a local Windows installation
running PowerShell or utilize WinRM to execute PowerShell remotely. In many environments, this is not preferable or
possible.

This, and other similar providers are configuring Windows DNS servers using the PowerShell module
[DnsService](https://learn.microsoft.com/en-us/powershell/module/dnsserver/?view=windowsserver2022-ps).
This module uses WinRM internally when talking to the DNS Server.

In an environment where the DNS server is running on a locked-down Domain Controller with WinRM disabled, one will thus
run into problems with the second hop WinRM when going through a jump host. We have yet to find a solution to making the second WinRM hop secure and easily.

This provider avoids the whole second hop concern by using SSH as the transport for the first hop when running PowerShell.


## Usage

Detailed documentation can be found [here](https://registry.terraform.io/providers/nrkno/windns/latest/docs).


### Example 
A minimal terraform configuration:

```
terraform {
  required_providers {
    windns = {
      source  = "nrkno/windns"
      version = "1.0.0"
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

