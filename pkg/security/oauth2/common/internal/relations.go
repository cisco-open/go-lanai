package internal

type RelationToken struct {
	TokenKey  string `json:"token"`
}

type RelationAccessRefresh struct {
	RelationToken
	RefreshTokenKey string `json:"refresh"`
}

type RelationTokenSession struct {
	RelationToken
	SessionId string `json:"sid"`
}

type RelationTokenUserClient struct {
	RelationToken
	Username string `json:"user"`
	ClientId string `json:"cid"`
}
