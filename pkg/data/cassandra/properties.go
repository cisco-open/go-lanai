package cassandra

import (
	"fmt"
	"strings"
)

const (
	CassandraPropertiesPrefix = "data.cassandra"
)

type CassandraProperties struct {
	ContactPoints      string        `json:"contact-points"` // comma separated
	Port               int           `json:"port"`
	KeySpaceName	   string `json:"keyspace-name"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func (p CassandraProperties) Hosts() []string {
	hosts := strings.Split(p.ContactPoints, ",")
	for i, h := range hosts {
		hostParts := strings.SplitN(h, ":", 2)
		if len(hostParts) == 1 {
			hosts[i] = fmt.Sprintf("%s:%d", h, p.Port)
		}
	}
	return hosts
}

func NewCassandraProperties() *CassandraProperties{
	return &CassandraProperties{
		ContactPoints: "127.0.0.1",
		Port: 9042,
	}
}