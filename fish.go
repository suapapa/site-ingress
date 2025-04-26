package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/goccy/go-yaml"
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
	r          *rand.Rand
	movieLines []*MovieLine
)

func init() {
	r = rand.New(rand.NewSource(time.Now().UnixNano()))
	if err := yaml.Unmarshal([]byte(movieLinesStr), &movieLines); err != nil {
		panic(err)
	}
}

type MovieLine struct {
	Movie     string `yaml:"movie"`
	Line      string `yaml:"line"`
	Character string `yaml:"character"`
}

func (ml *MovieLine) String() string {
	return fmt.Sprintf("“%s” - %s; %s", ml.Line, ml.Character, ml.Movie)
}

func randLine() string {
	i := r.Intn(len(movieLines))
	return movieLines[i].String()
}
