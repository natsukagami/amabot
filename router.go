package amabot

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

// Handle routes all the commands into place
func Handle(s *discordgo.Session, m *discordgo.MessageCreate) {
	content := m.Content
	// If message does not start with "ama!", skip
	if !strings.HasPrefix(content, "ama!") {
		return
	}
	content = content[len("ama!"):]
	// If system not init'd and message not from owner, skip
	initMtx.RLock()
	hasInit := serverID != ""
	initMtx.RUnlock()
	ownerPM, err := isOwnerPM(s, m.Message)
	if err != nil {
		goto handleError
	}
	if !hasInit && !ownerPM {
		// ignore
		return
	}

	fmt.Println(m, content)

	// If statement for each command
	if !hasInit && ownerPM && strings.HasPrefix(content, "init ") {
		err = Init(s, m.Message, content[len("init "):])
	}
	if hasInit {
		if m.ChannelID == askChannelID && strings.HasPrefix(content, "ask ") {
			err = Ask(s, m.Message, content[len("ask "):])
		}
		if m.ChannelID == queueChannelID {
			switch content {
			case "next":
				err = Next(s, m.Message)
			case "skip":
				err = Skip(s, m.Message)
			case "rotate":
				err = Rotate(s, m.Message)
			case "end":
				err = End(s, m.Message)
			}
		}
	}

	if err == nil {
		return
	}
handleError:
	log.Println(err)
	ms, _ := s.ChannelMessageSend(m.ChannelID, fmt.Sprint("An error has okuued, ", err))
	<-time.After(5 * time.Second)
	s.ChannelMessageDelete(ms.ChannelID, ms.ID)
	return
}
