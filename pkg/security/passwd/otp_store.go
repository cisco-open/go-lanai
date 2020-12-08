package passwd

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/redis"
	"fmt"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"time"
)

type OTP interface {
	ID() string
	Passcode() string
	TTL() time.Duration
	Expire() time.Time
	Attempts() uint
	Refreshes() uint
	IncrementAttempts()
	IncrementRefreshes()

	secret() string
}

type OTPStore interface {
	New() (OTP, error)
	Verify(id, passcode string) (OTP, error)
	Refresh(id string) (OTP, error)
	Delete(id string) error
}

/*****************************
	Common Implements
 *****************************/
// timeBasedOtp implements OTP
type timeBasedOtp struct {
	Identifier   string
	Value        TOTP
	AttemptCount uint
	RefreshCount uint
}

func (v *timeBasedOtp) secret() string {
	return v.Value.Secret
}

func (v *timeBasedOtp) ID() string {
	return v.Identifier
}

func (v *timeBasedOtp) Passcode() string {
	return v.Value.Passcode
}

func (v *timeBasedOtp) TTL() time.Duration {
	return v.Value.TTL
}

func (v *timeBasedOtp) Expire() time.Time {
	return v.Value.Expire
}

func (v *timeBasedOtp) Attempts() uint {
	return v.AttemptCount
}

func (v *timeBasedOtp) Refreshes() uint {
	return v.RefreshCount
}

func (v *timeBasedOtp) IncrementAttempts() {
	v.AttemptCount ++
}

func (v *timeBasedOtp) IncrementRefreshes() {
	v.RefreshCount ++
}

// redisTotpStore implements OTPStore
type redisTotpStore struct {
	redisClient     *redis.Connection
	factory         TOTPFactory
	ttl             time.Duration
	maxVerifyLimit  uint
	maxRefreshLimit uint
}

type redisTotpStoreOptions func(*redisTotpStore)

func newRedisTotpStore(redisClient *redis.Connection, options...redisTotpStoreOptions) *redisTotpStore {
	store := &redisTotpStore{
		redisClient: redisClient,
		factory: newTotpFactory(),
		ttl: time.Minute * 10,
		maxVerifyLimit: 3,
		maxRefreshLimit: 3,
	}

	for _,opt := range options {
		opt(store)
	}
	return store
}

func (os *redisTotpStore) New() (OTP, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to create TOTP")
	}

	value, err := os.factory.Generate(os.ttl)
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to create TOTP")
	}

	otp := &timeBasedOtp{
		Identifier: id.String(),
		Value: value,
	}

	// save
	if err := os.save(otp); err != nil {
		return nil, err
	}
	return otp, nil
}

func (os *redisTotpStore) Verify(id, passcode string) (OTP, error) {
	// load OTP by ID
	otp, err := os.load(id);
	if err != nil {
		return nil, errors.Wrapf(err, "Invalid passcode")
	}

	// schedule for post verification
	defer os.cleanup(otp)

	// check verification attempts
	if otp.IncrementAttempts(); otp.Attempts() >= os.maxVerifyLimit {
		return nil, fmt.Errorf("Max verification attempts exceeded")
	}

	toValidate := TOTP{
		Passcode: passcode,
		Secret:   otp.secret(),
		TTL:      otp.TTL(),
		Expire:   time.Now().Add(otp.TTL()),
	}

	if valid, err := os.factory.Validate(toValidate); err != nil {
		return nil, errors.Wrapf(err, "Invalid passcode")
	} else if !valid {
		return nil, fmt.Errorf("Invalid passcode")
	}
	return otp, nil
}

func (os *redisTotpStore) Refresh(id string) (OTP, error) {
	// load OTP by ID
	otp, err := os.load(id);
	if err != nil {
		return nil, errors.Wrapf(err, "No passcode available")
	}

	// schedule for post refresh
	defer os.cleanup(otp)

	// check refresh attempts
	if otp.IncrementRefreshes(); otp.Refreshes() >= os.maxRefreshLimit {
		return nil, fmt.Errorf("Max refresh/resend attempts exceeded")
	}

	// calculate remining time
	ttl := otp.Expire().Sub(time.Now())
	if ttl <= 0 {
		return nil, fmt.Errorf("Passcode already expired")
	}

	// do refresh
	refreshed, err := os.factory.Refresh(otp.secret(), ttl)
	if err != nil {
		return nil, errors.Wrapf(err, "Unabled to refresh/resend passcode")
	}
	otp.Value = refreshed
	return otp, nil
}

func (os *redisTotpStore) Delete(id string) error {
	return os.delete(id)
}

func (os *redisTotpStore) cleanup(otp *timeBasedOtp) {
	if time.Now().After(otp.Expire()) {
		// expired try to delete the record
		_ = os.delete(otp.ID())
	} else {
		// not expired, save it
		_ = os.save(otp)
	}
}

var devOtpMap map[string]*timeBasedOtp = make(map[string]*timeBasedOtp)

func (os *redisTotpStore) save(otp *timeBasedOtp) error {
	//TODO
	devOtpMap[otp.Identifier] = otp
	return nil
}

func (os *redisTotpStore) load(id string) (*timeBasedOtp, error) {
	//TODO
	if otp, ok := devOtpMap[id]; ok {
		return otp, nil
	}
	return nil, fmt.Errorf("not found with id %s", id)
}

func (os *redisTotpStore) delete(id string) error {
	//TODO
	if _, ok := devOtpMap[id]; ok {
		delete(devOtpMap, id)
		return nil
	}
	return fmt.Errorf("not found with id %s", id)
}