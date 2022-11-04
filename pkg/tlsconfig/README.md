# TLS Config

The TLS Config package provides an easy way for any clients that takes a ```tlsconfig.Provider```to configure itself for TLS

The TLs Config package does this by enabling the following properties

```go
type Properties struct {
	// type can be vault or file
	Type string `json:"type"`
	MinVersion string `json:"min-version"`

	// vault type related properties
	Path             string `json:"path"`
	Role             string `json:"role"`
	CN               string `json:"cn"`
	IpSans           string `json:"ip-sans"`
	AltNames         string `json:"alt-names"`
	Ttl              string `json:"ttl"`
	MinRenewInterval string `json:"min-renew-interval"`

	// file type related properties
	CaCertFile string `json:"ca-cert-file"`
	CertFile string `json:"cert-file"`
	KeyFile string `json:"key-file"`
	KeyPass string `json:"key-pass"`
}
```
You can use add this struct to any properties you define in the application, and use the ```ProviderFactory``` to get a 
```tlsconfig.Provider``` from the properties.
