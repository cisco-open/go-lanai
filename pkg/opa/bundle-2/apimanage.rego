package apimanage

import future.keywords

default required_permissions := []
required_permissions := permissions {
	some ctxPath, apis in data.api_access
    glob.match(concat("", [ctxPath, "/**"]), ["/"], input.request.path)
    some path, perms in apis
    glob.match(concat("", [ctxPath, path]), ["/"], input.request.path)
	permissions := perms[input.request.method]
}

default allow_api := false
allow_api if {
	some perm in required_permissions
	input.auth.userAuth.permissions[perm]
}
