package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/umekikazuya/gh-grass/internal/domain"
)

// MockRepository is a manual mock for domain.GrassRepository
type MockRepository struct {
	Users         map[string]*domain.User
	OrgMembers    map[string][]domain.User
	Contributions map[string]map[string]int // username -> date(YYYY-MM-DD) -> count
	Err           error
}

func (m *MockRepository) GetUser(ctx context.Context, login string) (*domain.User, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	if login == "" {
		// Mock behavior for "Self"
		return &domain.User{Login: "myself"}, nil
	}
	if user, ok := m.Users[login]; ok {
		return user, nil
	}
	return nil, errors.New("user not found")
}

func (m *MockRepository) ListOrgMembers(ctx context.Context, orgName string) ([]domain.User, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	if members, ok := m.OrgMembers[orgName]; ok {
		return members, nil
	}
	return nil, errors.New("org not found")
}

func (m *MockRepository) GetContributions(ctx context.Context, username string, date time.Time) (*domain.Contribution, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	dateStr := date.Format("2006-01-02")
	if userContribs, ok := m.Contributions[username]; ok {
		if count, ok := userContribs[dateStr]; ok {
			return &domain.Contribution{Date: date, Count: count}, nil
		}
	}
	return &domain.Contribution{Date: date, Count: 0}, nil
}

func TestGetContributionCount(t *testing.T) {
	mockRepo := &MockRepository{
		Contributions: map[string]map[string]int{
			"testuser": {
				"2023-10-01": 5,
			},
		},
	}
	usecase := NewGrassUsecase(mockRepo)

	date, err := time.Parse("2006-01-02", "2023-10-01")
	if err != nil {
		t.Errorf("failed to parse date: %v", err)
	}

	count, err := usecase.GetContributionCount(context.Background(), "testuser", date)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if count != 5 {
		t.Errorf("expected 5 contributions, got %d", count)
	}
}

func TestGetSelf(t *testing.T) {
	mockRepo := &MockRepository{}
	usecase := NewGrassUsecase(mockRepo)

	user, err := usecase.GetSelf(context.Background())
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if user.Login != "myself" {
		t.Errorf("expected login 'myself', got %s", user.Login)
	}
}

func TestListOrganizationMembers(t *testing.T) {
	mockRepo := &MockRepository{
		OrgMembers: map[string][]domain.User{
			"myorg": {{Login: "member1"}, {Login: "member2"}},
		},
	}
	usecase := NewGrassUsecase(mockRepo)

	members, err := usecase.ListOrganizationMembers(context.Background(), "myorg")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if len(members) != 2 {
		t.Errorf("expected 2 members, got %d", len(members))
	}
}

func TestGetContributionCount_Error(t *testing.T) {
	expectedErr := errors.New("repository error")
	mockRepo := &MockRepository{
		Err: expectedErr,
	}
	usecase := NewGrassUsecase(mockRepo)

	date, err := time.Parse("2006-01-02", "2023-10-01")
	if err != nil {
		t.Errorf("failed to parse date: %v", err)
	}
	_, err := usecase.GetContributionCount(context.Background(), "testuser", date)
	if err == nil {
		t.Error("expected error, got nil")
	}
	if err != expectedErr {
		t.Errorf("expected %v, got %v", expectedErr, err)
	}
}

func TestGetSelf_Error(t *testing.T) {
	expectedErr := errors.New("repository error")
	mockRepo := &MockRepository{
		Err: expectedErr,
	}
	usecase := NewGrassUsecase(mockRepo)

	_, err := usecase.GetSelf(context.Background())
	if err == nil {
		t.Error("expected error, got nil")
	}
}
