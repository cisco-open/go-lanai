# SAML SSO Module
This module allows a service to act as IDP (allow others to SSO with the service).

This module registers a feature configurer which does the following:

1. add metadata refresh middleware to the sso endpoint
2. add sso endpoint
3. add metadata endpoint
4. add error handling

## Example Usage

```go
saml_auth.Use()
```

```go
func (c *ExampleConfigurer) Configure(ws security.WebSecurity) {
    ws.Route(matcher.RouteWithPattern(c.config.Endpoints.SamlSso.Location.Path)).
		With(saml_auth.NewEndpoint().
			Issuer(c.config.Issuer).
			SsoCondition(c.config.Endpoints.SamlSso.Condition).
			SsoLocation(c.config.Endpoints.SamlSso.Location).
			MetadataPath(c.config.Endpoints.SamlMetadata))
	
	//Add more configuration to WS to finish the rest of the configuration for your app (i.e. what idp to use, etc)
}
```