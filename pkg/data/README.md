# Data

The ```data``` package is designed to provide a consistent programming model for data access regardless of the underlying data store.
It contains sub packages that are specific to given database. Postgresql and CockroachDB are the databases that are supported by default. 
Application can enable other databases as long as they are supported by [Gorm](https://gorm.io/docs/connecting_to_the_database.html).

## Example Usage

To use this package, include the following code snippet in your application. With these two lines, go-lanai will instantiate
all the components provided in the data package, as well as the components specifically for CockroachDB 

```go
	data.Use()
	cockroach.Use()
```

Add the following section in `application.yml`. These are the connection parameters to your database.

```yaml
data:
  logging:
    level: warn
    slow-threshold: 5s
  db:
    host: localhost
    port: 26257
    sslmode: disable
    username: my_user_name
    Password: my_password
    database: my_db_name
```

Define your database model. The following example is a model for a database table called `friend` that has three columns `id`, `first_name` and `last_name`

```go
type Friend struct {
    ID        uuid.UUID `gorm:"column:id;primary_key;type:UUID;default:gen_random_uuid();"`
    FirstName string    `gorm:"column:first_name;type:text;not null;"`
    LastName  string    `gorm:"column:last_name;type:text;not null;"`
}
```

Declare a repository for this model. The `repo.GormApi` interface allows you to write low level database queries. 
The `repo.CrudRepository` has convenient methods for CRUD operations.

```go
type FriendsRepository struct {
	repo.GormApi
	repo.CrudRepository
}

func NewFriendRepository(factory repo.Factory) *FriendsRepository {
	crud := factory.NewCRUD(&model.Friend{})

	ret := FriendsRepository{
		CrudRepository: crud,
	}
	if gf, ok := factory.(*repo.GormFactory); ok {
		ret.GormApi = gf.NewGormApi()
	}
	return &ret
}
```

Make this repository available for dependency injection, and you will be able to use it in your code. (Call this ```Use()``` function
in your setup code so that it's available for injection).

```go
func Use() {
	bootstrap.AddOptions(
		fx.Provide(
			NewFriendRepository,
		),
	)
}
```

## CRUD Repository
The `CrudRepository` interface is an abstraction that defines the most commonly used data access operations such as Create,
Read, Update, Delete. `go-lanai` provides implementation for these methods, so they don't have to be repeated in 
application code. This is all that is required to instantiate a `CrudRepository` for a model.

```go
type FriendRepository CrudRepository

func NewFriendRepository(factory Factory) FriendRepository {
    return factory.NewCRUD(&model.Friend{})
}
```

Most `CrudRepository` method takes `Condition` and `Options`. `Condition`s are conditional statements that are appended to the query,
`Options` defines how the query should be processed. For example, a query to return all the friends whose first name is John by page can
be written with a condition and option.

```go
	var friends []model.Friend

	err = r.FindAllBy(
		ctx,
		&friends,
        &model.Friend{FirstName: "John"},
		repo.Page(pageNumber, pageSize),
	)
```

## Gorm
Sometimes application have data access logic that are beyond the CRUD operations. For these situations, developer can 
work directly with the lower level [gorm](https://gorm.io/docs/) API. 

```go
	api := factory.NewGormApi(options...)
```

## Error Translation
Error originating from the database driver are mapped to hierarchical `DataError`. Application code can compare the error
they received to the errors defined in the error hierarchy to inspect the error case.

`go-lanai` also uses this error hierarchy to translate the data access error to web status code, so that if application code
returned the error directly as web response the http response status will be correct.

This is how the error handler translate the status code using the error hierarchy. Application code can also use similar technique
to inspect the error case.

```go
func (t WebDataErrorTranslator) Translate(ctx context.Context, err error) error {
	//nolint:errorlint
	if _, ok := err.(errorutils.ErrorCoder); !ok || !errors.Is(err, ErrorCategoryData) {
		return err
	}

	switch {
	case errors.Is(err, ErrorRecordNotFound), errors.Is(err, ErrorIncorrectRecordCount):
		return t.errorWithStatusCode(ctx, err, http.StatusNotFound)
	case errors.Is(err, ErrorSubTypeDataIntegrity):
        return t.errorWithStatusCode(ctx, err, http.StatusConflict)
	case errors.Is(err, ErrorSubTypeQuery):
		return t.errorWithStatusCode(ctx, err, http.StatusBadRequest)
	case errors.Is(err, ErrorSubTypeTimeout):
		return t.errorWithStatusCode(ctx, err, http.StatusRequestTimeout)
	case errors.Is(err, ErrorTypeTransient):
		return t.errorWithStatusCode(ctx, err, http.StatusServiceUnavailable)
	default:
		return t.errorWithStatusCode(ctx, err, http.StatusInternalServerError)
	}
}
```

## Transaction
The `tx` package provides two ways for application code that requires transaction. The ```func Transaction(ctx context.Context, tx TxFunc, opts ...*sql.TxOptions) error```
function allows application code to provide a function that will be run within a transaction. If this function returns error, any database
operation issued within this function will be rolled back. Otherwise, results will be committed.

In this example, if the second operation failed, the first operation will be rolled back.
```go
e = tx.Transaction(ctx, func(ctx context.Context) (err error) {
    // first operation
    firstFriend := model.Friend{firstName:"John", lastName:"Smith"}
    err = di.Repo.Create(ctx, firstFriend)
	if err != nil {
	    return err	
    }
    // second operation
    another := model.Friend{firstName:"Jane", lastName:"Doe"}
    err = di.Repo.Create(ctx, another)
    return err
})
```

Alternatively, application code can also handle transaction manually using the following set of methods.

```go
// Begin start a transaction. the returned context.Context should be used for any transactional operations.
// If an error is returned, the returned context.Context should be discarded.
func Begin(ctx context.Context, opts ...*sql.TxOptions) (context.Context, error) 
```

```go
// Rollback rollbacks a transaction. The returned context.Context is the original provided context when Begin is called.
// If an error is returned, the returned context.Context should be discarded.
func Rollback(ctx context.Context) (context.Context, error)
```

```go
// Commit commits a transaction. the returned context.Context is the original provided context when Begin is called.
// If an error is returned, the returned context.Context should be discarded.
func Commit(ctx context.Context) (context.Context, error)
```

```go
// SavePoint works with RollbackTo and have to be within a transaction.
// The returned context.Context should be used for any transactional operations between corresponding SavePoint and RollbackTo.
// If an error is returned, the returned context.Context should be discarded.
func SavePoint(ctx context.Context, name string) (context.Context, error)
```

```go
// RollbackTo works with SavePoint and have to be within a transaction.
// The returned context.Context should be used for any transactional operations between corresponding SavePoint and RollbackTo.
// If an error is returned, the returned context.Context should be discarded.
func RollbackTo(ctx context.Context, name string) (context.Context, error)
```

## Special Data Types
### EncryptedMap
`EncryptedMap` is useful when certain aspect of the data needs to be encrypted. The encryption is backed by [Vault](https://developer.hashicorp.com/vault/api-docs/secret/transit#encrypt-data)
transit secret engine.

The following snippet declares a model that has encrypted data.

```go
type EncryptedModel struct {
	ID    int    `gorm:"primaryKey;type:serial;"`
	Name  string `gorm:"uniqueIndex;not null;"`
	Value *EncryptedMap
}
```

Saving to the database is the same as any other model.

```go
v := map[string]interface{}{
    "key1": "value1",
    "key2": 2.0,
}

kid := uuid.New()

pqcrypt.CreateKeyWithUUID(ctx, kid)

m := EncryptedModel{
    ID: 12345678,
    Name:  "my_encrypted_model",
    Value: NewEncryptedMap(kid, v),
}

myRepo.Save(ctx, &m)
```

Reading from the database will decrypt the data.

```go
m := EncryptedModel{}
myRepo.FindById(ctx, &m, 12345678) // m's Value field will have the decrypted map
```

### Tenancy
If a model embeds the `Tenancy` type. This model gets two fields that facilitates multi tenant implementation. The `TenantId` column
will store the tenant ID of this record. The `TenantPath` column will store the path from the Tenant ID to the root tenant if
there is a hierarchical tenant relationship. Database operations on this model will automatically take tenancy into consideration
based on the current security context.

```go
// Tenancy is an embedded type for data model. It's responsible for populating TenantPath and check for Tenancy related data
// when crating/updating. Tenancy implements
// - callbacks.BeforeCreateInterface
// - callbacks.BeforeUpdateInterface
// When used as an embedded type, tag `filter` can be used to override default tenancy check behavior:
// - `filter:"w"`: 	create/update/delete are enforced (Default mode)
// - `filter:"rw"`: CRUD operations are all enforced,
//					this mode filters result of any Select/Update/Delete query based on current security context
// - `filter:"-"`: 	filtering is disabled. Note: setting TenantID to in-accessible tenant is still enforced.
//					to disable TenantID value check, use SkipTenancyCheck
// e.g.
// <code>
// type TenancyModel struct {
//		ID         uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid();"`
//		Tenancy    `filter:"rw"`
// }
// </code>
type Tenancy struct {
	TenantID   uuid.UUID  `gorm:"type:KeyID;not null"`
	TenantPath TenantPath `gorm:"type:uuid[];index:,type:gin;not null"  json:"-"`
}
```
### Misc
These models are provided as convenient types that can be embedded in application model.

```go
type Audit struct {
	CreatedAt time.Time      `json:"createdAt,omitempty"`
	UpdatedAt time.Time      `json:"updatedAt,omitempty"`
	CreatedBy uuid.UUID      `type:"KeyID;" json:"createdBy,omitempty"`
	UpdatedBy uuid.UUID      `type:"KeyID;" json:"updatedBy,omitempty"`
}

type SoftDelete struct {
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleteAt,omitempty"`
}
```

In addition, check the `pqx` package for common data types such as `Duration`, `Jsonb`, `TimeArray`, `UUIDArray`