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
