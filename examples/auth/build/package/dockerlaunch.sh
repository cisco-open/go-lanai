#!/bin/bash

if [ -z "$POPULATE" ]; then
    exec ./example-auth-service "$@"
fi
