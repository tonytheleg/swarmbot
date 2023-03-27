package helpers

import (
	"context"
	"fmt"
	"log"
	"os"

	jira "github.com/andygrunwald/go-jira"
)

// JiraHelper helps configure a client and settings needed to interact with Jira
type JiraHelper struct {
	Client   *jira.Client
	User     *jira.User
	BaseUrl  string
	Pressure int
	JqlQuery string
}

// GetJiraClient setups a Jira client
func (j *JiraHelper) GetJiraClient() {
	token := os.Getenv("JIRA_PAT_TOKEN")
	tp := jira.PATAuthTransport{
		Token: token,
	}
	j.Client, _ = jira.NewClient(tp.Client(), j.BaseUrl)
}

// GetJiraUser finds a Jira user using the provided user's email
func (j *JiraHelper) GetJiraUser(email string) error {
	user, _, err := j.Client.User.Find("username", jira.WithUsername(email))
	if err != nil {
		return fmt.Errorf("failed to find a Jira user with email %s: %v", email, err)
	}
	j.User = &user[0]
	return nil
}

// GetJiraIssues returns all issues based on the JQL query defined in the Helper
func (j *JiraHelper) GetJiraIssues() ([]string, error) {
	var openIssues []string
	opt := &jira.SearchOptions{
		MaxResults: 100,
		StartAt:    0,
	}

	issues, _, err := j.Client.Issue.Search(j.JqlQuery, opt)
	if err != nil {
		return nil, fmt.Errorf("failed to get Jira issues: %v", err)
	}

	for _, issue := range issues {
		openIssues = append(openIssues, issue.Key)
	}
	return openIssues, nil
}

// AssignJiraIssue assigns an issues to the helpers defined user
func (j *JiraHelper) AssignJiraIssue(issueKey string) error {
	resp, err := j.Client.Issue.UpdateAssigneeWithContext(context.Background(), issueKey, j.User)
	if err != nil {
		return fmt.Errorf("failed to update the assignee on issue %s: %v", issueKey, err)
	}
	if resp.StatusCode >= 400 {
		log.Printf("non 200 status code")
		return fmt.Errorf("received a non 200 status code: %d: %v", resp.StatusCode, err)
	}
	return nil
}

// CheckJiraPressure checks the number of issues and compares it to the Pressure value.
// It initializes a swarm of the value is exceeded.
func (j *JiraHelper) CheckJiraPressure() (bool, error) {
	issues, err := j.GetJiraIssues()
	if err != nil {
		return false, fmt.Errorf("failed to get Jira issues: %w", err)
	}

	if len(issues) > j.Pressure {
		return true, nil
	}
	return false, nil
}
