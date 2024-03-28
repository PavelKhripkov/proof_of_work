package server

import "math/rand"

// quotes a list of quote the server can share with clients that provable make some work.
var quotes = []string{
	"When the going gets rough - turn to wonder.",
	"If you have knowledge, let others light their candles in it.",
	"A bird doesn't sing because it has an answer, it sings because it has a song.",
	"We are not what we know but what we are willing to learn.",
	"Good people are good because they've come to wisdom through failure.",
	"Your word is a lamp for my feet, a light for my path.",
	"The first problem for all of us, men and women, is not to learn, but to unlearn.",
	"Be wise like serpents and harmless like doves.",
	"By three methods we may learn wisdom: First, by reflection, which is noblest; Second, by imitation, which is easiest; and third by experience, which is the bitterest.",
	"The reason people find it so hard to be happy is that they always see the past better than it was, the present worse than it is, and the future less resolved than it will be.",
}

// randQuote returns a random quote.
func randQuote() string {
	return quotes[rand.Intn(len(quotes))]
}

func init() {
	// This allows us to not panic in runtime. We better panic on startup. See rand.Intn description.
	if len(quotes) == 0 {
		panic("no quotes specified")
	}
}
