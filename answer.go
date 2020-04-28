package amabot

import (
	"fmt"
	"os"
	"time"

	"github.com/pkg/errors"

	"github.com/bwmarrin/discordgo"
)

// The AMA host should be the only ones with write access to the Answer channel (outside AMABot)

// There are 4 available commands from the bot
// - next: mark the current question as answered, announce the next one.
// - skip: mark the current question as not answered, announce the next one.
// - rotate: delay the current question to the end of the queue, and announce the next one.
// - end: stop the AMA. Bot will shutdown
type command int

const (
	next command = iota
	skip
	rotate
)

// Channels for loop control
var (
	commandChannel = make(chan command, 1)
	endChannel     = make(chan struct{})
	pendingChannel = make(chan *Question)
)

// Loop is the main AMA loop.
func Loop(s *discordgo.Session, stopListening func()) {
	// Question counter
	var (
		counter       = 0
		totalAnswered = 0
		totalSkipped  = 0
		totalPending  = 0
	)

mainLoop:
	for {
		select {
		// "Queue": assign an ID and continue
		case q := <-queue:
			counter++
			totalPending++
			q.ID = counter
			go func() { pendingChannel <- q }()
		// "End": stop listening for questions
		case <-endChannel:
			// Exit
			break mainLoop
		// Huh, what command?
		case <-commandChannel:
			s.ChannelMessageSend(queueChannelID, "Nothing to do, skipping command")
			continue mainLoop
		// There is a question
		case question := <-pendingChannel:
			totalPending--
			s.ChannelMessageSend(queueChannelID, question.Announce(true)+fmt.Sprintf("\n(%d pending)", totalPending))
			// Wait for a command, or end
			select {
			case <-endChannel:
				break mainLoop
			case c := <-commandChannel:
				if c == next {
					totalAnswered++
				} else if c == skip {
					totalSkipped++
				} else {
					totalPending++
					go func() { pendingChannel <- question }()
				}
			}
		}
	}

	// Cleanup
	// First stop taking questions
	stopListening()
	// Drain all queue
drain1:
	for {
		select {
		case q := <-queue:
			counter++
			totalPending++
			q.ID = counter
			go func(q *Question) { pendingChannel <- q }(q)
		default:
			break drain1
		}
	}
	// First announce that the AMA has ended.
	s.ChannelMessageSend(queueChannelID, fmt.Sprintf(`@everyone The AMA has officially ended! Thank you for participating.
Since the beginning, we've had a total of **%d questions**, of which **%d answers** have been provided.
Below are some unanswered questions:`, counter, totalAnswered))
	// Drain the unanswered questions
drain2:
	for {
		select {
		case q := <-pendingChannel:
			s.ChannelMessageSend(queueChannelID, q.Announce(false))
		default:
			break drain2
		}
	}
	// End the game
	os.Exit(0)
}

func sendCommand(s *discordgo.Session, m *discordgo.Message, c command) error {
	select {
	case commandChannel <- c:
		s.ChannelMessageDelete(m.ChannelID, m.ID)
		return nil
	default:
		return errors.New("Too many pending commands")
	}
}

// Next marks the current question as answered and announces the next one.
func Next(s *discordgo.Session, m *discordgo.Message) error {
	return sendCommand(s, m, next)
}

// Skip marks the current question as unanswered and announces the next one.
func Skip(s *discordgo.Session, m *discordgo.Message) error {
	return sendCommand(s, m, skip)
}

// Rotate moves the question to the end of the queue.
func Rotate(s *discordgo.Session, m *discordgo.Message) error {
	return sendCommand(s, m, rotate)
}

// End ends the AMA.
func End(s *discordgo.Session, m *discordgo.Message) error {
	// Should be good, we ask for confirmation and wait for reply
	ms, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Are you sure you want to end the AMA?"))
	if err != nil {
		return errors.WithStack(err)
	}
	// Add reactions
	s.MessageReactionAdd(ms.ChannelID, ms.ID, "✅")
	s.MessageReactionAdd(ms.ChannelID, ms.ID, "❌")

	confirm := make(chan bool, 1)
	// Watch for reactions
	go func() {
		for {
			accepted, err := s.MessageReactions(ms.ChannelID, ms.ID, "✅", 2)
			if err == nil && len(accepted) == 2 {
				confirm <- true
				break
			}
			rejected, err := s.MessageReactions(ms.ChannelID, ms.ID, "❌", 2)
			if err == nil && len(rejected) == 2 {
				confirm <- false
				break
			}
			<-time.After(time.Second / 2)
		}
	}()

	if <-confirm {
		endChannel <- struct{}{}
	}

	return nil
}
