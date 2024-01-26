// SPDX-License-Identifier: MIT

package provider

import "testing"

func Test_suppressRecordDiffForType(t *testing.T) {
	tests := []struct {
		name       string
		rrType     string
		oldRecords []string
		newRecords []string
		want       bool
	}{
		// rrType AAAA test cases
		{
			"test-uppercase-ipv6", "AAAA", []string{"2001:DB8::1"}, []string{"2001:db8::1"}, true,
		},
		{
			"test-lowercase-ipv6", "AAAA", []string{"2001:db8::1"}, []string{"2001:db8::1"}, true,
		},
		{
			"test-multiple-unsorted-casemix-ipv6", "AAAA", []string{"2001:DB8::1", "2001:db8::2"}, []string{"2001:db8::2", "2001:db8::1"}, true,
		},
		{
			"test-multiple-lengths-1", "AAAA", []string{"2001:DB8::1"}, []string{"2001:db8::2", "2001:db8::1"}, false,
		},
		{
			"test-multiple-lengths-2", "AAAA", []string{}, []string{"2001:db8::2", "2001:db8::1"}, false,
		},
		{
			"test-empty", "AAAA", []string{}, []string{}, true,
		},
		// rrType A test cases
		{
			"test-empty", "A", []string{}, []string{}, true,
		},
		{
			"test-ipv4", "A", []string{"203.0.113.11"}, []string{"203.0.113.11"}, true,
		},
		{
			"test-empty-ivp4", "A", []string{""}, []string{"203.0.113.11"}, false,
		},
		// rrType CNAME test cases
		{
			"test-dot-cname", "CNAME", []string{"example-host.example.com."}, []string{"example-host.example.com"}, true,
		},
		// rrType PTR test cases
		{
			"test-dot-ptr", "PTR", []string{"example-host.example.com."}, []string{"example-host.example.com"}, true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := suppressRecordDiffForType(tt.oldRecords, tt.newRecords, tt.rrType); got != tt.want {
				t.Errorf("suppressRecordDiffForType() = %v, want %v", got, tt.want)
			}
		})
	}
}
