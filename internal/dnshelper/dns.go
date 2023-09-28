package dnshelper

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/nrkno/terraform-provider-windns/internal/config"
)

const (
	IDSeparator = "_"

	RecordTypeAAAA  = "AAAA"
	RecordTypeA     = "A"
	RecordTypeTXT   = "TXT"
	RecordTypePTR   = "PTR"
	RecordTypeCNAME = "CNAME"
)

type Record struct {
	ZoneName   string   `json:"ZoneName"`
	HostName   string   `json:"HostName"`
	RecordType string   `json:"RecordType"`
	Records    []string `json:"Records"`
	CreatePtr  bool     `json:"CreatePtr"`
}

type DNSRecord struct {
	HostName   string     `json:"HostName"`
	RecordType string     `json:"RecordType"`
	DN         string     `json:"DistinguishedName"`
	RecordData RecordData `json:"RecordData"`
	TimeToLive TTL        `json:"TimeToLive"`
}

// The structure we get from powershell contains more fields, but we're only interested in CimInstanceProperties.
type RecordData struct {
	CimInstanceProperties []CimInstanceProperties `json:"CimInstanceProperties"`
}

// The structure we get from powershell contains more fields, but we're only interested in the Value.
type CimInstanceProperties struct {
	Value string `json:"value"`
}

// The structure we get from powershell contains more fields, but we're only interested in TotalSeconds.
type TTL struct {
	TotalSeconds int64 `json:"TotalSeconds"`
}

// windns has no concept of primary key so we need to create one based on inputs
func (r *Record) Id() string {
	return strings.Join([]string{r.HostName, r.ZoneName, r.RecordType}, IDSeparator)
}

// NewDNSRecordFromResource returns a new Record struct populated from resource data
func NewDNSRecordFromResource(d *schema.ResourceData) *Record {
	var records []string
	recordsSet := d.Get("records").(*schema.Set)

	for _, v := range recordsSet.List() {
		records = append(records, SanitiseString(v.(string)))
	}

	return &Record{
		ZoneName:   SanitiseTFInput(d, "zone_name"),
		HostName:   SanitiseTFInput(d, "name"),
		RecordType: SanitiseTFInput(d, "type"),
		CreatePtr:  d.Get("create_ptr").(bool),
		//		TTL:        d.Get("ttl").(int64),
		Records: records,
	}
}

func GetDNSRecordFromId(ctx context.Context, conf *config.ProviderConf, id string) (*Record, error) {
	idComponents := strings.Split(id, IDSeparator)
	hostName := idComponents[0]
	zoneName := idComponents[1]
	recordType := idComponents[2]

	cmd := fmt.Sprintf("Get-DnsServerResourceRecord -ZoneName %s -Name %s -RRType %s", zoneName, hostName, recordType)

	conn, err := conf.AcquireSshClient()
	if err != nil {
		return nil, fmt.Errorf("while acquiring ssh client: %s", err)
	}
	defer conf.ReleaseSshClient(conn)

	psOpts := CreatePSCommandOpts{
		JSONOutput: true,
		JSONDepth:  4,
		ForceArray: true,
		Username:   conf.Settings.SshUsername,
		Password:   conf.Settings.SshPassword,
		Server:     conf.Settings.DnsServer,
	}
	psCmd := NewPSCommand([]string{cmd}, psOpts)

	result, err := psCmd.Run(conf)
	if err != nil {
		return nil, fmt.Errorf("ssh execution failure in GetDNSRecordFromId: %s", err)
	}

	if result.ExitCode != 0 {
		return nil, fmt.Errorf("Get-DnsServerResourceRecord exited with a non zero exit code (%d), stderr: %s", result.ExitCode, result.StdErr)
	}

	record, err := unmarshallRecord(ctx, []byte(result.Stdout))
	if err != nil {
		return nil, fmt.Errorf("GetDNSRecordFromId: %s", err)
	}

	record.ZoneName = zoneName
	return record, nil
}

// Create creates a new DNSRecord object in DNS server
func (r *Record) Create(conf *config.ProviderConf) (string, error) {
	if r.ZoneName == "" {
		return "", fmt.Errorf("DNSRecord.Create: missing zone_name variable")
	}

	if r.HostName == "" {
		return "", fmt.Errorf("DNSRecord.Create: missing name variable")
	}

	if r.RecordType == "" {
		return "", fmt.Errorf("DNSRecord.Create: missing type variable")
	}

	if len(r.Records) == 0 {
		return "", fmt.Errorf("DNSRecord.Create: missing record variable")
	}

	for _, recordData := range r.Records {
		err := r.addRecordData(conf, recordData)
		if err != nil {
			return "", err
		}
	}

	// We don't get any unique ID from the create command, so we assume id is a composite of input variables.
	return r.Id(), nil
}

// Update updates an existing DNSRecord object in DNS server
func (r *Record) Update(ctx context.Context, conf *config.ProviderConf, changes map[string]interface{}) error {
	existing, err := GetDNSRecordFromId(ctx, conf, r.Id())
	if err != nil {
		return err
	}

	var records []string
	expectedRecords := changes["records"].(*schema.Set)

	for _, v := range expectedRecords.List() {
		records = append(records, SanitiseString(v.(string)))
	}

	toAdd, toRemove := diffRecordLists(records, existing.Records)
	for _, recordData := range toAdd {
		err = r.addRecordData(conf, recordData)
		if err != nil {
			return err
		}
	}

	for _, recordData := range toRemove {
		err = r.removeRecordData(conf, recordData)
		if err != nil {
			return err
		}
	}
	return nil
}

// Delete deletes an existing DNSRecord object in DNS server
func (r *Record) Delete(conf *config.ProviderConf) error {
	for _, recordData := range r.Records {
		err := r.removeRecordData(conf, recordData)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *Record) addRecordData(conf *config.ProviderConf, recordData string) error {
	cmd := fmt.Sprintf("Add-DNSServerResourceRecord -ZoneName %s -name %s -%s", r.ZoneName, r.HostName, r.RecordType)

	if r.RecordType == RecordTypeA {
		cmd = fmt.Sprintf("%s -IPv4Address %s", cmd, recordData)
	} else if r.RecordType == RecordTypeAAAA {
		cmd = fmt.Sprintf("%s -IPv6Address %s", cmd, strings.ToLower(recordData))
	} else if r.RecordType == RecordTypeTXT {
		cmd = fmt.Sprintf("%s -DescriptiveText %s", cmd, recordData)
	} else if r.RecordType == RecordTypePTR {
		cmd = fmt.Sprintf("%s -PtrDomainName %s", cmd, recordData)
	} else if r.RecordType == RecordTypeCNAME {
		cmd = fmt.Sprintf("%s -HostNameAlias %s", cmd, recordData)
	} else {
		return fmt.Errorf("record type %s is not supported", r.RecordType)
	}

	if (r.RecordType == RecordTypeA || r.RecordType == RecordTypeAAAA) && r.CreatePtr {
		cmd = fmt.Sprintf("%s -CreatePtr", cmd)
	}

	psOpts := CreatePSCommandOpts{
		JSONOutput: false,
		ForceArray: false,
		Username:   conf.Settings.SshUsername,
		Password:   conf.Settings.SshPassword,
		Server:     conf.Settings.DnsServer,
	}
	psCmd := NewPSCommand([]string{cmd}, psOpts)

	result, err := psCmd.Run(conf)
	if err != nil {
		return fmt.Errorf("ssh execution failure while creating a DNS object: %s", err)
	}

	if result.ExitCode != 0 {
		fmt.Printf("CMD: %s", psCmd.String())
		return fmt.Errorf("Add-DnsServerResourceRecord exited with a non zero exit code (%d), stderr: %s", result.ExitCode, result.StdErr)
	}
	return nil
}

func (r *Record) removeRecordData(conf *config.ProviderConf, recordData string) error {
	cmd := fmt.Sprintf("Remove-DnsServerResourceRecord -Force -ZoneName %s -RRType %s -Name %s -RecordData %s", r.ZoneName, r.RecordType, r.HostName, recordData)
	conn, err := conf.AcquireSshClient()
	if err != nil {
		return fmt.Errorf("while acquiring ssh client: %s", err)
	}
	defer conf.ReleaseSshClient(conn)
	psOpts := CreatePSCommandOpts{
		JSONOutput: false,
		ForceArray: false,
		Username:   conf.Settings.SshUsername,
		Password:   conf.Settings.SshPassword,
		Server:     conf.Settings.DnsServer,
	}

	psCmd := NewPSCommand([]string{cmd}, psOpts)

	result, err := psCmd.Run(conf)
	if err != nil {
		return fmt.Errorf("ssh execution failure while removing record object: %s", err)
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("Remove-DnsServerResourceRecord exited with a non zero exit code (%d), stderr: %s", result.ExitCode, result.StdErr)
	}
	return nil
}

// handle if powershell returns single object or list of objects.
func unmarshallRecord(ctx context.Context, input []byte) (*Record, error) {
	var err error
	var records []DNSRecord

	t := bytes.TrimSpace(input)
	if len(t) == 0 {
		return nil, fmt.Errorf("empty json document")
	}

	err = json.Unmarshal(input, &records)
	if err != nil {
		tflog.Debug(ctx, fmt.Sprintf("Failed to unmarshall an DNSRecord json document with error %q, document was %s", err, string(input)))
		return nil, fmt.Errorf("failed while unmarshalling DNSRecord json document: %s", err)
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("invalid data while unmarshalling DNSRecord data, json doc was: %s", string(input))
	}

	var rs []string
	for _, v := range records {
		recordData := v.RecordData.CimInstanceProperties[0].Value
		rs = append(rs, recordData)
	}

	record := Record{
		HostName:   records[0].HostName,
		RecordType: records[0].RecordType,
		//		TTL:        records[0].TimeToLive.TotalSeconds,
		Records: rs,
	}

	return &record, nil
}

func recordExistsInList(r string, list []string) bool {
	for _, item := range list {
		if r == item {
			return true
		}
	}
	return false
}

func diffRecordLists(expectedRecords, existingRecords []string) ([]string, []string) {
	var toAdd, toRemove []string

	for _, record := range expectedRecords {
		if !recordExistsInList(record, existingRecords) {
			toAdd = append(toAdd, record)
		}
	}

	for _, record := range existingRecords {
		if !recordExistsInList(record, expectedRecords) {
			toRemove = append(toRemove, record)
		}
	}
	return toAdd, toRemove
}
