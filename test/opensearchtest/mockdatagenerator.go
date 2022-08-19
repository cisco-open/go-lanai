package opensearchtest

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/opensearch"
	"math/rand"
	"strings"
	"time"
)

type GenericAuditEvent struct {
	Client_ID       string
	Description     string
	Details         string
	ID              string
	Keywords        string
	Orig_User       string
	Owner_Tenant_ID string
	Parent_Span_ID  string
	Provider_ID     string
	Security        string
	Service         string
	Severity        string
	Span_ID         string
	SubType         string
	Tenant_ID       string
	Tenant_Name     string
	Time            time.Time
	Time_Bucket     int
	Trace           string
	Trace_ID        string
	Type            string
	User_ID         string
	Username        string
}

func SetupPrepareOpenSearchData(
	ctx context.Context,
	repo opensearch.Repo[GenericAuditEvent],
	startDate time.Time,
	endDate time.Time,
) (context.Context, error) {
	// We don't care if we can't delete this indices - it might not exist
	//nolint:errcheck
	repo.IndicesDelete(context.Background(), "auditlog")
	events := []GenericAuditEvent{}
	CreateData(10, startDate, endDate, &events)
	for _, event := range events {
		err := repo.Index(
			context.Background(),
			"auditlog",
			event,
			opensearch.Index.WithRefresh("true"),
			opensearch.Index.WithWaitForActiveShards("all"),
		)
		if err != nil {
			return ctx, err
		}
	}
	return ctx, nil
}

// CreateData will create a slice of random generated GenericAuditEvents
// The time between each event will be uniformly distributed between the startT and endT
func CreateData(numOfDocuments int, startT time.Time, endT time.Time, dest *[]GenericAuditEvent) {
	timeDelta := endT.Sub(startT) / time.Duration(numOfDocuments)
	genericEvents := make([]GenericAuditEvent, numOfDocuments)
	for i := 0; i < numOfDocuments; i++ {
		PopulateSourceWithDeterministicData(&genericEvents[i])
		genericEvents[i].Time = startT.Add(timeDelta)
	}
	*dest = genericEvents
}

func PopulateSourceWithDeterministicData(source *GenericAuditEvent) {
	subTypes := []string{"W", "SCHEDULE_TASK", "SYNCHRONIZED"}
	Types := []string{"GP", "DEVICE", "DP"}

	source.Type = Types[int(src.Int63())%len(Types)]
	source.SubType = subTypes[int(src.Int63())%len(subTypes)]
	source.Trace_ID = RandStringBytesMaskImprSrcSB(5)
	source.Span_ID = RandStringBytesMaskImprSrcSB(5)
	source.Parent_Span_ID = RandStringBytesMaskImprSrcSB(5)
	source.Client_ID = RandStringBytesMaskImprSrcSB(5)
	source.Tenant_ID = RandStringBytesMaskImprSrcSB(5)
	source.Provider_ID = RandStringBytesMaskImprSrcSB(5)
	source.Owner_Tenant_ID = RandStringBytesMaskImprSrcSB(5)
	source.User_ID = RandStringBytesMaskImprSrcSB(5)
	source.Orig_User = RandStringBytesMaskImprSrcSB(5)
	source.Username = RandStringBytesMaskImprSrcSB(5)
	source.Keywords = RandStringBytesMaskImprSrcSB(5)
}

//const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const letterBytes = "abcdefghij" // limiting the combination of characters
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

// We don't want random data that changes from run to run
var src = rand.NewSource(4242)

// RandStringBytesMaskImprSrcSB from https://stackoverflow.com/a/31832326
func RandStringBytesMaskImprSrcSB(n int) string {
	sb := strings.Builder{}
	sb.Grow(n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			sb.WriteByte(letterBytes[idx])
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return sb.String()
}
