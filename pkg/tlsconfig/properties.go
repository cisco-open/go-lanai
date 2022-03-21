package tlsconfig

type Properties struct {
	// type can be vault or file
	Type string `json:"type"`

	// vault Related properties
	Path string `json:"path"`
	Role string `json:"role"`
	CN   string `json:"cn"`
	IpSans string `json:"ip-sans"`
	AltNames string `json:"alt-names"`
	Ttl string `json:"ttl"`
	MaxRenewFrequency string `json:"max-renew-frequency"`


	CaCertFile string `json:"ca-cert-file"`
	CertFile string `json:"cert-file"`
	KeyFile string `json:"key-file"`
	KeyPass string `json:"key-pass"`
}
