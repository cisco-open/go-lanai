package ownership

import future.keywords

# Check shared status
is_shared(op) if {
     op = input.resource.share[input.auth.user_id][_]
}