# +entrypoint       has_scope
# +short            oauth2 client have given scope approved
# +desc             Checks if current oauth2 client have given scope approved
package oauth2

import data.oauth2.jwt_claims

# +desc verify scope from input.auth.client
has_scope(s) {
    has_client
	s = input.auth.client.scopes[_]
}

# +desc verify scope from JWT
has_scope(s) {
	not has_client
    s = jwt_claims.scope[_]
}

has_client {
	_ = input.auth.client
}