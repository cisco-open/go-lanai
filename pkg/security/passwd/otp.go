package passwd

import (
	"crypto/rand"
	"encoding/base32"
	"fmt"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"time"
)

var b32NoPadding = base32.StdEncoding.WithPadding(base32.NoPadding)

type TOTP struct {
	Passcode string
	Secret string
	TTL time.Duration
	Expire time.Time
}

type TOTPFactory interface {
	Generate(ttl time.Duration) (totp TOTP, err error)
	Refresh(secret string, ttl time.Duration) (totp TOTP, err error)
	Validate(totp TOTP) (valid bool, err error)
}

type totpFactory struct {
	skew       uint
	digits     otp.Digits
	alg        otp.Algorithm
	secretSize int
}

func (f *totpFactory) Generate(ttl time.Duration) (ret TOTP, err error) {
	secret, err := f.generateSecret()
	if err != nil {
		return
	}

	return f.Refresh(secret, ttl)
}

func (f *totpFactory) Refresh(secret string, ttl time.Duration) (ret TOTP, err error) {
	if ttl < time.Second {
		return ret, fmt.Errorf("ttl should be greater or equals to 1 seconds")
	}

	now := time.Now()
	ttl = ttl.Round(time.Second)
	passcode, err := totp.GenerateCodeCustom(secret, now, totp.ValidateOpts{
		Period:    uint(ttl.Seconds()),
		Skew:      f.skew,
		Digits:    f.digits,
		Algorithm: f.alg,
	})
	if err != nil {
		return
	}

	ret = TOTP{
		Passcode: passcode,
		Secret: secret,
		TTL: ttl,
		Expire: now.Add(ttl),
	}
	return
}

func (f *totpFactory) Validate(value TOTP) (valid bool, err error) {
	if value.TTL < time.Second {
		return false, fmt.Errorf("ttl should be greater or equals to 1 seconds")
	}

	now := time.Now()
	return totp.ValidateCustom(value.Passcode, value.Secret, now, totp.ValidateOpts{
		Period:    uint(value.TTL.Round(time.Second).Seconds()),
		Skew:      f.skew,
		Digits:    f.digits,
		Algorithm: f.alg,
	})
}

func (f *totpFactory) generateSecret() (string, error) {
	secret := make([]byte, f.secretSize)

	_, err := rand.Reader.Read(secret)
	if err != nil {
		return "", err
	}
	return b32NoPadding.EncodeToString(secret), nil
}
