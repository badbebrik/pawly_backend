package handlers

import (
	"context"
	"strings"
	"time"

	"profile/internal/application/ports"
	profileuc "profile/internal/application/usecase"
	"profile/internal/domain/model"
	"profile/internal/transport/http/dto"
)

func (h *Handlers) profileToResponse(ctx context.Context, profile *model.Profile) dto.ProfileResponse {
	var avatarURL *string
	if profile.AvatarFileID != nil {
		if url, err := h.useCases.GetAvatarDownloadURL.Execute(ctx, *profile.AvatarFileID); err == nil {
			avatarURL = &url
		}
	}

	return dto.ProfileResponse{
		UserID:            profile.UserID,
		FirstName:         profile.FirstName,
		LastName:          profile.LastName,
		AvatarDownloadURL: avatarURL,
		Locale:            profile.Locale,
		Timezone:          profile.Timezone,
		CreatedAt:         profile.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:         profile.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func uploadInfoToResponse(upload ports.UploadInfo) dto.UploadInfoResponse {
	return dto.UploadInfoResponse{
		Method:    upload.Method,
		URL:       upload.URL,
		Headers:   upload.Headers,
		ExpiresAt: upload.ExpiresAt.UTC().Format(time.RFC3339),
	}
}

func profileBriefToResponse(item profileuc.ProfileBrief) dto.ProfileBriefResponse {
	return dto.ProfileBriefResponse{
		UserID:            item.UserID,
		FirstName:         item.FirstName,
		LastName:          item.LastName,
		DisplayName:       buildDisplayName(item.FirstName, item.LastName),
		AvatarDownloadURL: item.AvatarDownloadURL,
	}
}

func valueOrEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func buildDisplayName(firstName, lastName *string) *string {
	displayName := strings.TrimSpace(strings.Join([]string{valueOrEmpty(firstName), valueOrEmpty(lastName)}, " "))
	if displayName == "" {
		return nil
	}
	return &displayName
}
