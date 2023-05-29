package apimanage

import future.keywords
import data.roles.has_permission

# API access Rules
default allow_api := false

allow_api if {
	input.request.method == "GET"
	input.request.path == "/apim/api/v1/scopes/list"
	has_permission("VIEW_APIKEY")
}

allow_api if {
	input.request.method == "GET"
	input.request.path == "/apim/api/v1/keys"
	has_permission("VIEW_APIKEY")
}

allow_api if {
	input.request.method == "POST"
	input.request.path == "/apim/api/v1/keys/generate"
	has_permission("MANAGE_APIKEY")
}

allow_api if {
	input.request.method == "DELETE"
	glob.match("/apim/api/v1/keys/*", ["/"], input.request.path)
	has_permission("MANAGE_APIKEY")
}

allow_api if {
	input.request.method == "GET"
	input.request.path == "/apim/api/v1/enforce/scopes/lookup"
	has_permission("IS_API_ADMIN")
}

allow_api if {
	input.request.method == "GET"
	input.request.path == "/apim/api/v1/enforce/mappings/lookup"
	has_permission("IS_API_ADMIN")
}

allow_api if {
	input.request.method == "GET"
	input.request.path == "/apim/api/v1/enforce/keys/verify"
	has_permission("IS_API_ADMIN")
}
