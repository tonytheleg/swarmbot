package helpers

import (
	"context"
	"fmt"
	"os"

	"github.com/PagerDuty/go-pagerduty"
)

// PDHelper helps configure a client and settings needed to interact with PagerDuty
type PDHelper struct {
	Client           *pagerduty.Client
	User             *pagerduty.User
	ScheduleID       string
	EscalationPolicy string
	BaseUrl          string
	Pressure         int
}

// GetPDClient creates a PagerDuty client for interacting with the API
func (p *PDHelper) GetPDClient() {
	authtoken := os.Getenv("PAGERDUTY_TOKEN")
	p.Client = pagerduty.NewClient(authtoken)
}

// GetPDUser gets a user from PagerDuty
func (p *PDHelper) GetPDUser(userEmail string) error {
	pdUser, err := p.Client.ListUsersWithContext(context.Background(), pagerduty.ListUsersOptions{Query: userEmail})
	if err != nil {
		return fmt.Errorf("error fetching user by email %s -- %v", userEmail, err)
	}
	p.User = &pdUser.Users[0]
	return nil
}

/* This sample determines the number of PD incidents based on whats assigned to the current
on-call user. This was to get around a limitation of PD's free trial not including things like teams.
Instead of getting incidents by on-call user, we could target a team and urgency level instead (high/low)
The ListPDIncs call would just be updated for the correct team and urgency values */

// GetPrimaryOnCall gets the current on-call user in PagerDuty to determining incident counts
func (p *PDHelper) GetPrimaryOnCall() error {
	primary, err := p.Client.ListOnCallsWithContext(context.Background(), pagerduty.ListOnCallOptions{
		EscalationPolicyIDs: []string{p.EscalationPolicy},
		ScheduleIDs:         []string{p.ScheduleID},
	})
	if err != nil {
		return fmt.Errorf("error fetching on-call user: %v", err)
	}
	p.User = &primary.OnCalls[0].User
	return nil
}

// GetPDIncs lists all incidents assigned to a user
func (p *PDHelper) GetPDIncs() ([]string, error) {
	var openIncs []string
	incs, err := p.Client.ListIncidentsWithContext(context.Background(), pagerduty.ListIncidentsOptions{UserIDs: []string{p.User.ID}})
	if err != nil {
		return nil, fmt.Errorf("failed to list incidents: %v", err)
	}
	for _, inc := range incs.Incidents {
		openIncs = append(openIncs, inc.ID)
	}
	return openIncs, nil
}

// AssignPDInc handles assigning a PagerDuty incident to a user
func (p *PDHelper) AssignPDInc(incidentID string) error {
	// takes incident ID if it exists and assigns to user
	options := pagerduty.ManageIncidentsOptions{
		ID:   incidentID,
		Type: "incident",
		Assignments: []pagerduty.Assignee{
			pagerduty.Assignee{pagerduty.APIObject{
				ID:   p.User.ID,
				Type: "user_reference"}},
		},
	}
	_, err := p.Client.ManageIncidentsWithContext(context.Background(), p.User.Email, []pagerduty.ManageIncidentsOptions{options})
	if err != nil {
		return fmt.Errorf("failed to assign the PD incident to user: %w", err)
	}
	return nil
}

// CheckPDPressure checks the number of incidents and compares it to the Pressure value.
// It initializes a swarm of the value is exceeded.
func (p *PDHelper) CheckPDPressure() (bool, error) {
	// get primary on-call
	err := p.GetPrimaryOnCall()
	if err != nil {
		return false, fmt.Errorf("failed to fetch primary on-call user: %w", err)
	}

	incs, err := p.GetPDIncs()
	if err != nil {
		return false, fmt.Errorf("failed to get PD incidents: %w", err)
	}

	if len(incs) > p.Pressure {
		return true, nil
	}
	return false, nil
}
