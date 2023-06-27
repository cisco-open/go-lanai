package ops

import future.keywords

# Check Operation
is(op) if {
    input.resource.op == op
}

is(op) if {
    op != ""
    input.resource.op == ""
}
