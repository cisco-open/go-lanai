package roles

import future.keywords

# Check permissions
has_permission(p) {
	# Check permissions with Authentication from Input
    has_auth
	p in input.auth.permissions
}

has_permission(p) {
	# Check permissions with JWT
	not has_auth
    some role in jwt_claims.roles
	p in data.role_permissions[role]
}

jwt_claims := jwt {
	token := trim_prefix(input.request.header.Authorization[_], "Bearer ")
	[_, jwt, _] := io.jwt.decode(token)
}

has_auth {
	_ = input.auth
}