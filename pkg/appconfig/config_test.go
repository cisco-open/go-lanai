package appconfig

import (
	"fmt"
	"strings"
	"testing"
)

func TestParsePlaceHolders(t *testing.T) {
	v := "${my.place.holder}"
	placeHolderKeys, isEmbedded, error := parsePlaceHolder(v)

	if len(placeHolderKeys) != 1 || strings.Compare(placeHolderKeys[0], "my.place.holder") != 0 || isEmbedded != false || error != nil {
		t.Errorf("%s not parsed correctly", v)
	}

	v = "1${my.place.holder}"
	placeHolderKeys, isEmbedded, error = parsePlaceHolder(v)

	if len(placeHolderKeys) != 1 || strings.Compare(placeHolderKeys[0], "my.place.holder") != 0 || isEmbedded != true || error != nil {
		t.Errorf("%s not parsed correctly", v)
	}

	v = "${my.place.holder}1"
	placeHolderKeys, isEmbedded, error = parsePlaceHolder(v)

	if len(placeHolderKeys) != 1 || strings.Compare(placeHolderKeys[0], "my.place.holder") != 0 || isEmbedded != true || error != nil {
		t.Errorf("%s not parsed correctly", v)
	}


	v = "${my.place.holder}${my.second.holder}"
	placeHolderKeys, isEmbedded, error = parsePlaceHolder(v)

	if len(placeHolderKeys) != 2 || isEmbedded != true || error != nil {
		t.Errorf("%s not parsed correctly", v)
	}

	if strings.Compare(placeHolderKeys[0], "my.place.holder") != 0 {
		t.Errorf("%s not parsed correctly", v)
	}

	if strings.Compare(placeHolderKeys[1], "my.second.holder") != 0 {
		t.Errorf("%s not parsed correctly", v)
	}

	v = "${${my.second.holder}}"
	placeHolderKeys, isEmbedded, error = parsePlaceHolder(v)
	if error == nil {
		t.Errorf("%s not parsed correctly. Expected error", v)
	}
	fmt.Println(error)
}
