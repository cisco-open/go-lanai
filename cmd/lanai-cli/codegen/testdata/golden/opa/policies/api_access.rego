# METADATA
# title: testservice_api
# description: Checking if user is allowed to access API.
# authors:
#   - Your Name <your_email@email.com>
# custom:
#   short_description: user is allowed to access API
package testservice_api

import data.actuator.allow_endpoint
import data.actuator.allow_health_details

# TODO Update imports based on policies
#import data.rbac.has_permission
#import data.rbac.has_any_permission

# API access Rules

# METADATA
# description: GET /apim/api/v1/keys
allow_api {
	input.request.method == "GET"
	input.request.path == "/apim/api/v1/keys"
	# TODO Choose proper constraints
	#has_permission("PERMISSION")
	#has_any_permission({"PERMISSION1","PERMISSION2","PERMISSION3"})
}

# METADATA
# description: DELETE /my/api/v1/controllerResponsesTest
allow_api {
	input.request.method == "DELETE"
	input.request.path == "/my/api/v1/controllerResponsesTest"
	# TODO Choose proper constraints
	#has_permission("PERMISSION")
	#has_any_permission({"PERMISSION1","PERMISSION2","PERMISSION3"})
}

# METADATA
# description: GET /my/api/v1/controllerResponsesTest
allow_api {
	input.request.method == "GET"
	input.request.path == "/my/api/v1/controllerResponsesTest"
	# TODO Choose proper constraints
	#has_permission("PERMISSION")
	#has_any_permission({"PERMISSION1","PERMISSION2","PERMISSION3"})
}

# METADATA
# description: HEAD /my/api/v1/controllerResponsesTest
allow_api {
	input.request.method == "HEAD"
	input.request.path == "/my/api/v1/controllerResponsesTest"
	# TODO Choose proper constraints
	#has_permission("PERMISSION")
	#has_any_permission({"PERMISSION1","PERMISSION2","PERMISSION3"})
}

# METADATA
# description: PATCH /my/api/v1/controllerResponsesTest
allow_api {
	input.request.method == "PATCH"
	input.request.path == "/my/api/v1/controllerResponsesTest"
	# TODO Choose proper constraints
	#has_permission("PERMISSION")
	#has_any_permission({"PERMISSION1","PERMISSION2","PERMISSION3"})
}

# METADATA
# description: PUT /my/api/v1/controllerResponsesTest
allow_api {
	input.request.method == "PUT"
	input.request.path == "/my/api/v1/controllerResponsesTest"
	# TODO Choose proper constraints
	#has_permission("PERMISSION")
	#has_any_permission({"PERMISSION1","PERMISSION2","PERMISSION3"})
}

# METADATA
# description: TRACE /my/api/v1/controllerResponsesTest
allow_api {
	input.request.method == "TRACE"
	input.request.path == "/my/api/v1/controllerResponsesTest"
	# TODO Choose proper constraints
	#has_permission("PERMISSION")
	#has_any_permission({"PERMISSION1","PERMISSION2","PERMISSION3"})
}

# METADATA
# description: DELETE /my/api/v1/requestBodyTests/{id}
allow_api {
	input.request.method == "DELETE"
	glob.match("/my/api/v1/requestBodyTests/*", ["/"], input.request.path)
	# TODO Choose proper constraints
	#has_permission("PERMISSION")
	#has_any_permission({"PERMISSION1","PERMISSION2","PERMISSION3"})
}

# METADATA
# description: GET /my/api/v1/requestBodyTests/{id}
allow_api {
	input.request.method == "GET"
	glob.match("/my/api/v1/requestBodyTests/*", ["/"], input.request.path)
	# TODO Choose proper constraints
	#has_permission("PERMISSION")
	#has_any_permission({"PERMISSION1","PERMISSION2","PERMISSION3"})
}

# METADATA
# description: PATCH /my/api/v1/requestBodyTests/{id}
allow_api {
	input.request.method == "PATCH"
	glob.match("/my/api/v1/requestBodyTests/*", ["/"], input.request.path)
	# TODO Choose proper constraints
	#has_permission("PERMISSION")
	#has_any_permission({"PERMISSION1","PERMISSION2","PERMISSION3"})
}

# METADATA
# description: POST /my/api/v1/requestBodyTests/{id}
allow_api {
	input.request.method == "POST"
	glob.match("/my/api/v1/requestBodyTests/*", ["/"], input.request.path)
	# TODO Choose proper constraints
	#has_permission("PERMISSION")
	#has_any_permission({"PERMISSION1","PERMISSION2","PERMISSION3"})
}

# METADATA
# description: PUT /my/api/v1/requestBodyTests/{id}
allow_api {
	input.request.method == "PUT"
	glob.match("/my/api/v1/requestBodyTests/*", ["/"], input.request.path)
	# TODO Choose proper constraints
	#has_permission("PERMISSION")
	#has_any_permission({"PERMISSION1","PERMISSION2","PERMISSION3"})
}

# METADATA
# description: DELETE /my/api/v1/testpath/{scope}
allow_api {
	input.request.method == "DELETE"
	glob.match("/my/api/v1/testpath/*", ["/"], input.request.path)
	# TODO Choose proper constraints
	#has_permission("PERMISSION")
	#has_any_permission({"PERMISSION1","PERMISSION2","PERMISSION3"})
}

# METADATA
# description: GET /my/api/v1/testpath/{scope}
allow_api {
	input.request.method == "GET"
	glob.match("/my/api/v1/testpath/*", ["/"], input.request.path)
	# TODO Choose proper constraints
	#has_permission("PERMISSION")
	#has_any_permission({"PERMISSION1","PERMISSION2","PERMISSION3"})
}

# METADATA
# description: PATCH /my/api/v1/testpath/{scope}
allow_api {
	input.request.method == "PATCH"
	glob.match("/my/api/v1/testpath/*", ["/"], input.request.path)
	# TODO Choose proper constraints
	#has_permission("PERMISSION")
	#has_any_permission({"PERMISSION1","PERMISSION2","PERMISSION3"})
}

# METADATA
# description: POST /my/api/v1/testpath/{scope}
allow_api {
	input.request.method == "POST"
	glob.match("/my/api/v1/testpath/*", ["/"], input.request.path)
	# TODO Choose proper constraints
	#has_permission("PERMISSION")
	#has_any_permission({"PERMISSION1","PERMISSION2","PERMISSION3"})
}

# METADATA
# description: GET /my/api/v1/uuidtest/{id}
allow_api {
	input.request.method == "GET"
	glob.match("/my/api/v1/uuidtest/*", ["/"], input.request.path)
	# TODO Choose proper constraints
	#has_permission("PERMISSION")
	#has_any_permission({"PERMISSION1","PERMISSION2","PERMISSION3"})
}

# METADATA
# description: GET /my/api/v2/testArrayUUID
allow_api {
	input.request.method == "GET"
	input.request.path == "/my/api/v2/testArrayUUID"
	# TODO Choose proper constraints
	#has_permission("PERMISSION")
	#has_any_permission({"PERMISSION1","PERMISSION2","PERMISSION3"})
}

# METADATA
# description: PUT /my/api/v2/testArrayUUID
allow_api {
	input.request.method == "PUT"
	input.request.path == "/my/api/v2/testArrayUUID"
	# TODO Choose proper constraints
	#has_permission("PERMISSION")
	#has_any_permission({"PERMISSION1","PERMISSION2","PERMISSION3"})
}

# METADATA
# description: GET /my/api/v2/testpath
allow_api {
	input.request.method == "GET"
	input.request.path == "/my/api/v2/testpath"
	# TODO Choose proper constraints
	#has_permission("PERMISSION")
	#has_any_permission({"PERMISSION1","PERMISSION2","PERMISSION3"})
}

# METADATA
# description: GET /my/api/v3/anotherTest
allow_api {
	input.request.method == "GET"
	input.request.path == "/my/api/v3/anotherTest"
	# TODO Choose proper constraints
	#has_permission("PERMISSION")
	#has_any_permission({"PERMISSION1","PERMISSION2","PERMISSION3"})
}

# METADATA
# description: POST /my/api/v3/anotherTest
allow_api {
	input.request.method == "POST"
	input.request.path == "/my/api/v3/anotherTest"
	# TODO Choose proper constraints
	#has_permission("PERMISSION")
	#has_any_permission({"PERMISSION1","PERMISSION2","PERMISSION3"})
}

# METADATA
# description: POST /my/api/v4/testAPIThatDoesntUseAnyImports
allow_api {
	input.request.method == "POST"
	input.request.path == "/my/api/v4/testAPIThatDoesntUseAnyImports"
	# TODO Choose proper constraints
	#has_permission("PERMISSION")
	#has_any_permission({"PERMISSION1","PERMISSION2","PERMISSION3"})
}

# METADATA
# description: POST /my/api/v4/testRequestBodyUsingARefWithAllOf
allow_api {
	input.request.method == "POST"
	input.request.path == "/my/api/v4/testRequestBodyUsingARefWithAllOf"
	# TODO Choose proper constraints
	#has_permission("PERMISSION")
	#has_any_permission({"PERMISSION1","PERMISSION2","PERMISSION3"})
}
