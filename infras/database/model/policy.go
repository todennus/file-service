package model

import (
	"time"

	"github.com/todennus/file-service/domain"
	"github.com/xybor-x/snowflake"
)

type UploadPolicy struct {
	UserID       int64    `json:"uid"`
	AllowedTypes []string `json:"ats"`
	MaxSize      int64    `json:"msz"`
	ExpiresAt    int64    `json:"exp"`
}

func NewUploadPolicy(policy *domain.UploadPolicy) *UploadPolicy {
	return &UploadPolicy{
		UserID:       policy.UserID.Int64(),
		AllowedTypes: policy.AllowedTypes,
		MaxSize:      policy.MaxSize,
		ExpiresAt:    policy.ExpiresAt.Unix(),
	}
}

func (policy *UploadPolicy) To(token string) *domain.UploadPolicy {
	return &domain.UploadPolicy{
		Token:        token,
		UserID:       snowflake.ParseInt64(policy.UserID),
		AllowedTypes: policy.AllowedTypes,
		MaxSize:      policy.MaxSize,
		ExpiresAt:    time.Unix(policy.ExpiresAt, 0),
	}
}
