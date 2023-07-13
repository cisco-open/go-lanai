package ops

import future.keywords

# Check Operation
is(op) {
    op != ""
    input.resource.op == op
}

is(op) {
    op != ""
    input.resource.op == ""
}
