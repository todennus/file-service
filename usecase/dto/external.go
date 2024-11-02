package dto

import "github.com/todennus/file-service/domain"

type OverridenPolicyInfo struct {
	*domain.UploadPolicy
	PolicySourceMetadata string
}

func OverridePolicyInfo(override *OverridenPolicyInfo, policy *domain.UploadPolicy) {
	if len(override.AllowedTypes) > 0 {
		policy.AllowedTypes = override.AllowedTypes
	}

	if override.MaxSize > 0 {
		policy.MaxSize = override.MaxSize
	}
}

type StorageUploadFileMetadata struct {
	Size int
	Hash string
}

type StorageDownloadFileMetadata struct {
	Size int
	Type string
	Hash string
}
