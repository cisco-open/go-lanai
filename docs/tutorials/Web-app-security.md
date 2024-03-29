# Web Application Security

In this tutorial, we write an application that supports SAML single sign on to walk through the process of creating a web application
using go-lanai. By the end of this documentation, you will have a web application with a private API. You can use single sign on
to get access to this API. As we write the application, the corresponding module that is used will be explained in detail. This includes
bootstrap, web and security. The security module has a list of submodules, each corresponding to a feature. The security features that will be
covered by this document is access, session and SAML login.

**Additional Reading**

Typically, developers don't need GNU Make to test/build services. However, go-lanai also provides tooling around the testing, build and release
process that leverages GNU Make.

- [Get Started for Developers](../Develop.md)
- [Get Started for DevOps](../CICD.md)

## Step 1: Bootstrapping the Application
When writing a go-lanai application, the developer selects the go-lanai module they want to use. [Bootstrap](../../pkg/bootstrap/README.md) refers to the process
during which the application instantiate the components needed by the modules and wires them together.

In each go-lanai module, you can usually find a `Use()` function. This function registers the module with go-lanai's bootstrapper.
The application code calls this method to signal that this module should be activated.

Create a project with the following project structure. At the end of this step, you will have an empty application
that starts. It will be used as the base to add on more features in the later section of the document.

```
-example
  -cmd
    main.go
  -configs
    application.yml
    bootstrap.yml
  go.mod    
```

**cmd/main.go**

This is the entry point of the application. Here we create a new app and execute it. Notice the ```appconfig.Use()```
call in the ```init()``` function. We are declaring that we want to use the **appconfig** module. This module allows reading
configuration values from various sources (yaml files, consul, vault, etc) into golang structs.

```go
package main

import (
	appconfig "github.com/cisco-open/go-lanai/pkg/appconfig/init"
	"github.com/cisco-open/go-lanai/pkg/bootstrap"
	"go.uber.org/fx"
	"time"
)

func init() {
	appconfig.Use()
}

func main() {
	// bootstrapping
	bootstrap.NewAppCmd(
		"example",
		[]fx.Option{},
		[]fx.Option{
			fx.StartTimeout(60 * time.Second),
		},
	)
	bootstrap.Execute()
}
```

**configs/bootstrap.yml**

The bootstrap.yml file is one of the first property sources the application reads on startup. The consul and vault connection
properties are specified here because they can become property sources as well.

```yaml
application:
  name: example

cloud:
  discovery:
    consul:
      ip-address: ${spring.cloud.consul.discovery.ipaddress:}
  consul:
    host: ${spring.cloud.consul.host:localhost}
    port: 8500
    config:
      enabled: true
  vault:
    kv:
      enabled: true
    host: ${spring.cloud.vault.host:localhost}
    port: 8200
    scheme: http
    authentication: TOKEN
    token: replace_with_token_value #replace with actual token value or provide this value via other property source (i.e. env variable or commandline args)

server:
  port: 8090
  context-path: /example

# This section will refresh the logger configuration after bootstrap is invoked.
log:
  levels:
    Bootstrap: warn
    Web: debug
    Data: info
    Kafka: info
    SEC.Session: info
    OAuth2.Auth: info
```

**configs/application.yml**

More application specific properties can be configured here. The values configured here can be overridden by properties in
consul or vault.

```yaml
# standarized information on the service
info:
  app:
    show-build-info: true
    name: example
    description: A example go-lanai service
    version: ${project.version}
    attributes:
      displayName: Example Service
```

## Step 2: Add a REST API

One of the common use case of go-lanai is to write a web application. The [web](../../pkg/web/README.md) facilitates this by abstracting the boilerplate code
for running a web application, allowing application code to focus on writing the API endpoints. In this step, we will turn the application
into a web application, and add a REST endpoint to it.

Add a controller directory to your project. In this directory we will create a rest endpoint that prints hello when it's called.
```
-example
  -cmd
    main.go
  -configs
    application.yml
    bootstrap.yml
  -pkg
    -controller
      hello.go
      package.go  
  go.mod    
```

**hello.go**

In this file, we define a struct called ```helloController```. This struct implements ```web.Controller``` interface which
defines the ```Mappings() []web.Mapping``` method. The API endpoint to implementation is expressed as mappings.

```go
package controller

import (
	"context"
	"github.com/cisco-open/go-lanai/pkg/web"
	"github.com/cisco-open/go-lanai/pkg/web/rest"
)

type helloController struct{}

func newHelloController() web.Controller {
	return &helloController{}
}

func (c *helloController) Mappings() []web.Mapping {
	return []web.Mapping{
		rest.New("hello").Get("/api/v1/hello").
			EndpointFunc(c.Hello).Build(),
	}
}

func (c *helloController) Hello(_ context.Context) (interface{}, error) {
	return "hello", nil
}
```

**package.go**

Using the same pattern, we provide the controllers to make them available to the web module. The ```web.FxControllerProviders```
is a ```fx.Option``` that injects the controllers to the web module's registrar.

```go
package controller

import (
	"github.com/cisco-open/go-lanai/pkg/bootstrap"
	"github.com/cisco-open/go-lanai/pkg/web"
)

func Use() {
	bootstrap.AddOptions(
		web.FxControllerProviders(newHelloController),
	)
}
```

**main.go**

In the main file, we call ```web.Use()``` and ```controller.Use``` to activate these two modules.

```go
package main

import (
	appconfig "github.com/cisco-open/go-lanai/pkg/appconfig/init"
	"github.com/cisco-open/go-lanai/pkg/bootstrap"
	web "github.com/cisco-open/go-lanai/pkg/web/init"
	"github.com/cisco-open/example/pkg/controller"
	"go.uber.org/fx"
	"time"
)

func init() {
	appconfig.Use()
	web.Use()

	controller.Use()
}

func main() {
	// bootstrapping
	bootstrap.NewAppCmd(
		"example",
		[]fx.Option{},
		[]fx.Option{
			fx.StartTimeout(60 * time.Second),
		},
	)
	bootstrap.Execute()
}
```
At this point, if you run the application, you should be able to visit http://localhost:8090/example/api/v1/hello and the browser
should display "hello".

## Step 3: Add Access Restriction

The [security module](../../pkg/security/README.md) allows application code to configure security related requirements on its endpoints.
This module is organized into sub packages each corresponding to a security features. The top level ```security.Use()```
does nothing on its own. It simply provides a mechanism where application code can express its security requirements through configuration.

Application code defines its security requirements by providing security configurers. Each configurer implements a method that configures a ```WebSecurity``` instance.
Each ```WebSecurity``` instance holds a combination of ```Route``` (the path and method pattern for which this WebSecurity applies),
```Condition``` (additional conditions of incoming requests for which this WebSecurity applies to) and ```Features``` (security features to apply when the request matches the ```Route``` and ```Condition```).
Behind the scenes, the security module will add appropriate middleware and endpoints according to the ```WebSecurity``` configurations.

In this step, we will use the access control feature to restrict the hello endpoint to authenticated users only.

Add an init directory. The security configurations will be placed in this directory.
```
-example
  -cmd
    main.go
  -configs
    application.yml
    bootstrap.yml
  -pkg
    -controller
      hello.go
      package.go  
    -init
      package.go
      security.go  
  go.mod    
```

**init/package.go**

In this file we specify we want to use ```security.Use()```, and we add a ```fx.Invoke(configureSecurity)``` to register
our WebSecurity configurer.

```go
package serviceinit

import (
	appconfig "github.com/cisco-open/go-lanai/pkg/appconfig/init"
	"github.com/cisco-open/go-lanai/pkg/bootstrap"
	"github.com/cisco-open/go-lanai/pkg/security"
	web "github.com/cisco-open/go-lanai/pkg/web/init"
	"github.com/cisco-open/example/pkg/controller"
	"go.uber.org/fx"
)

func Use() {
	appconfig.Use()
	web.Use()
	security.Use()

	bootstrap.AddOptions(
		fx.Invoke(configureSecurity),
	)

	controller.Use()
}

type secDI struct {
	fx.In
	SecRegistrar security.Registrar
}

func configureSecurity(di secDI) {
	di.SecRegistrar.Register(&securityConfigurer{})
}
```

**init/security.go**

Here we implement our WebSecurity configurer. Our configurer specifies that requests to paths that starts with `/api/`
should be authenticated.

```go
package serviceinit

import (
	"github.com/cisco-open/go-lanai/pkg/security"
	"github.com/cisco-open/go-lanai/pkg/security/access"
	"github.com/cisco-open/go-lanai/pkg/web/matcher"
)

type securityConfigurer struct{}

func (c *securityConfigurer) Configure(ws security.WebSecurity) {

	// DSL style example
	// for REST API
	ws.Route(matcher.RouteWithPattern("/api/**")).
		With(access.New().
			Request(matcher.AnyRequest()).Authenticated(),
		)
}
```

**main.go**

We update our main file with ```serviceinit.Use()``` to make our security configuration active.
```go
package main

import (
	"github.com/cisco-open/go-lanai/pkg/bootstrap"
	serviceinit "github.com/cisco-open/example/pkg/init"
	"go.uber.org/fx"
	"time"
)

func init() {
	serviceinit.Use()
}

func main() {
	// bootstrapping
	bootstrap.NewAppCmd(
		"example",
		[]fx.Option{},
		[]fx.Option{
			fx.StartTimeout(60 * time.Second),
		},
	)
	bootstrap.Execute()
}
```

At this point if you run the application and visit the application again, you should get an unauthenticated error.

## Step 4: Add Session

Because we have restricted access to the `/hello` endpoint to authenticated user only, we need way to determine if the
request is authenticated or not. In order to do that we need to enable session. When session is enabled, session cookie
will be set on the response and sent with the request. We can then save the authentication state of the user in the session.
When session feature is enabled on a WebSecurity configuration, the request and response to those endpoints will load
and persist session information.

**init/package.go**

Session is stored in redis, so we add ```redis.Use()``` activate the redis module.

```go
func Use() {
	appconfig.Use()
	web.Use()
	security.Use()
	redis.Use() //needs redis for session store

	bootstrap.AddOptions(
		fx.Invoke(configureSecurity),
	)

	controller.Use()
}
```

**security.go**

Enable session on the web security configuration.

```go
func (c *securityConfigurer) Configure(ws security.WebSecurity) {

	// DSL style example
	// for REST API
	ws.Route(matcher.RouteWithPattern("/api/**")).
		With(access.New().
			Request(matcher.AnyRequest()).Authenticated(),
		).
		With(session.New())
}
```

**configs/application.yml**

Add redis and session related properties.

```yaml
redis:
  addrs: localhost:6379
  db: 0

security:
  session:
    cookie:
      domain: localhost
    max-concurrent-sessions: 0
    idle-timeout: "5400s"
    absolute-timeout: "10800s"
    db-index: 10
```

At this point, if you visit the endpoint again and look at the network traffic, you will see the request and response contains a session cookie.

## Step 5: Add SAML Login

Enable the SAML Login feature so that when user visits the `/hello` endpoint, they will be redirected to the single sign on
page first. This feature adds a number of middleware and endpoints to allow your application to act as a SAML service provider. See [SAML login feature](pkg/security/saml/sp)
for the list of middleware and endpoints added by this feature.

**init/package.go**

Activate the SAML login module with ```samlsp.Use()```. We also add a ```fx.Provide(authserver.BindAuthServerProperties)```.
This binds the auth related properties into a struct which we will use when configuring the SAML login feature on the web security instance.

We update the ```configureSecurity``` method to take these property structs as dependencies.

```go
package serviceinit

import (
	appconfig "github.com/cisco-open/go-lanai/pkg/appconfig/init"
	"github.com/cisco-open/go-lanai/pkg/bootstrap"
	"github.com/cisco-open/go-lanai/pkg/redis"
	"github.com/cisco-open/go-lanai/pkg/security"
	"github.com/cisco-open/go-lanai/pkg/security/config/authserver"
	samlsp "github.com/cisco-open/go-lanai/pkg/security/saml/sp"
	"github.com/cisco-open/go-lanai/pkg/web"
	webinit "github.com/cisco-open/go-lanai/pkg/web/init"
	"github.cisco.com/cisco-open/example/pkg/controller"
	"go.uber.org/fx"
)

func Use() {
	appconfig.Use()
	webinit.Use()
	security.Use()
	redis.Use() //needs redis for session store
	samlsp.Use() // enable this service to act as SAML sp

	bootstrap.AddOptions(
		fx.Provide(authserver.BindAuthServerProperties), //using the property already defined in go-lanai for convenience
		fx.Invoke(configureSecurity),
	)

	controller.Use()
}

type secDI struct {
	fx.In
	SecRegistrar     security.Registrar
	AuthProperties   authserver.AuthServerProperties
	ServerProperties web.ServerProperties
}

func configureSecurity(di secDI) {
	di.SecRegistrar.Register(&securityConfigurer{
		AuthProperties:   di.AuthProperties,
		ServerProperties: di.ServerProperties,
	})
}
```

**init/security.go**

Enable the SAML login feature on the web security configuration. This indicates that when authentication is needed,
we want to use SAML login.

```go
package serviceinit

import (
	"github.com/cisco-open/go-lanai/pkg/security"
	"github.com/cisco-open/go-lanai/pkg/security/access"
	"github.com/cisco-open/go-lanai/pkg/security/config/authserver"
	samlsp "github.com/cisco-open/go-lanai/pkg/security/saml/sp"
	"github.com/cisco-open/go-lanai/pkg/security/session"
	"github.com/cisco-open/go-lanai/pkg/web"
	"github.com/cisco-open/go-lanai/pkg/web/matcher"
)

type securityConfigurer struct {
	AuthProperties   authserver.AuthServerProperties
	ServerProperties web.ServerProperties
}

func (c *securityConfigurer) Configure(ws security.WebSecurity) {

	// DSL style example
	// for REST API
	ws.Route(matcher.RouteWithPattern("/api/**")).
		With(access.New().
			Request(matcher.AnyRequest()).Authenticated(),
		).
		With(session.New()).
		With(samlsp.New().
			Issuer(security.NewIssuer(func(opt *security.DefaultIssuerDetails) {
				*opt = security.DefaultIssuerDetails{
					Protocol:    c.AuthProperties.Issuer.Protocol,
					Domain:      c.AuthProperties.Issuer.Domain,
					Port:        c.AuthProperties.Issuer.Port,
					ContextPath: c.ServerProperties.ContextPath,
					IncludePort: true,
				}
			})))
}
```

**configs/application.yml**

Add SAML related properties to application.yml

```yaml
security:
  auth:
    issuer:
      domain: localhost
      protocol: http
      port: 8900
      context-path: ${server.context-path}
      include-port: true
    saml:
      certificate-file: "configs/saml.cert"
      key-file: "configs/saml.key"
      key-password: "foobar"
```

**configs/saml.cert** and **configs/saml.key**

You'll also need to add a cert and key pair to the configs directory. You can create them using `openssl`. For example:

Generating the saml.key

```shell
openssl genrsa -out saml.key -aes256 1024
```

Generating saml.cert from the key

```shell
openssl req -key saml.key -new -x509 -days 365 -out saml.cert
```

At this point, if you run the service you will get errors complaining about missing dependencies. This is because the SAML features
don't know where to load identity provider data and user data. For this the SAML feature defines the following interfaces. You will
need to provide implementation for them.

1. ```IdentityProviderManager```
2. ```SamlIdentityProviderManager```
3. ```FederatedAccountStore```

Add a service directory and put the implementation in there.

```
-example
  -cmd
    main.go
  -configs
    application.yml
    bootstrap.yml
  -pkg
    -controller
      hello.go
      package.go  
    -init
      package.go
      security.go  
    -service
       account_store.go
       idp_manager.go
       package.go
          
  go.mod    
```

**service/account_store.go**

This example implementation loads a user based on the value in the incoming assertion. A real implementation may need to
create the user record in the database, or look up existing user.

```go
package service

import (
	"context"
	"github.com/cisco-open/go-lanai/pkg/security"
	"fmt"
)

type AccountStore struct {
}

func NewAccountStore() security.FederatedAccountStore {
	return &AccountStore{}
}

func (a *AccountStore) LoadAccountByExternalId(ctx context.Context, externalIdName string, externalIdValue string, externalIdpName string, autoCreateUserDetails security.AutoCreateUserDetails, rawAssertion interface{}) (security.Account, error) {
	return &security.DefaultAccount{
		AcctDetails: security.AcctDetails{
			ID:       fmt.Sprintf("%s:%s", externalIdName, externalIdValue),
			Type:     security.AccountTypeFederated,
			Username: externalIdValue,
		}}, nil
}
```

**service/idp_manager.go**

This implements both the ```IdentityProviderManager``` and ```SamlIdentityProviderManager``` interfaces. This implementation always
returns a hardcoded IDP. A real implementation should return IDP from storage.

```go
package service

import (
	"context"
	"github.com/cisco-open/go-lanai/pkg/security/idp"
	"github.com/cisco-open/go-lanai/pkg/security/idp/extsamlidp"
)

var targetIdp = extsamlidp.SamlIdentityProvider{
	SamlIdpDetails: extsamlidp.SamlIdpDetails{
		EntityId:                 "http://msx.com:8900/auth",
		Domain:                   "localhost",
		MetadataLocation:         "http://msx.com:8900/auth/metadata",
		ExternalIdName:           "email",
		ExternalIdpName:          "msx",
		MetadataRequireSignature: false,
		MetadataTrustCheck:       false,
	},
}

type IdpManager struct {
}

func NewIdpManager() idp.IdentityProviderManager {
	return &IdpManager{}
}

func (i *IdpManager) GetIdentityProviderByEntityId(ctx context.Context, entityId string) (idp.IdentityProvider, error) {
	return targetIdp, nil
}

func (i *IdpManager) GetIdentityProvidersWithFlow(ctx context.Context, flow idp.AuthenticationFlow) []idp.IdentityProvider {
	return []idp.IdentityProvider{targetIdp}
}

func (i *IdpManager) GetIdentityProviderByDomain(ctx context.Context, domain string) (idp.IdentityProvider, error) {
	return targetIdp, nil
}
```

**service/package.go**

Provide our implementations in a ```Use()``` function.

```go
package service

import (
	"github.com/cisco-open/go-lanai/pkg/bootstrap"
	"go.uber.org/fx"
)

func Use() {
	bootstrap.AddOptions(
		fx.Provide(NewIdpManager),
		fx.Provide(NewAccountStore),
	)
}
```

**init/package.go**

Update init/package.go to provide these services

```go
package serviceinit

func Use() {
	appconfig.Use()
	webinit.Use()
	security.Use()
	redis.Use() //needs redis for session store
	samlsp.Use()

	bootstrap.AddOptions(
		fx.Provide(authserver.BindAuthServerProperties), //using the property already defined in go-lanai for convenience
		fx.Invoke(configureSecurity),
	)

	controller.Use()
	service.Use()
}
```

At this point, when you visit the `/api/v1/hello` endpoint, you should be redirected to SSO login and visit the page (if your IDP is running).

You can also update the controller so that it prints the current user's name.

```go
package controller

func (c *helloController) Hello(ctx context.Context) (interface{}, error) {
	authentication := security.Get(ctx)
	username, _ := security.GetUsername(authentication)
	return fmt.Sprintf("hello %s", username), nil
}
```