# +short current user has any given permission
# +desc Checks if current user has given string array of permissions
package rbac

import future.keywords
import data.oauth2.jwt_claims

# +desc Check permissions with Auth from input
# +entrypoint
has_any_permission(p) {
    has_auth
    # iterate through both list and set and return intersection
    result := {x | x = input.auth.permissions[_]; x = p[_]}
    count(result) > 0
}


# +desc check permissions with JWT
has_any_permission(p) {
    not has_auth
    some role in jwt_claims.roles
    # iterate through both list and set and return intersection
    result := {x | x = data.role_permissions[role][_]; x = p[_]}
    count(result) > 0
}