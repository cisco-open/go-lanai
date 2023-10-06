# +short        current user can update resource's tenant
# +desc         Checks for resource's tenancy is changed and such action is allowed by current user
# +unknowns     input.resource.delta.tenant_id input.resource.delta.tenant_path input.resource.tenant_id # list of unknown inputs that this policy support
package tenancy

# Check updated tenancy (don't allow unless kept same)
# +desc tenancy not changed
allow_change_tenant {
    not input.resource.delta.tenant_id
    not input.resource.delta.tenant_path
}

# +desc tenancy is same as before
allow_change_tenant {
    input.resource.delta.tenant_id = input.resource.tenant_id
    input.resource.delta.tenant_path[_] = input.resource.tenant_id
}