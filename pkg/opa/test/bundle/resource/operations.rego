# +short        resource operation is given value
# +desc         check if resource operation is given value
package resource

# Check Operation
is_op(op) {
    op != ""
    input.resource.op == op
}

is_op(op) {
    op != ""
    input.resource.op == ""
}
