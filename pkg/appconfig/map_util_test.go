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
	"strings"
	"testing"
)

func TestNormalizeKey(t *testing.T) {
	key := "acme.my-project.person.first-name"
	expected := "acme.my-project.person.first-name"

	actual := NormalizeKey(key)

	if strings.Compare(expected, actual) != 0 {
		t.Errorf("expected %s, got %s", expected, actual)
	}

	key = "acme.myProject.person.firstName"
	actual = NormalizeKey(key)

	if strings.Compare(expected, actual) != 0 {
		t.Errorf("expected %s, got %s", expected, actual)
	}

	key = "Acme.MyProject.Person.FirstName"
	actual = NormalizeKey(key)

	if strings.Compare(expected, actual) != 0 {
		t.Errorf("expected %s, got %s", expected, actual)
	}

	key = "AcmE.MyProjecT.PersoN.FirstNamE"
	expected = "acm-e.my-projec-t.perso-n.first-nam-e"
	actual = NormalizeKey(key)

	if strings.Compare(expected, actual) != 0 {
		t.Errorf("expected %s, got %s", expected, actual)
	}

	key = "ACME.MYPROJECT.PERSON.FIRSTNAME"
	expected = "acme.myproject.person.firstname"
	actual = NormalizeKey(key)

	if strings.Compare(expected, actual) != 0 {
		t.Errorf("expected %s, got %s", expected, actual)
	}

}
