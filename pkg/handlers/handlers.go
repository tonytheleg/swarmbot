package handlers

import (
	"fmt"
	"strings"

	"github.com/slack-go/slack"
	"github.com/tonytheleg/swarmbot/pkg/helpers"
)

// HandleSlashCommand will take a slash command and route to the appropriate function
func HandleSlashCommand(command slack.SlashCommand, helper helpers.Helper) (interface{}, error) {
	// We need to switch depending on the command
	switch command.Command {
	case "/swarm-init":
		return nil, HandleSwarmInitCommand(command, helper)
	case "/assign":
		return nil, handleAssignCommand(command, helper)
	case "/list-incs":
		return nil, handleListCommand(command, helper)
	}
	return nil, nil
}

// HandleSwarmInitCommand begins the swarm process (for both manual and automatic)
func HandleSwarmInitCommand(command slack.SlashCommand, helper helpers.Helper) error {
	attachment := slack.Attachment{}

	attachment.Pretext = fmt.Sprintln("Swarm Initiated! -- <@sd-sre-nasa>, please help on-call engineers work the queue")
	_, _, err := helper.Slack.Client.PostMessage(command.ChannelID, slack.MsgOptionAttachments(attachment))
	if err != nil {
		return fmt.Errorf("failed to post message: %v", err)
	}

	// duplicate the slash command so we can call list and not modify the original object
	listCommand := command
	handleListCommand(listCommand, helper)

	// ping primary and ask what they are working on to prevent duplication of efforts
	attachment = slack.Attachment{}
	attachment.Text = fmt.Sprintf("<@%v> <@%v>, what incident are you working on?", "sre-platform-primary", "sre-platform-secondary")

	// send message, channel ID is available in command object
	_, _, err = helper.Slack.Client.PostMessage(command.ChannelID, slack.MsgOptionAttachments(attachment))
	if err != nil {
		return fmt.Errorf("failed to post message: %v", err)
	}
	return nil
}

// handleAssignCommand handles the /assign Slack command
func handleAssignCommand(command slack.SlashCommand, helper helpers.Helper) error {
	// adding some logic to check and make sure an INC was provided would be good...
	incToAssign := command.Text
	attachment := slack.Attachment{}

	// gets requestors email to query PD or Jira with for the respective user ID to assign to
	err := helper.Slack.GetSlackUserEmail(command.UserID)
	if err != nil {
		return fmt.Errorf("failed to get slack user's email: %v", err)
	}

	// hardcoding is bad, but for demo this is fine -- need a better way to determine incident type (PD vs Jira)
	if strings.Contains(incToAssign, "OHSS") {
		// email in my test slack doesn't match Jira, this is here for overwriting it in my demo
		// helper.Slack.UserEmail = os.Getenv("JIRA_EMAIL")

		// get requesting users Jira User ID to assign later
		err = helper.Jira.GetJiraUser(helper.Slack.UserEmail)
		if err != nil {
			return fmt.Errorf("failed to get PD User: %v", err)
		}

		err = helper.Jira.AssignJiraIssue(incToAssign)
		if err != nil {
			return fmt.Errorf("failed to assign incident %s to user %s: %v", incToAssign, helper.Jira.User.Name, err)
		}
	} else {
		// email in my test slack doesn't match PD, this is here for overwriting it in my demo
		// helper.Slack.UserEmail = os.Getenv("PD_EMAIL")

		// get requesting users PD User ID to assign later
		err = helper.PD.GetPDUser(helper.Slack.UserEmail)
		if err != nil {
			return fmt.Errorf("failed to get PD User: %v", err)
		}

		err = helper.PD.AssignPDInc(incToAssign)
		if err != nil {
			return fmt.Errorf("failed to assign incident %s to user %s: %v", incToAssign, helper.PD.User.ID, err)
		}
	}
	// send page to channel
	attachment.Text = fmt.Sprintf("Incident %v Assigned, thank you <@%v>!", incToAssign, command.UserName)
	_, _, err = helper.Slack.Client.PostMessage(command.ChannelID, slack.MsgOptionAttachments(attachment))
	if err != nil {
		return fmt.Errorf("failed to post message: %v", err)
	}
	return nil
}

// handleListCommand handles the /list-incs command and provides list of incs for Swarm process
func handleListCommand(command slack.SlashCommand, helper helpers.Helper) error {
	attachment := slack.Attachment{}

	err := helper.PD.GetPrimaryOnCall()
	if err != nil {
		return fmt.Errorf("failed to fetch primary on-call user: %v", err)
	}

	pdIncs, jiraIssues, err := helper.GetAllIncidents()
	if err != nil {
		return fmt.Errorf("failed to fetch all incidents: %v", err)
	}

	if len(pdIncs) == 0 && len(jiraIssues) == 0 {
		attachment.Text = "No incidents to address"
		_, _, err = helper.Slack.Client.PostMessage(command.ChannelID, slack.MsgOptionAttachments(attachment))
		if err != nil {
			return fmt.Errorf("failed to post message: %v", err)
		}
		return nil
	}

	// Slack messages take strings, and clickable links are handy
	// The format functions are providing new clean strings with new lines that properly print in Slack
	pdList := helpers.FormatIncList(pdIncs, helper.PD.BaseUrl)
	jiraList := helpers.FormatIncList(jiraIssues, helper.Jira.BaseUrl+"browse/")

	attachment.Pretext = "Here are a list of incidents needing review still:"
	attachment.Text = fmt.Sprintf(pdList + "\n" + jiraList)

	// send page to channel
	_, _, err = helper.Slack.Client.PostMessage(command.ChannelID, slack.MsgOptionAttachments(attachment))
	if err != nil {
		return fmt.Errorf("failed to post message: %v", err)
	}
	return nil
}
