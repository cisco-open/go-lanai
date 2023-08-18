# +short        resource type is given value
# +desc         check if resource type is given value
package resource

is_type(t) {
    t != ""
    input.resource.type == t
}