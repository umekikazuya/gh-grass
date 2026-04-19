package domain

import (
	"context"
	"time"
)

// User はGithub Userを表現
type User struct {
	Login string
}

// Contribution は指定日のGithub Contributionの数を表現
type Contribution struct {
	Date  time.Time
	Count int
}

// ContributionDay は特定日のコントリビューション情報。
type ContributionDay struct {
	Date  time.Time
	Count int
}

// ContributionCalendar は週ごとに配列化されたコントリビューション一覧。
// Weeks[i] は通常7要素（日〜土、GitHub APIの返却順）だが、端の週は不完全な場合がある。
type ContributionCalendar struct {
	Weeks [][]ContributionDay
}

// GrassRepository はGithub上のユーザーデータとContribution数データを取得する抽象化インターフェース
type GrassRepository interface {
	// GetUser はユーザー情報を取得する
	GetUser(ctx context.Context, login string) (*User, error)

	// ListOrgMembers は入力値(第二引数)の組織名からユーザー情報をリストで取得する
	ListOrgMembers(ctx context.Context, orgName string) ([]User, error)

	// GetContributions は指定したユーザー・指定した日時のコントリビュート数を取得する
	GetContributions(ctx context.Context, username string, date time.Time) (*Contribution, error)

	// GetContributionCalendar は from〜to 期間のコントリビューションカレンダーを取得する
	GetContributionCalendar(ctx context.Context, username string, from, to time.Time) (*ContributionCalendar, error)
}
