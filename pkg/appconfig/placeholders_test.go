// Copyright 2023 Cisco Systems, Inc. and its affiliates
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
//
// SPDX-License-Identifier: Apache-2.0

package appconfig

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/appconfig/parser"
	"fmt"
	. "github.com/onsi/gomega"
	"io"
	"os"
	"reflect"
	"testing"
	"time"
)

func TestResolvePlaceHolders(t *testing.T) {
	fullPath := "appconfig_test/placeholders_test.yml"
	file, e := os.Open(fullPath)
	if e != nil {
		t.Errorf("can't open test file")
	}
	defer func() {
		_ = file.Close()
	}()

	p := newTestProvider(fullPath, file)
	c := NewApplicationConfig(NewStaticProviderGroup(0, p))

	if e := c.Load(context.Background(), true); e != nil {
		t.Errorf("Load() returns error %v", e)
	}

	expectedKVs := map[string]interface{}{
		"application.profiles.additional": []string{},
		"a.a-1.a-1-1":                     "1_my_value",
		"a.a-1.a-1-2":                     "200s",
		"a.a-1.a-1-3":                     200.0,
		"b.b-1.b-1-1":                     "my_value",
		"b.b-1.b-1-2":                     200.0,
		"b.b-1.b-1-3":                     "default_value",
		"b.b-1.b-1-4":                     1000.0,
		"b.b-1.b-1-5":                     "${x.y.z}",
		"c.c-1.c-1-1":                     "my_value",
		"c.c-1.c-1-2":                     200.0,
		"d.d-1[0]":                        "1_my_value",
		"d.d-1[1]":                        "200s",
		"d.d-1[2]":                        200.0,
		"d.d-1[3]":                        "my_value",
		"d.d-1[4]":                        200.0,
		"d.d-2":                           "1_my_valuemy_value",
		"e.e-1[0]":                        "1_my_value",
		"e.e-1[1]":                        "200s",
		"e.e-1[2]":                        200.0,
		"e.e-1[3]":                        "my_value",
		"e.e-1[4]":                        200.0,
		"f.f-1":                           "1_my_value",
		"f.f-2":                           "200s",
		"f.f-3":                           200.0,
		"f.f-4":                           "1_default_value",
		"f.f-5":                           "1000s",
		"f.f-6":                           1000.0,
		"g.g-1":                           "default_value",
		"g.g-2":                           "1000s",
		"g.g-3":                           1000.0,
		"g.g-4":                           "default_value1000my_value",
		"h.h-1":                           "default_value1000my_value${x.y.z}",
		"h.h-2":                           "default_value1000my_value${b.b-1.b-1-5}",
		"i.i-1":                           "default_value",
		"i.i-2":                           "default_value",
		"i.i-3":                           "default_value",
	}

	actualKVs := make(map[string]interface{})
	gatherKvs := func(key string, value interface{}) error {
		actualKVs[key] = value
		return nil
	}

	if e := c.Each(gatherKvs); e != nil {
		t.Errorf("Each() returns error %v", e)
	}

	g := NewWithT(t)
	for k, v := range expectedKVs {
		g.Expect(actualKVs).To(HaveKeyWithValue(k, v), "actual KVs should have %s=%v", k, v)
	}

	if !reflect.DeepEqual(expectedKVs, actualKVs) {
		t.Errorf("resolved map not expected")
	}
}

func TestResolvePlaceHoldersWithCircularReference(t *testing.T) {
	// note, circular reference test may cause infinite loop
	timeout := 60 * time.Second
	if deadline, ok := t.Deadline(); ok {
		timeout = deadline.Sub(time.Now())
	}
	timer := time.After(timeout)
	done := make(chan bool)

	go func() {
		defer func() {
			done <- true
		}()
		doLoadConfigWithCircularReference(t)
		doTestResolvePlaceHoldersWithCircularReference(t)
	}()

	select {
	case <-timer:
		t.Error("test didn't finish in time")
	case <-done:
	}
}

func doLoadConfigWithCircularReference(t *testing.T) {
	fullPath := "appconfig_test/placeholders_circular_test.yml"
	file, e := os.Open(fullPath)
	if e != nil {
		t.Errorf("can't open test file")
	}
	defer func() {
		_ = file.Close()
	}()

	p := newTestProvider(fullPath, file)
	c := NewApplicationConfig(NewStaticProviderGroup(0, p))

	e = c.Load(context.Background(), true)

	if e == nil {
		t.Errorf("expected error")
	}

	t.Logf("got error as expected: %v", e)
}

// Note: "Load" function attempts to resolve placeholders in undefined order, and resolve function fail fast
// we need to make sure that attempting to resolve any circular reference would cause error, regardless of the resolving order
func doTestResolvePlaceHoldersWithCircularReference(t *testing.T) {
	fullPath := "appconfig_test/placeholders_circular_test.yml"
	file, e := os.Open(fullPath)
	if e != nil {
		t.Errorf("can't open test file")
	}
	defer func() {
		_ = file.Close()
	}()
	raw, e := loadYaml(file)
	if e != nil {
		t.Errorf("failed to load file: %v", e)
	}

	expectedBadKeys := []string{
		"a.b.c", "d.e.f", "g.h.i", "j.k.l",
	}
	count := 0
	for _, key := range expectedBadKeys {
		v := value(raw, key)
		if _, e := resolveValue(context.Background(), key, v, raw, nil); e != nil {
			count++
			t.Logf("got error as expected: %v", e)
		}
	}
	if count != len(expectedBadKeys) {
		t.Errorf("expected errors")
	}
}

func loadYaml(reader io.Reader) (map[string]interface{}, error) {

	encoded, e := io.ReadAll(reader)
	if e != nil {
		return nil, e
	}

	return parser.NewYamlPropertyParser()(encoded)
}

type testConfigProvider struct {
	ProviderMeta
	reader   io.Reader
	filePath string
}

func newTestProvider(filePath string, reader io.Reader) *testConfigProvider {
	return &testConfigProvider{
		ProviderMeta: ProviderMeta{Precedence: 0},
		reader:       reader,
		filePath:     filePath,
	}
}

func (p *testConfigProvider) Name() string {
	return fmt.Sprintf("file:%s", p.filePath)
}

func (p *testConfigProvider) Load(_ context.Context) (loadError error) {
	defer func() {
		if loadError != nil {
			p.Loaded = false
		} else {
			p.Loaded = true
		}
	}()

	p.Settings, loadError = loadYaml(p.reader)

	return nil
}
