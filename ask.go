package amabot

import (
	"strings"

	"github.com/bwmarrin/discordgo"
)

// Ask posts a question. ONLY WORKS IN THE ASK CHANNEL.
//
// Command format: `ama!ask [question]`
func Ask(s *discordgo.Session, m *discordgo.Message, content string) error {
	// Just push the thing into the queue
	queue <- &Question{
		Author:  m.Author,
		Content: strings.TrimSpace(content),
	}

	// React to inform
	s.MessageReactionAdd(m.ChannelID, m.ID, "👍")

	return nil
}
