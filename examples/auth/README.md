1. create module.yml
2. add go.mod, add go-lanai requirement, go mod download to generate go.sum 
3. add Makefile (for the above steps see develop.md)
4. make init CLI_TAG="develop" (make init should ignore no go.sum error)
5. we didn't end up using lanai-cli codegen -o ./ (TODO, it should support no contract)
6. main file
7. serviceinit
8. bootstrap.yml, application.yml - don't have to do this if code gen could work without contract
9. also need the key files