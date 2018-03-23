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

// +build !race

package cache

import (
	"testing"
	"time"
)

func TestTTLBasic(t *testing.T) {
	ttl := NewTTL(5*time.Second, 1*time.Millisecond)
	testCacheBasic(ttl, t)
}

func TestTTLConcurrent(t *testing.T) {
	ttl := NewTTL(5*time.Second, 1*time.Second)
	testCacheConcurrent(ttl, t)
}

func TestTTLExpiration(t *testing.T) {
	ttl := NewTTL(5*time.Second, 100*time.Second).(*ttlWrapper)
	testCacheExpiration(ttl, ttl.evictExpired, t)
}

func TestTTLEvicter(t *testing.T) {
	ttl := NewTTL(5*time.Second, 1*time.Millisecond)
	testCacheEvicter(ttl, t)
}

func TestTTLEvictExpired(t *testing.T) {
	ttl := NewTTL(5*time.Second, 0)
	testCacheEvictExpired(ttl, t)
}

func TestTTLFinalizer(t *testing.T) {
	c := NewTTL(5*time.Second, 1*time.Millisecond).(*ttlWrapper)
	gate := &c.evicterTerminated
	testCacheFinalizer(gate, t)
}

func BenchmarkTTLGet(b *testing.B) {
	c := NewTTL(5*time.Minute, 1*time.Minute)
	benchmarkCacheGet(c, b)
}

func BenchmarkTTLGetConcurrent(b *testing.B) {
	c := NewTTL(5*time.Minute, 1*time.Minute)
	benchmarkCacheGetConcurrent(c, b)
}

func BenchmarkTTLSet(b *testing.B) {
	c := NewTTL(5*time.Minute, 1*time.Minute)
	benchmarkCacheSet(c, b)
}

func BenchmarkTTLSetConcurrent(b *testing.B) {
	c := NewTTL(5*time.Minute, 1*time.Minute)
	benchmarkCacheSetConcurrent(c, b)
}

func BenchmarkTTLGetSetConcurrent(b *testing.B) {
	c := NewTTL(5*time.Minute, 1*time.Minute)
	benchmarkCacheGetSetConcurrent(c, b)
}

func BenchmarkTTLSetRemove(b *testing.B) {
	c := NewTTL(5*time.Minute, 1*time.Minute)
	benchmarkCacheSetRemove(c, b)
}
