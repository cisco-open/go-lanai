package apimanage

import future.keywords

# API permissions lookup from Data
default required_permissions := []
required_permissions := permissions {
	some ctxPath, apis in data.api_access
    glob.match(concat("", [ctxPath, "/**"]), ["/"], input.request.path)
    some path, perms in apis
    glob.match(concat("", [ctxPath, path]), ["/"], input.request.path)
	permissions := perms[input.request.method]
}

# API access with authentication from input
default allow_api := false
allow_api if {
	some perm in required_permissions
	input.auth.userAuth.permissions[perm]
}

# API access with JWT from request
# TBD