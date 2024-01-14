# Example of Using the Security Package to Write an Authorization Server

In this example, we create an authorization service that is capable of serving as an OAuth2 authorization server.

## Prerequisites

- Golang 1.20+
- GNU Makefile
- Docker for Mac
- `$GOBIN` & `$GOROOT` are correctly set
- `GO111MODULE` is "on"

## Step-by-Step Bootstrapping

In this example, the source code contains the fully finished service. This step by step will walk through the steps
involved in writing this service.

### 1. Create the Initial Project
This involves the following steps. For detail explanation of each step, see the [developer guide](../docs/Develop.md)

1. Create Module.yml
2. Add go.mod 
3. Add Makefile 
4. call ```shell make init CLI_TAG="develop"``` to initialize the project 

### 2. Adding Main File
Add the main file corresponding to the definition in Module.yml. The main file is the entry point for this service.
The main method implementation is boilerplate. Its only purpose is to start the application execution. The OAuth2 features
are configured via the ```serviceinit.Use()``` call.

### 3. Configure the Security Package
Add ```pkg/init``` directory. In this directory, the ```package.go``` file implements the ```Use()``` method. In this method,
all the `go-lanai` packages that this service needs are declared.

The `authserver_configurer.go`, `serserver_configurer.go` and `security_configurer.go` methods provides further customization to the
declared packages. In `authserver_configurer.go`, we provide the implementations that the security packages requires.

### 4. Implement the Security Package Interfaces
The security packages requires the application to provide implementation to interfaces. For example the ```AccountStore```
implementation tells the security package how to look up a user. In this example, all the implementation are in memory.
They are implemented in the `pkg/service` directory.

## Running the Service

```shell
go run cmd/auth-service/main.go
```

Navigate to http://localhost:8900/auth/login, you will see the login page. To see the auth service in full action, run it
together with another example such as the [database example](../database). Auth service will be used to authenticate the user
in those examples. See `configs/application.yml`'s ```security.in-memory.accounts``` property to see the user you can use with this example.