package amabot

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

// Question holds the information of a queued Question.
type Question struct {
	ID      int
	Author  *discordgo.User
	Content string
}

// Announce formats the question as an announcement.
func (q *Question) Announce(mention bool) string {
	var user = q.Author.Mention()
	if !mention {
		user = q.Author.String()
	}
	return fmt.Sprintf(`**Question #%d** by **%s**:
%s`, q.ID, user, q.Content)
}

var (
	// Really, the queue is just a giant question channel
	queue = make(chan *Question, 50 /*50 is a lot of questions*/)
)
