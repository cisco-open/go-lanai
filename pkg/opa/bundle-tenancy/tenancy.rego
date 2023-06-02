package tenancy

import future.keywords

default allow_resource := false
allow_resource if {
	input.resource.type == "poc"
	input.resource.tenant_id = input.auth.tenant_id
}