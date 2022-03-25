package pqcrypt

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/dbtest"
	"embed"
	"fmt"
	"github.com/google/uuid"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"testing"
)

const (
	testModelNameV1PlainMap        = "v1_plain_map"
	testModelNameV2PlainMap        = "v2_plain_map"
	testModelNameV1MockedMap       = "v1_mock_map"
	testModelNameV2MockedMap       = "v2_mock_map"
	testModelNameV2InvalidPlainMap = "v2_invalid_plain_map"
	testModelNameV2InvalidVaultMap = "v2_invalid_mock_map"
)

/*************************
	Models
 *************************/

type EncryptedModel struct {
	ID    int    `gorm:"primaryKey;type:serial;"`
	Name  string `gorm:"uniqueIndex;not null;"`
	Value *EncryptedMap
}

func (EncryptedModel) TableName() string {
	return "data_encryption_test"
}

/*************************
	Test Cases
 *************************/

//func TestMain(m *testing.M) {
//	suitetest.RunTests(m,
//		dbtest.EnableDBRecordMode(),
//	)
//}

//go:embed testdata/*.sql
var testDataFS embed.FS

func SetupEncryptedMapTestPrepareTables(di *dbDI) test.SetupFunc {
	return dbtest.PrepareData(&di.DI,
		dbtest.SetupUsingSQLFile(testDataFS, "testdata/tables.sql"),
		dbtest.SetupTruncateTables("data_encryption_test"),
		dbtest.SetupUsingSQLFile(testDataFS, "testdata/data.sql"),
	)
}

type dbDI struct {
	fx.In
	dbtest.DI
}

func TestEncryptedMapWithEncryptionEnabled(t *testing.T) {
	v := map[string]interface{}{
		"key1": "value1",
		"key2": 2.0,
	}
	di := dbDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		dbtest.WithDBPlayback("testdb"),
		apptest.WithModules(Module),
		apptest.WithFxOptions(
			fx.Provide(newMockedEncryptor(true)),
		),
		apptest.WithDI(&di),
		test.SubTestSetup(SetupEncryptedMapTestPrepareTables(&di)),
		test.GomegaSubTest(SubTestMapSuccessfulSqlScan(&di, testModelNameV1PlainMap, expectMap(V1, AlgPlain, v)), "SuccessfulSqlScanWithV1PlainText"),
		test.GomegaSubTest(SubTestMapSuccessfulSqlScan(&di, testModelNameV2PlainMap, expectMap(V2, AlgPlain, v)), "SuccessfulSqlScanWithV2PlainText"),
		test.GomegaSubTest(SubTestMapSuccessfulSqlScan(&di, testModelNameV1MockedMap, expectMap(V1, AlgVault, v)), "SuccessfulSqlScanWithV1MockedVault"),
		test.GomegaSubTest(SubTestMapSuccessfulSqlScan(&di, testModelNameV2MockedMap, expectMap(V2, AlgVault, v)), "SuccessfulSqlScanWithV2MockedVault"),
		test.GomegaSubTest(SubTestMapSuccessfulSqlValue(&di, v, AlgVault), "SuccessfulSqlValue"),

		test.GomegaSubTest(SubTestMapFailedSqlScan(&di, testModelNameV2InvalidPlainMap), "FailedSqlScanWithV2PlainText"),
		test.GomegaSubTest(SubTestMapFailedSqlScan(&di, testModelNameV2InvalidVaultMap), "FailedSqlScanWithV2MockedVault"),
		test.GomegaSubTest(SubTestMapFailedSqlValue(&di), "FailedSqlValueWithInvalidKeyIDAndAlg"),
	)
}

func TestEncryptedMapWithEncryptionDisabled(t *testing.T) {
	v := map[string]interface{}{
		"key1": "value1",
		"key2": 2.0,
	}
	di := dbDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		dbtest.WithDBPlayback("testdb"),
		apptest.WithModules(Module),
		apptest.WithFxOptions(
			fx.Provide(newMockedEncryptor(false)),
		),
		apptest.WithDI(&di),
		test.SubTestSetup(SetupEncryptedMapTestPrepareTables(&di)),
		test.GomegaSubTest(SubTestMapSuccessfulSqlScan(&di, testModelNameV1PlainMap, expectMap(V1, AlgPlain, v)), "SuccessfulSqlScanWithV1PlainText"),
		test.GomegaSubTest(SubTestMapSuccessfulSqlScan(&di, testModelNameV2PlainMap, expectMap(V2, AlgPlain, v)), "SuccessfulSqlScanWithV2PlainText"),
		test.GomegaSubTest(SubTestMapSuccessfulSqlValue(&di, v, AlgPlain), "SuccessfulSqlValue"),

		test.GomegaSubTest(SubTestMapFailedSqlScan(&di, testModelNameV1MockedMap), "FailedSqlScanWithV1MockedVault"),
		test.GomegaSubTest(SubTestMapFailedSqlScan(&di, testModelNameV2MockedMap), "FailedSqlScanWithV2MockedVault"),
		test.GomegaSubTest(SubTestMapFailedSqlScan(&di, testModelNameV2InvalidPlainMap), "FailedSqlScanWithV2PlainText"),
		test.GomegaSubTest(SubTestMapFailedSqlScan(&di, testModelNameV2InvalidVaultMap), "FailedSqlScanWithV2MockedVault"),
		test.GomegaSubTest(SubTestMapFailedSqlValue(&di), "FailedSqlValueWithInvalidKeyIDAndAlg"),
	)
}

/*************************
	Sub-Test Cases
 *************************/

func SubTestMapSuccessfulSqlScan(di *dbDI, name string, expected *testSpecs) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		m := EncryptedModel{}
		r := di.DB.WithContext(ctx).
			Where("Name = ?", name).
			Take(&m)
		g.Expect(r.Error).To(Succeed(), "db select shouldn't return error")
		g.Expect(m.Value).To(Not(BeNil()), "encrypted field shouldn't be nil")
		g.Expect(m.Value.Ver).To(BeIdenticalTo(expected.ver), "encrypted field's data should have correct Ver")
		g.Expect(m.Value.KeyID).To(Not(Equal(uuid.UUID{})), "encrypted field's data should have valid KeyID")
		g.Expect(m.Value.Alg).To(Equal(expected.alg), "encrypted field's data should have correct Alg")
		if expected.data != nil {
			g.Expect(m.Value.Data).To(Equal(expected.data), "encrypted field's data should have correct Data")
		} else {
			g.Expect(m.Value.Data).To(BeNil(), "encrypted field's data should have correct Data")
		}
	}
}

func SubTestMapFailedSqlScan(di *dbDI, name string) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		m := EncryptedModel{}
		r := di.DB.WithContext(ctx).
			Where("Name = ?", name).
			Take(&m)
		g.Expect(r.Error).To(Not(Succeed()), "db select should return error")
	}
}

func SubTestMapSuccessfulSqlValue(di *dbDI, v map[string]interface{}, expectedAlg Algorithm) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		kid := uuid.MustParse("aa74a96c-c0f4-4a29-9c76-e643ff29dee8")
		m := EncryptedModel{
			ID: 12345678,
			Name:  fmt.Sprintf("temp_%s_%s", expectedAlg, utils.RandomString(8)),
			Value: NewEncryptedMap(kid, v),
		}

		r := di.DB.WithContext(ctx).Save(&m)
		g.Expect(r.Error).To(Succeed(), "db save shouldn't return error")
		defer func() {
			di.DB.Delete(&EncryptedModel{}, m.ID)
		}()
		g.Expect(m.Value.Ver).To(BeIdenticalTo(V2), "encrypted field's data Ver should be correct")

		// fetch back
		decrypted := EncryptedModel{}
		r = di.DB.WithContext(ctx).Take(&decrypted, m.ID)
		g.Expect(r.Error).To(Succeed(), "db select shouldn't return error")

		g.Expect(decrypted.Value).To(Not(BeNil()), "decrypted field shouldn't be nil")
		g.Expect(decrypted.Value.Ver).To(BeIdenticalTo(V2), "decrypted field's data should have correct Ver")
		g.Expect(decrypted.Value.KeyID).To(Equal(kid.String()), "decrypted field's data should have correct KeyID")
		g.Expect(decrypted.Value.Alg).To(Equal(expectedAlg), "decrypted field's data should have correct Alg")
		if v != nil {
			g.Expect(decrypted.Value.Data).To(Equal(v), "decrypted field's data should have correct Data")
		} else {
			g.Expect(decrypted.Value.Data).To(BeNil(), "decrypted field's data should have correct Data")
		}
	}
}

func SubTestMapFailedSqlValue(di *dbDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		m := EncryptedModel{
			Name:  fmt.Sprintf("temp_invalid_%s", utils.RandomString(8)),
			Value: NewEncryptedMap(uuid.UUID{}, nil),
		}
		r := di.DB.WithContext(ctx).Save(&m)
		g.Expect(r.Error).To(Not(Succeed()), "db select should return error")
	}
}

/*************************
	Helper
 *************************/

type testSpecs struct {
	ver Version
	alg Algorithm
	data interface{}
}

func expectMap(ver Version, alg Algorithm, data map[string]interface{}) *testSpecs {
	return &testSpecs{
		ver:  ver,
		alg:  alg,
		data: data,
	}
}

func newMockedEncryptor(enabled bool) func() Encryptor {
	if !enabled {
		return func() Encryptor {
			return plainTextEncryptor{}
		}
	}
	return func() Encryptor {
		return compositeEncryptor{
			newMockedVaultEncryptor(),
			plainTextEncryptor{},
		}
	}
}