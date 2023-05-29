package apimanage

import future.keywords
import data.roles.has_permission

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
	has_permission(perm)
}
