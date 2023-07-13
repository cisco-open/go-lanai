package tenancy

import future.keywords

# Check Tenancy
allow_tenant_access {
    input.resource.tenant_id = input.auth.tenant_id
}

allow_tenant_access {
#    input.resource.tenant_path[_] = input.auth.accessible_tenants[_]
    input.resource.tenant_path[_] = input.auth.tenant_id
#    input.auth.tenant_id in input.resource.tenant_path
}

# Check updated tenancy (don't allow unless kept same
allow_change_tenant {
    not input.resource.delta.tenant_id
    not input.resource.delta.tenant_path
}

allow_change_tenant {
    input.resource.delta.tenant_id = input.resource.tenant_id
    input.resource.delta.tenant_path[_] = input.resource.tenant_id
}

