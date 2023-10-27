package tlsconfig

type Properties struct {
	// type can be vault or file
	Type       string `json:"type"`
	MinVersion string `json:"min-version"`

	// vault type related properties
	Path             string    `json:"path"`
	Role             string    `json:"role"`
	CN               string    `json:"cn"`
	IpSans           string    `json:"ip-sans"`
	AltNames         string    `json:"alt-names"`
	Ttl              string    `json:"ttl"`
	MinRenewInterval string    `json:"min-renew-interval"`
	FileCache        FileCache `json:"file-cache"`

	// file type related properties
	CaCertFile string `json:"ca-cert-file"`
	CertFile   string `json:"cert-file"`
	KeyFile    string `json:"key-file"`
	KeyPass    string `json:"key-pass"`
}

type FileCache struct {
	Enabled bool   `json:"enabled"`
	Path    string `json:"path"`
	Prefix  string `json:"prefix"`
}
