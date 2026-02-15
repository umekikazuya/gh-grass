package usecase

import (
	"context"
	"time"

	"github.com/umekikazuya/gh-grass/internal/domain"
)

type GrassUsecase struct {
	repo domain.GrassRepository
}

// NewGrassUsecase は、与えられた domain.GrassRepository を注入して初期化された *GrassUsecase を返します。
// repo はユースケースが利用するリポジトリ実装です。
func NewGrassUsecase(repo domain.GrassRepository) *GrassUsecase {
	return &GrassUsecase{repo: repo}
}

// GetContributionCount returns the contribution count for a user on a given date.
func (u *GrassUsecase) GetContributionCount(ctx context.Context, username string, date time.Time) (int, error) {
	contrib, err := u.repo.GetContributions(ctx, username, date)
	if err != nil {
		return 0, err
	}
	return contrib.Count, nil
}

// ListOrganizationMembers returns a list of members for an organization.
func (u *GrassUsecase) ListOrganizationMembers(ctx context.Context, orgName string) ([]domain.User, error) {
	return u.repo.ListOrgMembers(ctx, orgName)
}

// GetSelf returns the authenticated user.
func (u *GrassUsecase) GetSelf(ctx context.Context) (*domain.User, error) {
	// Assuming empty string implies "self" for the repository
	return u.repo.GetUser(ctx, "")
}