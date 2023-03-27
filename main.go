package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
	"github.com/tonytheleg/swarmbot/pkg/handlers"
	"github.com/tonytheleg/swarmbot/pkg/helpers"
)

func main() {
	channelID := os.Getenv("SLACK_CHANNEL_ID")
	helper := helpers.NewHelper()

	// go-slack comes with a SocketMode package that accepts a Slack client and outputs a Socket mode client instead
	socketClient := socketmode.New(
		helper.Slack.Client,
		socketmode.OptionDebug(true),
		// Option to set custom logger
		socketmode.OptionLog(log.New(os.Stdout, "socketmode: ", log.Lshortfile|log.LstdFlags)),
	)

	// create a context that can be used to cancel goroutine
	ctx, cancel := context.WithCancel(context.Background())
	// Make this cancel called properly in a real program (graceful shutdown)
	defer cancel()

	// goroutine to handle pressure check on time interval
	go func(ctx context.Context, client *slack.Client) {
		for {
			select {
			case <-ctx.Done():
				log.Println("Shutting down socketmode listener")
				return
			default:
				log.Println("checking incident count for pressure")
				// Checks the queues to see if we've hit our pre-defined pressure count, triggers Swarm if yes
				pdPressureHit, err := helper.PD.CheckPDPressure()
				if err != nil {
					log.Printf("failed to check pressure: %v", err)
				}
				jiraPressureHit, err := helper.Jira.CheckJiraPressure()
				if err != nil {
					log.Printf("failed to check pressure: %v", err)
				}
				if pdPressureHit || jiraPressureHit {
					// easiest to just use the existing slash command function
					fakeCommand := slack.SlashCommand{
						ChannelID: channelID,
					}
					err := handlers.HandleSwarmInitCommand(fakeCommand, helper)
					if err != nil {
						log.Printf("failed to check pressure: %v", err)
					}
					// prevents swarm init from running for a while if its just run
					time.Sleep(60 * time.Minute)
				} else {
					time.Sleep(60 * time.Second)
				}
			}
		}
	}(ctx, helper.Slack.Client)

	go func(ctx context.Context, client *slack.Client, socketClient *socketmode.Client) {
		// Create a for loop that selects either the context cancellation or the events incomming
		for {
			select {
			// inscase context cancel is called exit the goroutine
			case <-ctx.Done():
				log.Println("Shutting down socketmode listener")
				return
			case event := <-socketClient.Events:
				// type switch the event
				switch event.Type {
				// handle slash commands
				case socketmode.EventTypeSlashCommand:
					// type case to correct event type
					command, ok := event.Data.(slack.SlashCommand)
					if !ok {
						log.Printf("Cloud not type cast message to a SlashCommand: %v/n", command)
					}
					// Curerntly only handling slash commands but other event types could be added
					// in this switch like bot mentions and such
					payload, err := handlers.HandleSlashCommand(command, helper)
					if err != nil {
						log.Fatal(err)
					}
					// dont forget to ack it
					socketClient.Ack(*event.Request, payload)
				}
			}
		}
	}(ctx, helper.Slack.Client, socketClient)
	socketClient.Run()
}
