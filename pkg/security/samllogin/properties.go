package samllogin

const ServiceProviderPropertiesPrefix = "security.auth.saml"

type ServiceProviderProperties struct {
	RootUrl string `json:"root-url"`
	CertificateFile string `json:"certificate-file"`
	KeyFile string  `json:"key-file"`
	KeyPassword string `json:"key-password"`
}

func NewServiceProviderProperties() *ServiceProviderProperties {
	return &ServiceProviderProperties{}
}