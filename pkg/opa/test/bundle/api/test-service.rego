package testservice

import future.keywords
import data.rbac.has_permission

# API access Rules

# /test/api/get
allow_api if {
	input.request.method == "GET"
	input.request.path == "/test/api/get"
	has_permission("VIEW")
}

allow_api if {
	input.request.method == "GET"
	input.request.path == "/test/api/get"
	has_permission("ADMIN")
}

# /test/api/post
allow_api if {
	input.request.method == "POST"
	input.request.path == "/test/api/post"
	has_permission("MANAGE")
}

allow_api if {
	input.request.method == "POST"
	input.request.path == "/test/api/post"
	has_permission("ADMIN")
}