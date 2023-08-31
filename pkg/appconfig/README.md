# Application Config

The ```appconfig``` package allows application properties to be defined in multiple sources including environment variables,
command line, yaml file, consul and vault.

## Example Usage

To use the package, include ```appconfig.Use()``` in your application.

To bind properties to a golang struct, use the ```func (c *config) Bind(target interface{}, prefix string) error``` method.
The following example will bind the following yaml property into the ```Properties``` struct.
```yaml
info:
  app:
    name: my_app
    description: an example application
```

```go
const (
	PropertiesPrefix = "info.app"
)

type Properties struct {
	name string `json:"name"`
	description string `json:"description"`
}


//NewProperties create a Properties with default values
func NewProperties() *Properties {
	return &Properties{}
}

//BindProperties create and bind SessionProperties, with a optional prefix
func BindProperties(ctx *bootstrap.ApplicationContext) Properties {
	props := NewProperties()
	if err := ctx.Config().Bind(props, PropertiesPrefix); err != nil {
		panic(errors.Wrap(err, "failed to bind Properties"))
	}
	return *props
}
```

## Property Source and Precedence
Properties are loaded from multiple sources in two stages. This is because in order to connect to external property sources like Vault and Consul
the application needs to read their connection properties. The bootstrap stage allows these to be read in so that they can be used in the application stage.

### Bootstrap Stage
In this stage, properties are loaded from the following sources. The sources are listed based on priority order from low to high. High
priority sources overrides lower priority sources

* embedded default yaml
* bootstrap yaml 
* environment variables
* command line

### Application Stage
In the application stage, more property sources are added so the complete list becomes

* embedded default yaml
* bootstrap yaml
* application yaml
* environment variables
* command line
* consul and vault default name space
* consul and vault application name space

### Profiles
Application can be run using profiles to load profile specific properties. The property sources that supports profile are:

* bootstrap yaml
* application yaml
* consul
* vault

For yaml files, when application is run with profile the application will load those yaml files that have the profile name appended.
i.e. bootstrap-{profile_name}.yml and application-{profile_name}.yml in addition to the default bootstrap.yml and application.yml file.
The properties in the profile specific yml file will override the properties in the default file.

For consul, properties in the name space that have the profile appended will be loaded in addition to the non-profiled name space. i.e.
e.g. properties in ```userviceconfiguration/defaultapplication,{profile-name}``` in addition to ```userviceconfiguration/defaultapplication``` 

For vault, the properties in the profile context will be loaded in addition to those in the non-profiled context. e.g. ```defaultapplication/{profile-name}```
in addition to ```defaultapplication/```

#### active profile vs additional profile
Profiles can be specified either using ```application.profiles.active``` or ```application.profiles.additional```. To provide these two property
values via command line, use ```active-profiles``` and ```additional-profiles``` (e.g. ```service --active-profiles a```)

```application.profiles.active```
This property gets overridden like any other property if it's defined in multiple property sources. 

```application.profiles.additional```
Unlike other properties, this property is appended if it's defined in multiple property sources.

Because each property source can add more profiles, the property loading process in both bootstrap and application stage will continue to refresh the list of property sources until there are no new
property sources. i.e. when the value of these two properties stabilizes.