package passwd

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/redis"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
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
	// New create new OTP and save it
	New() (OTP, error)

	// Get loads OTP by ID
	Get(id string) (OTP, error)

	// Verify use Get to load OTP and check the given passcode against the loaded OTP.
	// It returns the loaded OTP regardless the verification result.
	// It returns false if it reaches maximum attempts limit. otherwise returns true
	// error parameter indicate wether the given passcode is valid. It's nil if it's valid
	Verify(id, passcode string) (loaded OTP, hasMoreChances bool, err error)

	// Refresh regenerate OTP passcode without changing secret and ID
	// It returns the loaded or refreshed OTP regardless the verification result.
	// It returns false if it reaches maximum attempts limit. otherwise returns true
	// error parameter indicate wether the passcode is refreshed
	Refresh(id string) (refreshed OTP, hasMoreChances bool, err error)

	// Delete delete OTP by ID
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

func (os *redisTotpStore) Get(id string) (OTP, error) {
	otp, err := os.load(id);
	if err != nil {
		return nil, err
	}
	return otp, nil
}

func (os *redisTotpStore) Verify(id, passcode string) (loaded OTP, hasMoreChances bool, err error) {
	// load OTP by ID
	otp, e := os.load(id);
	if otp == nil || e != nil {
		return nil, false, security.NewCredentialsExpiredError("Passcode already expired", e)
	}

	// schedule for post verification
	defer os.cleanup(otp)

	// check verification attempts
	if otp.IncrementAttempts(); otp.Attempts() >= os.maxVerifyLimit {
		return nil, false, security.NewMaxAttemptsReachedError("Max verification attempts exceeded")
	}

	toValidate := TOTP{
		Passcode: passcode,
		Secret:   otp.secret(),
		TTL:      otp.TTL(),
		Expire:   time.Now().Add(otp.TTL()),
	}

	loaded = otp
	hasMoreChances = otp.Attempts() < os.maxVerifyLimit
	if valid, e := os.factory.Validate(toValidate); e != nil || !valid {
		err = security.NewBadCredentialsError("Passcode doesn't match", e)
	}
	return
}

func (os *redisTotpStore) Refresh(id string) (loaded OTP, hasMoreChances bool, err error) {
	// load OTP by ID
	otp, e := os.load(id);
	if e != nil {
		return nil, false, security.NewCredentialsExpiredError("Max refresh/resend attempts exceeded", e)
	}

	// schedule for post refresh
	loaded = otp
	defer os.cleanup(otp)

	// check refresh attempts
	if otp.IncrementRefreshes(); otp.Refreshes() >= os.maxRefreshLimit {
		return loaded, false, security.NewMaxAttemptsReachedError("Max refresh/resend attempts exceeded")
	}

	// calculate remining time
	ttl := otp.Expire().Sub(time.Now())
	if ttl <= 0 {
		return loaded, false, security.NewCredentialsExpiredError("Passcode already expired")
	}

	// do refresh
	hasMoreChances = otp.Refreshes() < os.maxRefreshLimit
	refreshed, e := os.factory.Refresh(otp.secret(), ttl)
	if e != nil {
		return loaded, hasMoreChances, security.NewAuthenticationError("Unabled to refresh/resend passcode", e)
	}
	otp.Value = refreshed
	return
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