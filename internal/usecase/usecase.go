package usecase

import (
	"context"
	"fmt"
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
		return 0, fmt.Errorf("get contributions for %q on %s: %w", username, date.Format("2006-01-02"), err)
	}
	return contrib.Count, nil
}

// ListOrganizationMembers returns a list of members for an organization.
func (u *GrassUsecase) ListOrganizationMembers(ctx context.Context, orgName string) ([]domain.User, error) {
	members, err := u.repo.ListOrgMembers(ctx, orgName)
	if err != nil {
		return nil, fmt.Errorf("list organization members for %q: %w", orgName, err)
	}
	return members, nil
}

// GetSelf returns the authenticated user.
func (u *GrassUsecase) GetSelf(ctx context.Context) (*domain.User, error) {
	// Assuming empty string implies "self" for the repository
	user, err := u.repo.GetUser(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("get authenticated user: %w", err)
	}
	return user, nil
}

// GetContributionCalendar は until を終点とする weeks 週間分のカレンダーを返す。
func (u *GrassUsecase) GetContributionCalendar(ctx context.Context, username string, until time.Time, weeks int) (*domain.ContributionCalendar, error) {
	from := until.AddDate(0, 0, -(weeks*7 - 1))
	calendar, err := u.repo.GetContributionCalendar(ctx, username, from, until)
	if err != nil {
		return nil, fmt.Errorf("get contribution calendar for %q (%s to %s): %w", username, from.Format("2006-01-02"), until.Format("2006-01-02"), err)
	}
	return calendar, nil
}
