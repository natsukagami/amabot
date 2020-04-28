package amabot

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

var (
	initMtx        sync.RWMutex
	serverID       string
	askChannelID   string
	queueChannelID string
)

// Checks whether the given message is a PM sent by the owner
func isOwnerPM(s *discordgo.Session, m *discordgo.Message) (bool, error) {
	if m.Author.ID != owner {
		return false, nil
	}

	channel, err := s.Channel(m.ChannelID)
	if err != nil {
		return false, errors.WithStack(err)
	}
	return channel.Type == discordgo.ChannelTypeDM, nil
}

// --- THE INIT HANDLERS

// Init handles the ama!init command.
// It locks the whole bot's activity to a single server.
// The command can only be run once by the owner, and will hang the bot until
// completion.
// When this command has not been run, ONLY messages from the owner's DM are allowed.
//
// Command format: `ama!init [server-id] [ask-channel-id] [queue-channel-id]`
//
// NO CHECKS are performed, as the bot WILL ASSUME its reading and reacting permissions on the ask channel and
// its reading, reacting and posting permissions on the queue channel.
func Init(s *discordgo.Session, m *discordgo.Message, content string) error {
	initMtx.Lock()
	defer initMtx.Unlock()
	if serverID != "" {
		return errors.New("The init command has already been ran")
	}

	// Try to parse the content into 3 IDs
	IDs := strings.Split(content, " ")
	if len(IDs) != 3 {
		return errors.New("Incorrect command")
	}

	// Try to parse each ID to get the name out
	server, err := s.Guild(IDs[0])
	if err != nil {
		return errors.WithStack(err)
	}
	askChannel, err := s.Channel(IDs[1])
	if err != nil || askChannel.GuildID != server.ID {
		return errors.New("Invalid ask channel")
	}
	queueChannel, err := s.Channel(IDs[2])
	if err != nil || queueChannel.GuildID != server.ID {
		return errors.New("Invalid queue channel")
	}

	// Should be good, we ask for confirmation and wait for reply
	ms, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Are you sure you want to activate the bot on server **%s**, with ask channel **%s** and queue channel **%s**? This is irreversible until the bot restarts.", server.Name, askChannel.Name, queueChannel.Name))
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
		serverID = server.ID
		askChannelID = askChannel.ID
		queueChannelID = queueChannel.ID
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Set up complete. The AMA may now begin."))
	}

	// Start the AMA
	ms, err = s.ChannelMessageSend(queueChannelID, fmt.Sprintf(`Welcome @everyone to the grand AMA! I am **AMABot**, glad to service you holding this event! 

The event will take place on two channels %s and %s. You can send questions (if you have any) on the first channel, but most of the time the event will take place on this channel.

Here are some commands that you need to use in order to post questions and find answers:

**In the %s channel**
`+"- `ama!ask [question]`: Use this command to post a question! Any question is okay, and you will see me giving you a 👍, that's when you know your question is in a queue!"+`

**In the %s channel, for AMA hosts**
`+"- `ama!next`: Mark the answer is done, and move to the next question.\n"+
		"- `ama!skip`: Mark the question as unanswered, and move to the next question.\n"+
		"- `ama!rotate`: Move the current question to the end of the queue.\n"+
		"- `ama!end`: End the AMA\n\n"+
		"That's it from my side! I hope you enjoy this AMA 😍😍😍", askChannel.Mention(), queueChannel.Mention(), askChannel.Mention(), queueChannel.Mention()))

	return nil
}

// Avatar sets the avatar of the bot.
//
// Command format: `ama!avatar [avatar-url]`
func Avatar(s *discordgo.Session, m *discordgo.Message, content string) error {
	resp, err := http.Get(content)
	if err != nil {
		return errors.Wrap(err, "Error retrieving the file")
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	img, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "Error reading the response")
	}

	contentType := http.DetectContentType(img)
	base64img := base64.StdEncoding.EncodeToString(img)

	avatar := fmt.Sprintf("data:%s;base64,%s", contentType, base64img)
	_, err = s.UserUpdate("", "", "", avatar, "")

	if err != nil {
		return errors.WithStack(err)
	}

	s.MessageReactionAdd(m.ChannelID, m.ID, "👍")
	return nil
}
