package helpers

import (
	"fmt"
	"os"
	"strings"

	"github.com/slack-go/slack"
)

type Helper struct {
	PD    PDHelper
	Jira  JiraHelper
	Slack SlackHelper
}

func NewHelper() Helper {
	var helper Helper
	token := os.Getenv("SLACK_AUTH_TOKEN")
	appToken := os.Getenv("SLACK_APP_TOKEN")

	// setup constants
	helper.Jira.BaseUrl = "http://localhost:8080/"
	helper.Jira.Pressure = 3
	helper.Jira.JqlQuery = "project = 'OpenShift Hosted Support' AND status = 'To Do' OR status = 'In Progress' AND priority >= 'High'"
	helper.PD.ScheduleID = "P5LOJUX"
	helper.PD.EscalationPolicy = "PA9G4O0"
	helper.PD.BaseUrl = "https://pdotest.pagerduty.com/incidents/"
	helper.PD.Pressure = 3

	// setup clients
	helper.Slack.Client = slack.New(token, slack.OptionDebug(true), slack.OptionAppLevelToken(appToken))
	helper.PD.GetPDClient()
	helper.Jira.GetJiraClient()

	return helper
}

func (h *Helper) GetAllIncidents() ([]string, []string, error) {
	incs, err := h.PD.GetPDIncs()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get PD incidents: %v", err)
	}
	issues, err := h.Jira.GetJiraIssues()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get Jira issues: %v", err)
	}
	return incs, issues, nil
}

// FormatIncList pretty prints the incident list with clickable links
func FormatIncList(incs []string, baseurl string) string {
	var incLinks []string

	for _, inc := range incs {
		incLinks = append(incLinks, baseurl+inc)
	}
	incList := strings.Join(incLinks[:], "\n\n")
	return incList
}
