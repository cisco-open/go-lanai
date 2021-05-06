# go-lanai

Microservice Library

Readings:

- [Get Started for Developers](docs/Develop.md)
- [Get Started for DevOps](docs/CICD.md)

## Framework and Library Evaluation

### Dependency Injection
Dependency injection is a design pattern that helps developer decouple the components of their program. It helps achieve the separation of
concern of construction and use of objects, which increases readablity and code reuse. In addition, it facilitates testing of individual components
because the dependent object can be easily replaced with mocked objects.

We have evaluated the following libraries:

#### [dig](https://github.com/uber-go/dig)
Pros:
1. It requires very few lines of boiler plate code. Its concepts will be familiar to those who come from a Spring (Java) background. Developer defines providers
which are similar to bean definitions in Spring, and an entry point to the dependency graph called "invoke". When invoke is called, Dig will go through
the defined providers to construct the required dependencies which in turn resolves the dependency graph. Under the hood, it implements a container
where the components are kept and looked up when its needed as a dependency.

2. Providers can be grouped in modules. This makes it easy to organize code, especially when we want to separate some functionality into libraries.

Cons:
1. Dependencies are resolved at run time. This may increase start up time compared to other compile time solutions.

#### [wire](https://github.com/google/wire)
Pros:
1. Unlike dig, wire is a code generator instead of dependency injection framework. Developer defines providers and initializers, but instead
of resolving the dependency at run time, developer uses wire to generate constructor code. The end result is constructor calls that wires up the program
as if its written manually be the developer. This results in compile time dependency injection, which is faster in theory.

Cons:
1. This results in a lot of generated boilerplate code that needs to be checked in the repository.
2. When dependency changes, the code need to be re-generated.

#### Conclusion
We decided to use dig for its ease of use. While using dig, we can always hand wire the lower level objects, or use wire to generate code for the lower
level objects for performance reason, and let dig handle the top level organization. That is to say, we can use dig until there is a need to optimize the 
application startup time.

### Application Lifecycle
Because we are using dig for dependency injection, we can conveniently use [FX](https://github.com/uber-go/fx) for application lifecycle.
Fx lets components hook into application life cycle events.

### Web Framework
For web framework, we need it to be able to:
1. route requests to handlers
2. bind path and request parameter variables, parse request body
3. allow middleware definition
4. render views (html)

We have evaluated the following web framework

#### [Gin Gonic](https://github.com/gin-gonic)
Pros
1. It supports all the requirements we have for a web framework.
2. It has a community of extensions
3. Actively maintained

Cons
1. It has its own handler interface

We have also looked into other alternatives such as go-restful, gorrila which has smaller adoption and less functionalities. In the end 
we are leaning towards Gin because of its larger adoption and richer feature set.

### Microservice and Cloud Support
For Microservice and cloud support, we are looking for functionality that directly relate to running the application as micro services with cloud infrastructure.
This means things like consul discovery, remote configuration, load balanced clients etc.

The choices here are limited. The libaries we looked at are [go-kit](https://github.com/go-kit/kit) and [micro](https://github.com/micro/go-micro).

Since these two libaries have very different approaches where go-kit is a collection of components where developers can pick and choose functionalities, and micro
is more leaning towards a platform where there is a very opinionated way of doing things. We have to use go-kit because we need to build the msx platform to our own specifications.

