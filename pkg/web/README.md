# Web

The web module enables the application to become a web application. When this module is used, a [gin web server](https://github.com/gin-gonic/gin) is started. The web module
allows endpoints and middlewares to be added to the web server using dependency injection. The web module abstracts away the
boilerplate code of running the web server, allowing application code to focus on writing the endpoints.

The web module achieves this by providing the following components:

**NewEngine** - this is a wrapped gin web server. Our wrapper allows the request to be pre-processed before being handled by gin web server.
The only request pre-processor we currently provide is a CachedRequestPreProcessors. This is used during the auth process so that the auth server can
replay the original request after the session is authenticated.

**NewRegistrar** - this registrar is used by other packages to register middlewares, endpoints, error translations etc. This registrar is provided so that
any other feature that wants to add to the web server can do so via the registrar.

The web module also has a ```fx.Invoke```. Because of it, when the web module is activated, it starts the web server and adds all the component in the registrar on it when 
the application starts.

