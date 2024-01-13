# Example of Using the Data Package

In this example, we create a web service that allows you to add and retrieve your list of friends.

## Prerequisites

- Golang 1.20+
- GNU Makefile
- Docker for Mac
- `$GOBIN` & `$GOROOT` are correctly set
- `GO111MODULE` is "on"

## Step-by-Step Bootstrapping

In this example, the source code contains the fully finished service. This step by step will walk through the steps
involved in writing this service.

### 1. Bootstrap the service via `lanai-cli codegen`
See Instructions in [skeleton-service](../skeleton/README.md)

At this point, the controllers are implemented with no-op implementation.

### 2. Enable the data packages
In `pkg/init/pakcage.go` initialize the `data` and `cockroach` package.

```go
func Use() {
    // ...
	
	// data related
	data.Use()
	cockroach.Use()
	
	// ...
}
```

In `application.yml` configure the connection properties to the database.

```yaml
data:
  logging:
    level: warn
    slow-threshold: 5s
  cockroach:
    host: ${db.cockroach.host:localhost}
    port: ${db.cockroach.port:26257}
    sslmode: ${db.cockroach.sslmode:disable}
    username: ${spring.datasource.username:root}
    Password: ${spring.datasource.password:root}
    database: skeleton
```

### 3. Define the model and repository

In `pkg/model/friend.go` define the golang struct that maps to the table in the database

In `pkg/repository/friend.go` define the CRUD repository for this model, and define the constructor for this repository.

In `pkg/repository/package.go` define the `Use()` method that would provide the repository instance for injection.

In `pkg/init/package.go` initialize the repository package so that it's available for injection.

```go
func Use() {
    // ...

	repository.Use()
}
```

### 4. Use the repository to finish the controller implementation

Update the controller constructor so that it injects a repository instance.

```go
type ExampleFriendsController struct {
	friendRepo *repository.FriendsRepository
}

type exampleFriendsControllerDI struct {
	fx.In
	FriendsRepo *repository.FriendsRepository // This tells go-lanai that the NewExampleFriendsController constructor requires a FriendRepository instance
}

func NewExampleFriendsController(di exampleFriendsControllerDI) web.Controller {
	return &ExampleFriendsController{
		friendRepo: di.FriendsRepo, // Sets the controller's friendRepo reference so that it can be used later. 
	}
}
```

In each of the controller's method, use the `friendRepo` reference to carry out the database operations.

### 5. Add a data migration step to create the initial table

In `cmd/skeleton-service-migrate/migrate.go` add the main method for the data migration application. 

In `pkg/migrate` add the data migration implementation. `pkg/migrate/package.go` contains the boilder plate code to setting
up the `go-lanai` components needed for running a data migration. `pkg/migrate/migration_v1` includes the migration steps.
In this example there is only one step, which is to create the database table like this.

```sql
DROP TABLE IF EXISTS "public"."users";
CREATE TABLE friends
(
    id                  UUID NOT NULL DEFAULT gen_random_uuid(),
    first_name          STRING NOT NULL,
    last_name           STRING NOT NULL,
    CONSTRAINT          "primary" PRIMARY KEY (id ASC)
);
```

## Running the Service

### 1. Running the database migration

This will create the table if it has not been created. Re-running this command will only execute migration steps that has not been
executed already.

```shell
go run cmd/skeleton-service-migrate/migrate.go
```

### 2. Running the service

```shell
go run cmd/skeleton-service/main.go
```

Navigate to http://localhost:9898/skeleton/swagger. Use the GET API to retrieve all the items stored in the database. Use the POST
API to add more items to the database.