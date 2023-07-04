## go-msx support
npm package `@msx/nfv-swagger-ui`(under `./nodejs`) is also used by go-msx.
After updating the project, please follow the steps below to publish the npm package.

- update the `@msx/nfv-swagger-ui` package version in `./package.json`
- disable `CompressionPlugin` in `./webpack.prod.js` (go-msx doesn't support compressed js right now)
- `npm set registry "http://engci-maven.cisco.com/artifactory/api/npm/vms-npm/"`
- `npm build` 
- `npm publish`  
- update `nfv-swagger-ui` version in https://cto-github.cisco.com/NFV-BU/go-msx-build