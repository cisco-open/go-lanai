package envprovider

import (
	"os"
	"strings"
	"testing"
)

func TestLoadEnvironmentProperties(t *testing.T) {
	os.Setenv("MY_PROJECT_NAME", "test_project")
	os.Setenv("MY_PROJECT_AUTHOR", "test_author")

	p := NewEnvProvider(0)
	p.Load()

	actual := p.GetSettings()

	my := actual["MY"].(map[string]interface{})
	project := my["PROJECT"].(map[string]interface{})

	if strings.Compare(project["NAME"].(string), "test_project") != 0 {
		t.Errorf("expected MY_PROJECT_NAME with value %s, but got %s", "test_project", project["NAME"].(string))
	}

	if strings.Compare(project["AUTHOR"].(string), "test_author") != 0 {
		t.Errorf("expected MY_PROJECT_NAME with value %s, but got %s", "test_author", project["AUTHOR"].(string))
	}
}