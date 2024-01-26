// SPDX-License-Identifier: MIT

package config

import (
	"sync"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/melbahja/goph"
)

type Settings struct {
	SshUsername string
	SshPassword string
	SshHostname string
	DnsServer   string
	Version     string
}

func NewConfig(d *schema.ResourceData) (*Settings, error) {
	sshUsername := d.Get("ssh_username").(string)
	sshPassword := d.Get("ssh_password").(string)
	sshHost := d.Get("ssh_hostname").(string)
	dnsServer := d.Get("dns_server").(string)

	cfg := &Settings{
		SshHostname: sshHost,
		SshUsername: sshUsername,
		SshPassword: sshPassword,
		DnsServer:   dnsServer,
	}

	return cfg, nil
}

func GetSSHConnection(settings *Settings) (*goph.Client, error) {
	auth := goph.Password(settings.SshPassword)
	client, err := goph.NewUnknown(settings.SshUsername, settings.SshHostname, auth)
	if err != nil {
		return nil, err
	}

	return client, err
}

type ProviderConf struct {
	Settings   *Settings
	sshClients []*goph.Client
	mx         *sync.Mutex
}

func NewProviderConf(settings *Settings) *ProviderConf {
	pcfg := &ProviderConf{
		Settings:   settings,
		sshClients: make([]*goph.Client, 0),
		mx:         &sync.Mutex{},
	}
	return pcfg
}

func (c *ProviderConf) AcquireSshClient() (client *goph.Client, err error) {
	c.mx.Lock()
	defer c.mx.Unlock()
	if len(c.sshClients) == 0 {
		client, err = GetSSHConnection(c.Settings)
		if err != nil {
			return nil, err
		}
	} else {
		client = c.sshClients[0]
		c.sshClients = c.sshClients[1:]
	}
	return client, nil
}

func (c *ProviderConf) ReleaseSshClient(client *goph.Client) {
	c.mx.Lock()
	defer c.mx.Unlock()
	c.sshClients = append(c.sshClients, client)
}
