package main

import (
	"math/rand"
	"time"

	"github.com/suapapa/site-ingress/internal/ingress"
)

const (
	movieLinesStr = `
- movie: "Matrix"
  line: "Neo, sooner or later you’re going to realize, just as I did, that there’s a difference between knowing the path and walking the path."
  character: "Morpheus"
- movie: "Forrest Gump"
  line: "Life is like a box of chocolates. You never know what you're gonna get."
  character: "Forrest Gump"
- movie: "Dune (2021)"
  line: "A great man doesn't seek to lead, he's called to it."
  character: "Leto"
- movie: "Interstella"
  line: "We’ll find a way, we always have."
  character: "Cooper"
- movie: "The Lord of the Rings: The Fellowship of the Ring"
  line: "All we have to decide is what to do with the time that is given us."
  character: "Gandalf"
- movie: "Star Wars: Episode V - The Empire Strikes Back"
  line: "Do or do not. There is no try."
  character: "Yoda"
`
)

var (
	r    *rand.Rand
	says []*ingress.Say
)

func init() {
	r = rand.New(rand.NewSource(time.Now().UnixNano()))
}

func GetRandomSay() *ingress.Say {
	i := r.Intn(len(says))
	return says[i]
}
