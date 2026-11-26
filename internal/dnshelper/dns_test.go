// SPDX-License-Identifier: MIT

package dnshelper

import (
	"context"
	"reflect"
	"testing"
)

func TestUnmarshallRecord_RecordData(t *testing.T) {
	tests := []struct {
		name string
		json string
		want []string
	}{
		{
			name: "A record with typed IPv4Address property",
			json: `[{"HostName":"host","RecordType":"A","RecordData":{"CimInstanceProperties":[{"Name":"IPv4Address","Value":"192.168.0.1"}]}}]`,
			want: []string{"192.168.0.1"},
		},
		{
			name: "A record returned as generic hex Data property",
			json: `[{"HostName":"host","RecordType":"A","RecordData":{"CimInstanceProperties":[{"Name":"Data","Value":"C0A80001"}]}}]`,
			want: []string{"192.168.0.1"},
		},
		{
			name: "A record with lowercase hex",
			json: `[{"HostName":"host","RecordType":"A","RecordData":{"CimInstanceProperties":[{"Name":"Data","Value":"c0a800ea"}]}}]`,
			want: []string{"192.168.0.234"},
		},
		{
			name: "multiple A records, mixed formats",
			json: `[{"HostName":"host","RecordType":"A","RecordData":{"CimInstanceProperties":[{"Name":"Data","Value":"0A000001"}]}},` +
				`{"HostName":"host","RecordType":"A","RecordData":{"CimInstanceProperties":[{"Name":"IPv4Address","Value":"10.0.0.2"}]}}]`,
			want: []string{"10.0.0.1", "10.0.0.2"},
		},
		{
			name: "AAAA record returned as generic hex Data property",
			json: `[{"HostName":"host","RecordType":"AAAA","RecordData":{"CimInstanceProperties":[{"Name":"Data","Value":"20010DB8000000000000000000000001"}]}}]`,
			want: []string{"2001:db8::1"},
		},
		{
			name: "AAAA record with typed IPv6Address property",
			json: `[{"HostName":"host","RecordType":"AAAA","RecordData":{"CimInstanceProperties":[{"Name":"IPv6Address","Value":"2001:db8::1"}]}}]`,
			want: []string{"2001:db8::1"},
		},
		{
			name: "property selected by name, not position",
			json: `[{"HostName":"host","RecordType":"A","RecordData":{"CimInstanceProperties":[{"Name":"Other","Value":"x"},{"Name":"IPv4Address","Value":"10.0.0.3"}]}}]`,
			want: []string{"10.0.0.3"},
		},
		{
			name: "CNAME record unchanged",
			json: `[{"HostName":"alias","RecordType":"CNAME","RecordData":{"CimInstanceProperties":[{"Name":"HostNameAlias","Value":"target.example.com."}]}}]`,
			want: []string{"target.example.com."},
		},
		{
			name: "TXT record that looks like hex is not decoded",
			json: `[{"HostName":"txt","RecordType":"TXT","RecordData":{"CimInstanceProperties":[{"Name":"DescriptiveText","Value":"C0A80001"}]}}]`,
			want: []string{"C0A80001"},
		},
		{
			name: "missing Name falls back to first property",
			json: `[{"HostName":"host","RecordType":"A","RecordData":{"CimInstanceProperties":[{"Value":"10.0.0.4"}]}}]`,
			want: []string{"10.0.0.4"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := unmarshallRecord(context.Background(), []byte(tt.json))
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			if !reflect.DeepEqual(got.Records, tt.want) {
				t.Errorf("got %v, want %v", got.Records, tt.want)
			}
		})
	}
}

func TestUnmarshallRecord_NoRecordData(t *testing.T) {
	json := `[{"HostName":"host","RecordType":"A","RecordData":{"CimInstanceProperties":[]}}]`
	if _, err := unmarshallRecord(context.Background(), []byte(json)); err == nil {
		t.Fatal("expected an error for empty record data, got nil")
	}
}
