package apimanage

import future.keywords

default allow_api := false
allow_api if {
	input.request.method == "GET"
	input.request.path == "/apim/api/v1/scopes/list"
	input.auth.userAuth.permissions["VIEW_APIKEY"]
}

allow_api if {
	input.request.method == "GET"
	input.request.path == "/apim/api/v1/keys"
	input.auth.userAuth.permissions["VIEW_APIKEY"]
}

allow_api if {
	input.request.method == "POST"
	input.request.path == "/apim/api/v1/keys/generate"
	input.auth.userAuth.permissions["MANAGE_APIKEY"]
}

allow_api if {
	input.request.method == "DELETE"
	glob.match("/apim/api/v1/keys/*", ["/"], input.request.path)
	input.auth.userAuth.permissions["MANAGE_APIKEY"]
}

allow_api if {
	input.request.method == "GET"
	input.request.path == "/apim/api/v1/enforce/scopes/lookup"
	input.auth.userAuth.permissions["IS_API_ADMIN"]
}

allow_api if {
	input.request.method == "GET"
	input.request.path == "/apim/api/v1/enforce/mappings/lookup"
	input.auth.userAuth.permissions["IS_API_ADMIN"]
}

allow_api if {
	input.request.method == "GET"
	input.request.path == "/apim/api/v1/enforce/keys/verify"
	input.auth.userAuth.permissions["IS_API_ADMIN"]
}
