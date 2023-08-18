# +name         jwt_claims
# +short        extract jwt claims from request
# +desc         Extract jwt claims from request. Returns a object instead of boolean
package oauth2

jwt_claims := jwt {
	token := trim_prefix(input.request.header.Authorization[_], "Bearer ")
	[_, jwt, _] := io.jwt.decode(token)
}