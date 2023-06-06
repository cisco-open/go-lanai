package ownership

import future.keywords

# Check shared status
default op := ""
is_shared(op) if {
    op in input.resource.share[input.auth.user_id]
}