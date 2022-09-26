# Writing Applications Using Go-Lanai

Go-Lanai is a set of application frameworks and modules that make writing applications (especially micro-services) easy. A module represents a feature provided by go-lanai. 
Go-lanai's module framework is built on top 
of the [dependency injection](https://en.wikipedia.org/wiki/Dependency_injection) framework provided by [uber-go/fx](https://github.com/uber-go/fx). 
Understanding of the fx framework is needed to understand the rest of the documentation, especially [fx.Provide](https://pkg.go.dev/go.uber.org/fx#Provide) and [fx.Invoke](https://pkg.go.dev/go.uber.org/fx#Invoke) 

Modules provided by
go-lanai includes:
- actuator
- appconfig
- boostrap
- consul
- data
- discovery
- dsync
- integrate
- kafka
- log
- migration
- opensearch
- profiler
- redis
- scheduler
- security
- swagger
- tenancy
- tlsconfig
- tracing
- vault
- web

This document uses an application that supports SAML single sign on to walk through the process of creating a web application
using go-lanai. At the end of this documentation. You will have a web application that has a private API. You can use single sign on
to get access to this API. As we write the application, the corresponding module that is used will be explained in detail. This includes
bootstrap, web, security. The security module has a list of submodules, each corresponding to a feature. The security features that will be
covered by this document is access, session and saml login.

**Additional Reading**

Typically, developers don't need GNU Make to test/build services. However, go-lanai also provides tooling around the testing, build and release
process that leverages GNU Make.

- [Get Started for Developers](docs/Develop.md)
- [Get Started for DevOps](docs/CICD.md)

## Bootstrap
When writing a go-lanai application, the developer selects the go-lanai module they want to use. Bootstrap refers to the process of 
how the application instantiate the components needed by the modules and wire them together. 

Under the hood, the bootstrapper keeps a registry of modules that's enabled in this application. 
A module is implemented as a group of ```fx.Provide``` and ```fx.Invoke``` options.
When the app starts, all module added ```fx.provide``` and ```fx.invoke``` options are sorted and executed.

In go-lanai's module packages, you'll usually see a function like this: 

```go
package init

var Module = &bootstrap.Module{
	Name: "web",
	Precedence: web.MinWebPrecedence,
	PriorityOptions: []fx.Option{
		appconfig.FxEmbeddedDefaults(defaultConfigFS),
		fx.Provide(
			web.BindServerProperties,
			web.NewEngine,
			web.NewRegistrar),
		fx.Invoke(setup),
	},
}

// Use Allow service to include this module in main()
func Use() {
	bootstrap.Register(Module)
	bootstrap.Register(cors.Module)
}
```

This Use() method registers the module with the bootstrapper. The application 
code just needs to call the ```Use()``` function to indicate this module should be activated in the application.

### Tutorial
Create a project with the following project structure. At the end of this tutorial section, you will have an empty application
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
	appconfig "cto-github.cisco.com/NFV-BU/go-lanai/pkg/appconfig/init"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
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
  consul:
    host: ${spring.cloud.consul.host:localhost}
    port: 8500
    config:
      enabled: true
    discovery:
      health-check-critical-timeout: 1h
      ip-address: ${spring.cloud.consul.discovery.ipaddress:}
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
  # should use bridged
#  context-path: ${server.servlet.context-path:/auth}

# This section will refresh the logger configuration after bootstrap is invoked.
log:
  levels:
    Bootstrap: warn
    Web: debug
    Data: info
    Kafka: info
    SEC.Session: info
    OAuth2.Auth: info
#  loggers:
#    text-file:
#      type: file
#      format: text
#      location: "logs/text.log"
#      template: '{{pad .time -25}} {{lvl . 5}} [{{pad .caller 25 | blue}}] {{pad .logger 12 | green}}: [{{trace .traceId .spanId .parentId}}] {{.msg}} {{kv .}}'
#      fixed-keys: "spanId, traceId, parentId, http"
#    json-file:
#      type: file
#      format: json
#      location: "logs/json.log"
```

**configs/application.yml**

More application specific properties can be configured here. The values configured here can be overridden by properties in 
consul or vault.

```yaml
# standarized information on the service
info:
  app:
    msx:
      # defaults to this services version, can be overridden in consul via installer
      version: ${info.app.version}
      show-build-info: true
    name: example
    description: A example go-lanai service
    version: ${project.version}
    attributes:
      displayName: Example Service
```

## Web Module
The web module enables the application to become a web application. When this module is used, a [gin web server](https://github.com/gin-gonic/gin) is started. The web module
allows endpoints and middlewares to be added to the web server using dependency injection. The web module abstracts away some of the
boiler plate code of running the web server, allowing application code to focus on writing the endpoints.

The web module achieves this by providing the following components:

**NewEngine** - this is a wrapped gin web server. Our wrapper allows the request to be pre-processed before being handled by gin web server. 
The only request pre-processor we currently provide is a CachedRequestPreProcessors. This is used during the auth process so that the auth server can 
replay the original request after the session is authenticated.

**NewRegistrar** - this registrar is used by other packages to register middlewares, endpoints, error translations etc. This registrar is provided so that
any other feature that wants to add to the web server can do so via the registrar.

The web module also has a ```fx.Invoke``` which starts the web server and adds all the component in the registrar on it.

### Tutorial
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
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/rest"
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
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
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
	appconfig "cto-github.cisco.com/NFV-BU/go-lanai/pkg/appconfig/init"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	web "cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/init"
	"cto-github.cisco.com/tishi/example/pkg/controller"
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

## Security Module

The security module is organized into sub packages each corresponding to a security features. The top level ```security.Use()```
does nothing on its own. It simply provides a mechanism where application code can express its security requirements through configuration.

The security module does this by providing a ```Initializer``` and a ```Registrar```. 

The registrar's job is to keep list of two things:
1. **WebSecurity Configurer**

    A ```WebSecurity``` struct holds information on security configuration. This is expressed through a combination of ```Route``` (the path and method pattern which this WebSecurity applies),
    ```Condition``` (additional conditions of incoming requests, which this WebSecurity applies to) and ```Features``` (security features to apply).

    To define the desired security configuration, calling code provides implementation of the ```security.Configurer``` interface. It requires a ```Configure(WebSecurity)``` method in 
    which the calling code can configure the ```WebSecurity``` instance. Usually this is provided by application code.


2. **Feature Configurer**
    
    A ```security.FeatureConfigurer``` is internal to the security package, and it's not meant to be used by application code.
    It defines how a particular feature needs to modify ```WebSecurity```. Usually in terms of what middleware handler functions need to be added.
    For example, the Session feature's configurer will add a couple of middlewares handler functions to the ```WebSecurity``` to load and persist session.

The initializer's job is to apply the security configuration expressed by all the WebSecurity configurers. It does so by looping through 
the configurers. Each configurer is given a new WebSecurity instance, so that the configurer can express its security configuration on this WebSecurity instance. 
Then the features specified on this ```WebSecurity``` instance is resolved using the corresponding feature configurer. At this point the ```WebSecurity``` is
expressed in request patterns and middleware handler functions. The initializer then adds the pattern and handler functions as
mappings to the web registrar. The initializer repeats this process until all the WebSecurity configurers are processed. 

### Access Module
This module provides access control feature. You can use this feature on a WebSecurity instance to indicate which endpoints
have what kind of access control.

#### Tutorial
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
	appconfig "cto-github.cisco.com/NFV-BU/go-lanai/pkg/appconfig/init"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	web "cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/init"
	"cto-github.cisco.com/tishi/example/pkg/controller"
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

Here we implement our WebSecurity configurer. Our configurer specifies that request to any path that starts with "/api/"
should be authenticated.

```go
package serviceinit

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/access"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/matcher"
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
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	serviceinit "cto-github.cisco.com/tishi/example/pkg/init"
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

### Session Module
When session feature is enabled on a WebSecurity configuration, the request and response to those endpoints will load
and persist session information.

#### Tutorial
Because we have restricted access to the /hello endpoint to authenticated user only, we need way to determine if the 
request is authenticated or not. In order to do that we need to enable session. When session is enabled, session cookie
will be set on the response and sent with the request. We can then use to save the authentication state of the user session.

**init/package.go**

Session is stored in redis so we add ```redis.Use()``` activate the redis module.

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

### Saml Login Module
This module enables a service to act as an SP (allows login with third party using SAML protocol). This feature has two 
feature configurers. 

**login feature configurer** does the following:

1. Add metadata endpoint (/saml/metadata)
2. Add ACS endpoint (/saml/SSO)
3. Add metadata refresh middleware that covers the above two endpoints
4. Make the metadata endpoint and acs endpoint public
5. Add an authentication entry point that will trigger the saml login process

**logout feature configurer** does the following:

1. Add single logout endpoint
2. Add metadata refresh middleware that covers the endpoint
3. Add logout handler
4. Add logout entry point (the entry point to send out the logout request to the IDP)

When SAML login feature is enabled, these middleware and endpoints are added to the web security configuration.

#### Tutorial
Enable the SAML login feature so that when user visits the /hello endpoint, they will be redirected to the single sign on
page first.

**init/package.go**

Activate the saml login module with ```samllogin.Use()```. We also add a ```fx.Provide(authserver.BindAuthServerProperties)```.
This binds the auth related properties into a struct which we will use when configuring the SAML login feature on the web security instance.

We update the ```configureSecurity``` method to take these property structs as dependencies.

```go
package serviceinit

import (
	appconfig "cto-github.cisco.com/NFV-BU/go-lanai/pkg/appconfig/init"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/redis"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/config/authserver"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/saml/samllogin"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	webinit "cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/init"
	"cto-github.cisco.com/tishi/example/pkg/controller"
	"go.uber.org/fx"
)

func Use() {
	appconfig.Use()
	webinit.Use()
	security.Use()
	redis.Use() //needs redis for session store
	samllogin.Use() // enable this service to act as saml sp

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
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/access"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/config/authserver"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/saml/samllogin"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/session"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/matcher"
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
		With(samllogin.New().
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

You'll also need to add a cert and key pair to the configs directory.

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
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
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
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/idp"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/idp/samlidp"
)

var targetIdp = samlidp.SamlIdentityProvider{
	SamlIdpDetails: samlidp.SamlIdpDetails{
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
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"go.uber.org/fx"
)

func Use() {
	bootstrap.AddOptions(
		fx.Provide(NewIdpManager),
		fx.Provide(NewAccountStore),
	)
}
```

**init.go**

Update init.go to provide these services.

```go
package serviceinit

func Use() {
	appconfig.Use()
	webinit.Use()
	security.Use()
	redis.Use() //needs redis for session store
	samllogin.Use()

	bootstrap.AddOptions(
		fx.Provide(authserver.BindAuthServerProperties), //using the property already defined in go-lanai for convenience
		fx.Invoke(configureSecurity),
	)

	controller.Use()
	service.Use()
}
```

At this point, when you visit the /api/v1/hello endpoint, you should be redirected to SSO login and visit the page (if your IDP is running).

You can also update the controller so that it prints the current user's name.

```go
package controller

func (c *helloController) Hello(ctx context.Context) (interface{}, error) {
	authentication := security.Get(ctx)
	username, _ := security.GetUsername(authentication)
	return fmt.Sprintf("hello %s", username), nil
}
```


### SAML SSO Module
This module allows a service to act as IDP (allow others to SSO with the service).

This module registers a feature configurer which does the following:

1. add metadata refresh middleware to the sso endpoint
2. add sso endpoint
3. add metadata endpoint
4. add error handling

#### Tutorial

TODO:

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

#### Pre-made configuration for auth server and resource server
TODO