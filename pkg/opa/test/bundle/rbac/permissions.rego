# +entrypoint   has_permission    # name of the policy to be queried
# +short        current user have given permission
# +desc         Checks if current user have given permission
package rbac

import future.keywords
import data.oauth2.jwt_claims

# +desc check permissions with Authentication from Input
has_permission(p) {
	# Check permissions with Authentication from Input
    has_auth
	p = input.auth.permissions[_]
}

# +desc check permissions with JWT
has_permission(p) {
	not has_auth
    some role in jwt_claims.roles
	p = data.role_permissions[role][_]
}

has_auth {
	_ = input.auth
}