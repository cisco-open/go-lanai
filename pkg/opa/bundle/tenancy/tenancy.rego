package tenancy

import future.keywords

# Check Tenancy
default allow_tenant_access := false
allow_tenant_access if {
	input.resource.tenant_id = input.auth.tenant_id
}

allow_tenant_access if {
	input.resource.tenant_id in input.auth.tenant_path
}
