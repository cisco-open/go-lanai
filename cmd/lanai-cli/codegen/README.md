# Golanai Code Generator

Run with `./lanai-cli codegen`

## Using the Filesystem

To manipulate where the files are generated to, look at the `templates` directory:
```
templates/
    common/
    srcRoot/
```

### common
Templates inside `common` can be imported into any other templates in use with a call of `template`, e.g if you have a common.tmpl like
```
Hello from common
```
Then, from any other template:
```
Hello!
{{ template "common.tmpl" }}
```

Will output

```
Hello!
Hello from common
```

### srcRoot

Files generated from templates inside `srcRoot` will be output to a similar hierarchy to where the template is located.
e.g. If you use `srcRoot/inner/inner.tmpl`, then your generated file will be sent to `<output-dir>/inner/inner.go`

If a folder has a name surrounded by `@`, e.g `@VERSION@`, then that value can be replaced programatically.
`GetOutputFilePath` takes a modifier map that has the special name, and what want to resolve to.

For example:
```go
//Say we want to make a file from `srcRoot/@VERSION@/version.tmpl`
modifiers := map[string]string{"VERSION": "v2",}
fsm.GetOutputFilePath("version.tmpl", "version.go", modifiers) // Will generate to <output-dir>/v2/version.tmpl
```
