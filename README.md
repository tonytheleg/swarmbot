# SwarmBot

SwarmBot is a Slack bot PoC to aid Red Hat SREP in intiating and facilitaing Swarms for Primary and Secondary. 

The Swarmbots main goals are:

* Monitor PagerDuty and Jira Incident queues and determine workload levels
* Upon hitting a pre-defined metric that indicates On-Call is under pressure, pages out to SREP in Slack to initate a Swarm
* Provides a simple interface for SREP Engineers to list incidents and assign one to themselves to work

Commands:

* /swarm-init: Manually starts the Swarm process
* /list-incs: Lists all incidents available to be worked
* /assign INC-ID: Assigns incident INC-ID to the requestor

Note: This POC was done using a personal Slack Workspace and trial PagerDuty account. Key environment variables would need to be updated for your envrionment if you wished to test this, its merely for demo/code sample purposes. It also uses probably not recommended auth methods but good for testing.

```bash
export PAGERDUTY_TOKEN="YOUR_PD_API_TOKEN"
export SLACK_AUTH_TOKEN="YOUR_SLACK_AUTH_TOKEN"
export SLACK_CHANNEL_ID="YOUR_SLACK_CHANNEL_ID (Not name but actual ID listed in API)"
export SLACK_APP_TOKEN="YOUR_SLACK_APP_TOKEN" 
export JIRA_PAT_TOKEN="YOUR_JIRA_PAT_TOKEN"
```