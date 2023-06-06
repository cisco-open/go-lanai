package roles

import future.keywords

# Check permissions
has_permission(p) if {
	# Check permissions with Authentication from Input
    has_auth
	p in input.auth.permissions
} else {
	# Check permissions with JWT
	not has_auth
    some role in jwt_claims.roles
	p in data.role_permissions[role]
}

default jwt_claims := {}

jwt_claims := jwt if {
	token := trim_prefix(input.request.header.Authorization[_], "Bearer ")
	[_, jwt, _] := io.jwt.decode(token)
}

has_auth if {
	_ = input.auth
}