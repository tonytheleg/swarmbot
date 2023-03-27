package helpers

import (
	"fmt"
	"os"
	"strings"

	"github.com/slack-go/slack"
)

const JiraBaseUrl string = "http://localhost:8080/"
const JiraPressure int = 3
const JiraJqlQuery string = "project = 'OpenShift Hosted Support' AND status = 'To Do' OR status = 'In Progress' AND priority >= 'High'"
const PDScheduleID string = "P5LOJUX"
const PDEscalationPolicy string = "PA9G4O0"
const PDBaseUrl string = "https://pdotest.pagerduty.com/incidents/"
const PDPressure int = 3

// Helper is a class-like struct to help quickly init other needed helpers and provide a single access point to values
type Helper struct {
	PD    PDHelper
	Jira  JiraHelper
	Slack SlackHelper
}

// NewHelper builds a new Helper and its underlying clients and settings
func NewHelper() Helper {
	var helper Helper
	token := os.Getenv("SLACK_AUTH_TOKEN")
	appToken := os.Getenv("SLACK_APP_TOKEN")

	// setup constants
	helper.Jira.BaseUrl = JiraBaseUrl
	helper.Jira.Pressure = JiraPressure
	helper.Jira.JqlQuery = JiraJqlQuery
	helper.PD.ScheduleID = PDScheduleID
	helper.PD.EscalationPolicy = PDEscalationPolicy
	helper.PD.BaseUrl = PDBaseUrl
	helper.PD.Pressure = PDPressure

	// setup clients
	helper.Slack.Client = slack.New(token, slack.OptionDebug(true), slack.OptionAppLevelToken(appToken))
	helper.PD.GetPDClient()
	helper.Jira.GetJiraClient()

	return helper
}

// GetAllIncidents fetches incidents from PagerDuty and Jira for listing and checking pressure
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
