package main

import (
	"context"
	"github.com/google/uuid"
	"time"
)

type OTP struct {
	Key     string
	Created time.Time
}

type RetentionMap map[string]OTP

func NewRetentionMap(ctx context.Context, retentionPeriod time.Duration) RetentionMap {

	rm := make(RetentionMap)

	go rm.Retention(ctx, retentionPeriod)

	return rm

}
func (rm RetentionMap) NewOtp() OTP {
	otp := OTP{
		Key:     uuid.NewString(),
		Created: time.Now(),
	}
	rm[otp.Key] = otp
	return otp

}

func (rm RetentionMap) VerifyOtp(otp string) bool {

	if _, ok := rm[otp]; !ok {
		return false
	}
	delete(rm, otp)
	return true

}

func (rm RetentionMap) Retention(ctx context.Context, retentionPeriod time.Duration) {
	ticker := time.NewTicker(400 * time.Millisecond)

	select {
	case <-ticker.C:
		for _, otp := range rm {
			if otp.Created.Add(retentionPeriod).Before(time.Now()) {
				delete(rm, otp.Key)

			}

		}
	case <-ctx.Done():
		return

	}
}
