package test

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/appconfig"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/appconfig/fileprovider"
	"fmt"
	"os"
	"reflect"
	"testing"
)



func TestResolvePlaceHolders(t *testing.T) {
	fullPath := "test_placeholders.yml"
	file, e := os.Open(fullPath);

	if e != nil {
		t.Errorf("can't open test file")
	}

	p := fileprovider.NewProvider("test", 0, fullPath, file)
	c := appconfig.NewApplicationConfig(p)

	c.Load(true)

	expectedKVs := map[string]interface{} {
		"a.b.c": "1_my_value",
		"d.e.f": "my_value",
		"g.h.i": "my_value",
		"j.k[0]" : "1_my_value",
		"j.k[1]" : "my_value",
		"l" : "1_my_valuemy_value",
		"m[0]" : "1_my_value",
		"m[1]" : "my_value",
	}

	actualKVs := make(map[string]interface{})
	gatherKvs := func(key string, value interface{})error {
		actualKVs[key]=value
		return nil
	}

	c.Each(gatherKvs)

	if !reflect.DeepEqual(expectedKVs, actualKVs) {
		t.Errorf("resolved map not expected")
	}
}

func TestResolvePlaceHoldersWithCircularReference(t *testing.T) {
	fullPath := "test_placeholders_circular.yml"
	file, error := os.Open(fullPath);

	if error != nil {
		t.Errorf("can't open test file")
	}

	p := fileprovider.NewProvider("test", 0, fullPath, file)
	c := appconfig.NewApplicationConfig(p)

	error = c.Load(true)

	if error == nil {
		t.Errorf("expected error")
	}

	fmt.Println(error)

}