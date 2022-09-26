package main

import (
	"fmt"
	"math/rand"
)

var (
	movieLines = []*MovieLine{
		{
			Line:      "Neo, sooner or later you’re going to realize, just as I did, that there’s a difference between knowing the path and walking the path.",
			Character: "Morpheus",
		},
		{
			Line:      "Life is like a box of chocolates. You never know what you're gonna get.",
			Character: "Forrest Gump",
		},
	}
)

type MovieLine struct {
	Line, Character string
}

func (ml *MovieLine) String() string {
	return fmt.Sprintf("“%s” ― %s", ml.Line, ml.Character)
}

func randLine() string {
	i := rand.Intn(len(movieLines))
	return movieLines[i].String()
}
