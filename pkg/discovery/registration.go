package discovery

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/consul"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/cryptoutils"
	netutil "cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/net"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"fmt"
	kitconsul "github.com/go-kit/kit/sd/consul"
	"github.com/google/uuid"
	"github.com/hashicorp/consul/api"
	"go.uber.org/fx"
	"strings"
)

var logger = log.New("Discovery")

func Register(ctx context.Context, connection *consul.Connection, registration *api.AgentServiceRegistration) {
	registrar := kitconsul.NewRegistrar(kitconsul.NewClient(connection.Client()),
		registration,
		logger.WithContext(ctx).WithLevel(log.LevelInfo).WithKV(log.LogKeyMessage, "Register"))
	registrar.Register()
}

func Deregister(ctx context.Context, connection *consul.Connection, registration *api.AgentServiceRegistration) {
	registrar := kitconsul.NewRegistrar(kitconsul.NewClient(connection.Client()),
		registration,
		logger.WithContext(ctx).WithLevel(log.LevelInfo).WithKV(log.LogKeyMessage, "Deregister"))
	registrar.Deregister()
}

type regDI struct {
	fx.In
	AppContext          *bootstrap.ApplicationContext
	DiscoveryProperties DiscoveryProperties
	ServerProperties    web.ServerProperties `optional:"true"`
}

func NewRegistration(di regDI) *api.AgentServiceRegistration {
	var ipAddress string

	if di.DiscoveryProperties.IpAddress != "" {
		ipAddress = di.DiscoveryProperties.IpAddress
	} else {
		ipAddress, _ = netutil.GetIp(di.DiscoveryProperties.Interface)
	}

	appName := di.AppContext.Name()
	registration := &api.AgentServiceRegistration{
		Kind:    api.ServiceKindTypical,
		ID:      fmt.Sprintf("%s-%d-%x", appName, di.DiscoveryProperties.Port, cryptoutils.RandomBytes(5)),
		Name:    appName,
		Tags:    createTags(di.DiscoveryProperties, di.ServerProperties),
		Port:    di.DiscoveryProperties.Port,
		Address: ipAddress,
		Check: &api.AgentServiceCheck{
			HTTP:                           fmt.Sprintf("%s://%s:%d%s", di.DiscoveryProperties.Scheme, ipAddress, di.DiscoveryProperties.Port, di.DiscoveryProperties.HealthCheckPath),
			Interval:                       di.DiscoveryProperties.HealthCheckInterval,
			DeregisterCriticalServiceAfter: di.DiscoveryProperties.HealthCheckCriticalTimeout},
	}
	return registration
}

func createTags(discoveryProperties DiscoveryProperties, serverProperties web.ServerProperties) []string {
	tags := make([]string, len(discoveryProperties.Tags))
	copy(tags, discoveryProperties.Tags)
	tags = append(tags, fmt.Sprintf("secure=%t", discoveryProperties.Scheme == "https"),
		fmt.Sprintf("contextPath=%s", serverProperties.ContextPath))
	return tags
}

var COMPONENT_ATTRIBUTES_MAPPING = map[string]string{
	"serviceName": "application.name",
	"context":     "server.context-path",
	"name":        "info.app.attributes.displayName",
	"description": "info.app.description",
	"parent":      "info.app.attributes.parent",
	"type":        "info.app.attributes.type",
}

type DefaultCustomizer struct {
	instanceUuid        uuid.UUID
	componentAttributes map[string]string
}

func NewDefaultCustomizer(appContext *bootstrap.ApplicationContext) *DefaultCustomizer {
	return &DefaultCustomizer{
		instanceUuid:        uuid.New(),
		componentAttributes: getComponentAttributes(appContext),
	}
}

func (d *DefaultCustomizer) Customize(ctx context.Context, registration *api.AgentServiceRegistration) {
	//The tag was extracted by admin services and injected into DB for UI to list all components
	registration.Tags = append(registration.Tags, kvTag(TAG_INSTANCE_ID, d.instanceUuid.String()), TAG_MANAGED_SERVICE)

	if registration.Meta == nil {
		registration.Meta = make(map[string]string)
	}
	registration.Meta[TAG_INSTANCE_ID] = d.instanceUuid.String()

	var attributeStrings []string
	for k, v := range d.componentAttributes {
		registration.Meta[k] = v
		attributeStrings = append(attributeStrings, fmt.Sprintf("%s:%s", k, v))
	}
	registration.Tags = append(registration.Tags, kvTag(TAG_COMPONENT_ATTRIBUTES, strings.Join(attributeStrings, COMPONENT_ATTRIBUTE_DELIMITER)))
}

func kvTag(k string, v string) string {
	return fmt.Sprintf("%s=%s", k, v)
}

func getComponentAttributes(appContext *bootstrap.ApplicationContext) map[string]string {
	attributes := make(map[string]string)
	for k, v := range COMPONENT_ATTRIBUTES_MAPPING {
		attributes[k] = fmt.Sprintf("%v", appContext.Value(v))
	}
	return attributes
}
