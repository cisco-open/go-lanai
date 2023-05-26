package apimanage

import future.keywords

default required_permissions := []
required_permissions := permissions {
    some path, methods in data.api_access[_]
    path == input.request.path
	permissions := methods[input.request.method]
}

default allow_api := false
allow_api if {
	some perm in required_permissions
	input.auth.userAuth.permissions[perm]
}