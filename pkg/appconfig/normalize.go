package appconfig

import (
	"strings"
)

/*
Allow relaxed binding. The following should all bind to the same

acme.my-project.person.first-name
acme.myProject.person.firstName

Therefore, our algorithm is to remove the dash, and make everything lower case.

TODO: environment variables have ACME_MYPROJECT_PERSON_FIRSTNAME.
 This should be taken care of by the environment provider first to acme.myproject.person.firstname

*/
func NormalizeKey(key string) string {
	return strings.ToLower(
		strings.ReplaceAll(
			key,
			"-",
			""),
	)
}