# Security

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
