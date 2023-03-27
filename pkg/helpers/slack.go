package helpers

import (
	"context"
	"fmt"

	"github.com/slack-go/slack"
)

type SlackHelper struct {
	Client    *slack.Client
	Username  string
	UserEmail string
}

// GetSlackUserEmail gets the email address of a user
func (s *SlackHelper) GetSlackUserEmail(userID string) error {
	userprofile, err := s.Client.GetUserProfileContext(context.Background(), &slack.GetUserProfileParameters{UserID: userID})
	if err != nil {
		return fmt.Errorf("failed to get slack user email: %w", err)
	}
	s.UserEmail = userprofile.Email
	return nil
}

// GetSlackUsername gets the slack handle for a user by email
func (s *SlackHelper) GetSlackUsername(email string) error {
	user, err := s.Client.GetUserByEmailContext(context.Background(), email)
	if err != nil {
		return fmt.Errorf("failed to get slack user id: %w", err)
	}
	s.Username = user.Name
	return nil
}
