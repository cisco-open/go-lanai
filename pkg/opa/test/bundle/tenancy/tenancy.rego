# +short        current user can access resource's tenant
# +desc         Checks for resource's tenancy against current user
# +unknowns     input.resource.tenant_id input.resource.tenant_path  # list of unknown inputs that this policy support
package tenancy

# Check Tenancy
# +desc user and resource belong to same tenant
allow_tenant_access {
    input.resource.tenant_id = input.auth.tenant_id
}

# +desc resource belong to child tenant of user's tenant
allow_tenant_access {
#    input.resource.tenant_path[_] = input.auth.accessible_tenants[_]
    input.resource.tenant_path[_] = input.auth.tenant_id
}


