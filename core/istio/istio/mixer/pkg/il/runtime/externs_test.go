// Copyright 2017 Istio Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package runtime

import (
	"bytes"
	"net"
	"testing"
	"time"
)

func TestExternIp(t *testing.T) {
	b, err := externIP("1.2.3.4")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !bytes.Equal(b, net.ParseIP("1.2.3.4")) {
		t.Fatalf("Unexpected output: %v", b)
	}
}

func TestExternIp_Error(t *testing.T) {
	_, err := externIP("A.A.A.A")
	if err == nil {
		t.Fatalf("Expected error not found.")
	}
}

func TestExternIpEqual_True(t *testing.T) {
	b := externIPEqual(net.ParseIP("1.2.3.4"), net.ParseIP("1.2.3.4"))
	if !b {
		t.Fatal()
	}
}

func TestExternIpEqual_False(t *testing.T) {
	b := externIPEqual(net.ParseIP("1.2.3.4"), net.ParseIP("1.2.3.5"))
	if b {
		t.Fatal()
	}
}

func TestExternTimestamp(t *testing.T) {
	ti, err := externTimestamp("2015-01-02T15:04:35Z")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if ti.Year() != 2015 || ti.Month() != time.January || ti.Day() != 2 || ti.Hour() != 15 || ti.Minute() != 4 {
		t.Fatalf("Unexpected time: %v", ti)
	}
}

func TestExternTimestamp_Error(t *testing.T) {
	_, err := externTimestamp("AAA")
	if err == nil {
		t.Fatalf("Expected error not found.")
	}
}

func TestExternTimestampEqual_True(t *testing.T) {
	t1, _ := externTimestamp("2015-01-02T15:04:35Z")
	t2, _ := externTimestamp("2015-01-02T15:04:35Z")
	b := externTimestampEqual(t1, t2)
	if !b {
		t.Fatal()
	}
}

func TestExternTimestampEqual_False(t *testing.T) {
	t1, _ := externTimestamp("2015-01-02T15:04:35Z")
	t2, _ := externTimestamp("2018-11-11T15:04:35Z")
	b := externTimestampEqual(t1, t2)
	if b {
		t.Fatal()
	}
}

func TestExternMatch(t *testing.T) {
	var cases = []struct {
		s string
		p string
		e bool
	}{
		{"ns1.svc.local", "ns1.*", true},
		{"ns1.svc.local", "ns2.*", false},
		{"svc1.ns1.cluster", "*.ns1.cluster", true},
		{"svc1.ns1.cluster", "*.ns1.cluster1", false},
	}

	for _, c := range cases {
		if externMatch(c.s, c.p) != c.e {
			t.Fatalf("externMatch failure: %+v", c)
		}
	}
}

func TestExternMatches(t *testing.T) {
	var cases = []struct {
		p string
		s string
		e bool
	}{
		{"ns1\\.svc\\.local", "ns1.svc.local", true},
		{"ns1.*", "ns1.svc.local", true},
		{"ns2.*", "ns1.svc.local", false},
	}

	for _, c := range cases {
		m, err := externMatches(c.p, c.s)
		if err != nil {
			t.Fatalf("Unexpected error: %+v, %v", c, err)
			if m != c.e {
				t.Fatalf("matches failure: %+v", c)
			}
		}
	}
}

func TestExternStartsWith(t *testing.T) {
	var cases = []struct {
		s string
		p string
		e bool
	}{
		{"abc", "a", true},
		{"abc", "", true},
		{"abc", "abc", true},
		{"abc", "abcd", false},
		{"cba", "a", false},
	}

	for _, c := range cases {
		m := externStartsWith(c.s, c.p)
		if m != c.e {
			t.Fatalf("startsWith failure: %+v", c)
		}
	}
}

func TestExternEndsWith(t *testing.T) {
	var cases = []struct {
		s string
		u string
		e bool
	}{
		{"abc", "c", true},
		{"abc", "", true},
		{"abc", "abc", true},
		{"abc", "dabc", false},
		{"cba", "c", false},
	}

	for _, c := range cases {
		m := externEndsWith(c.s, c.u)
		if m != c.e {
			t.Fatalf("endsWith failure: %+v", c)
		}
	}
}

func TestExternEmptyStringMap(t *testing.T) {
	m := externEmptyStringMap()
	if len(m) != 0 {
		t.Errorf("emptyStringMap() returned non-empty map: %#v", m)
	}
}
