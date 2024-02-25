// Copyright 2023 Cisco Systems, Inc. and its affiliates
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package passwd

import (
	"bytes"
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/redis"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"encoding/gob"
	"fmt"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"io"
	"strings"
	"time"
)

const (
	redisKeyPrefixOtp = "OTP-"
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

type OTPManager interface {
	// New create new OTP and save it
	New() (OTP, error)

	// Get loads OTP by Domain
	Get(id string) (OTP, error)

	// Verify use Get to load OTP and check the given passcode against the loaded OTP.
	// It returns the loaded OTP regardless the verification result.
	// It returns false if it reaches maximum attempts limit. otherwise returns true
	// error parameter indicate wether the given passcode is valid. It's nil if it's valid
	Verify(id, passcode string) (loaded OTP, hasMoreChances bool, err error)

	// Refresh regenerate OTP passcode without changing secret and Domain
	// It returns the loaded or refreshed OTP regardless the verification result.
	// It returns false if it reaches maximum attempts limit. otherwise returns true
	// error parameter indicate wether the passcode is refreshed
	Refresh(id string) (refreshed OTP, hasMoreChances bool, err error)

	// Delete delete OTP by Domain
	Delete(id string) error
}

type OTPStore interface {
	Save(OTP) error
	Load(id string) (OTP, error)
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
	v.AttemptCount++
}

func (v *timeBasedOtp) IncrementRefreshes() {
	v.RefreshCount++
}

// totpManager implements OTPManager
type totpManager struct {
	factory         TOTPFactory
	store           OTPStore
	ttl             time.Duration
	maxVerifyLimit  uint
	maxRefreshLimit uint
}

type totpManagerOptionsFunc func(*totpManager)

func newTotpManager(options ...totpManagerOptionsFunc) *totpManager {
	manager := &totpManager{
		store:           inmemOtpStore(make(map[string]OTP)),
		ttl:             time.Minute * 10,
		maxVerifyLimit:  3,
		maxRefreshLimit: 3,
	}

	for _, opt := range options {
		opt(manager)
	}
	return manager
}

func (m *totpManager) New() (OTP, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to create TOTP")
	}

	value, err := m.factory.Generate(m.ttl)
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to create TOTP")
	}

	otp := &timeBasedOtp{
		Identifier: id.String(),
		Value:      value,
	}

	// save
	if err := m.store.Save(otp); err != nil {
		return nil, err
	}
	return otp, nil
}

func (m *totpManager) Get(id string) (OTP, error) {
	otp, err := m.store.Load(id)
	if err != nil {
		return nil, err
	}
	return otp, nil
}

func (m *totpManager) Verify(id, passcode string) (loaded OTP, hasMoreChances bool, err error) {
	// load OTP by Domain
	otp, e := m.store.Load(id)
	if otp == nil || e != nil {
		return nil, false, security.NewCredentialsExpiredError("Passcode already expired", e)
	}

	// schedule for post verification
	defer m.cleanup(otp)

	// check verification attempts
	if otp.IncrementAttempts(); otp.Attempts() > m.maxVerifyLimit {
		return nil, false, security.NewMaxAttemptsReachedError("Max verification attempts exceeded")
	}

	toValidate := TOTP{
		Passcode: passcode,
		Secret:   otp.secret(),
		TTL:      otp.TTL(),
		Expire:   time.Now().Add(otp.TTL()),
	}

	loaded = otp
	hasMoreChances = otp.Attempts() < m.maxVerifyLimit
	if valid, e := m.factory.Validate(toValidate); e != nil || !valid {
		if hasMoreChances {
			err = security.NewBadCredentialsError("Passcode doesn't match", e)
		} else {
			err = security.NewMaxAttemptsReachedError("Passcode doesn't match and max verification attempts exceeded")
		}
	}
	return
}

func (m *totpManager) Refresh(id string) (loaded OTP, hasMoreChances bool, err error) {
	// load OTP by id
	loaded, e := m.store.Load(id)
	if e != nil {
		return nil, false, security.NewCredentialsExpiredError("Passcode expired", e)
	}

	otp, ok := loaded.(*timeBasedOtp)
	if !ok {
		return nil, false, security.NewCredentialsExpiredError("Passcode expired", e)
	}

	// schedule for post refresh
	defer m.cleanup(otp)

	// check refresh attempts
	if otp.IncrementRefreshes(); otp.Refreshes() > m.maxRefreshLimit {
		return loaded, false, security.NewMaxAttemptsReachedError("Max refresh/resend attempts exceeded")
	}

	// calculate remining time
	ttl := otp.Expire().Sub(time.Now())
	if ttl <= 0 {
		return loaded, false, security.NewCredentialsExpiredError("Passcode already expired")
	}

	// do refresh
	hasMoreChances = otp.Refreshes() < m.maxRefreshLimit
	refreshed, e := m.factory.Refresh(otp.secret(), ttl)
	if e != nil {
		if hasMoreChances {
			return loaded, hasMoreChances, security.NewAuthenticationError("Unable to refresh/resend passcode", e)
		} else {
			return loaded, hasMoreChances, security.NewMaxAttemptsReachedError("Unable to refresh/resend passcode and max refresh/resend attempts exceeded", e)
		}
	}
	otp.Value = refreshed
	return
}

func (m *totpManager) Delete(id string) error {
	return m.store.Delete(id)
}

func (m *totpManager) cleanup(otp OTP) {
	if time.Now().After(otp.Expire()) {
		// expired try to delete the record
		_ = m.store.Delete(otp.ID())
	} else {
		// not expired, save it
		_ = m.store.Save(otp)
	}
}

// inmemOtpStore implements OTPStore
type inmemOtpStore map[string]OTP

func (s inmemOtpStore) Save(otp OTP) error {
	s[otp.ID()] = otp
	return nil
}

func (s inmemOtpStore) Load(id string) (OTP, error) {
	if otp, ok := s[id]; ok {
		return otp, nil
	}
	return nil, fmt.Errorf("not found with id %s", id)
}

func (s inmemOtpStore) Delete(id string) error {
	if _, ok := s[id]; ok {
		delete(s, id)
		return nil
	}
	return fmt.Errorf("not found with id %s", id)
}

// redisOtpStore implements OTPStore
type redisOtpStore struct {
	redisClient redis.Client
}

func newRedisOtpStore(redisClient redis.Client) *redisOtpStore {
	return &redisOtpStore{
		redisClient: redisClient,
	}
}

func (s *redisOtpStore) Save(otp OTP) error {

	bytes, err := serialize(otp)
	if err != nil {
		return err
	}

	key := s.key(otp.ID())
	ttl := otp.Expire().Sub(time.Now())
	cmd := s.redisClient.Set(context.Background(), key, bytes, ttl)
	return cmd.Err()
}

func (s *redisOtpStore) Load(id string) (OTP, error) {
	key := s.key(id)
	cmd := s.redisClient.Get(context.Background(), key)
	val, err := cmd.Result()
	if err != nil {
		return nil, err
	}

	return deserialize(strings.NewReader(val))
}

func (s *redisOtpStore) Delete(id string) error {
	key := s.key(id)
	cmd := s.redisClient.Del(context.Background(), key)
	return cmd.Err()
}

func (s *redisOtpStore) key(id string) string {
	return redisKeyPrefixOtp + id
}

func serialize(otp OTP) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(&otp); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func deserialize(src io.Reader) (OTP, error) {
	dec := gob.NewDecoder(src)
	var otp OTP
	if err := dec.Decode(&otp); err != nil {
		return nil, err
	}
	return otp, nil
}
