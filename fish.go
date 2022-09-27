package main

import (
	"fmt"
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().Unix())
}

var (
	movieLines = []*MovieLine{
		{
			Movie:     "Matrix",
			Line:      "Neo, sooner or later you’re going to realize, just as I did, that there’s a difference between knowing the path and walking the path.",
			Character: "Morpheus",
		},
		{
			Movie:     "Forrest Gump",
			Line:      "Life is like a box of chocolates. You never know what you're gonna get.",
			Character: "Forrest Gump",
		},
		{
			Movie:     "Dune (2021)",
			Line:      "A great man doesn't seek to lead, he's called to it",
			Character: "Leto",
		},
	}
)

type MovieLine struct {
	Movie, Line, Character string
}

func (ml *MovieLine) String() string {
	return fmt.Sprintf("“%s” - %s; %s", ml.Line, ml.Character, ml.Movie)
}

func randLine() string {
	i := rand.Intn(len(movieLines))
	return movieLines[i].String()
}
