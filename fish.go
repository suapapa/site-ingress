package main

import (
	"math/rand"
	"time"

	"github.com/suapapa/site-ingress/internal/ingress"
)

var (
	r    *rand.Rand
	says []*ingress.Say
)

func init() {
	r = rand.New(rand.NewSource(time.Now().UnixNano()))
}

func GetRandomSay() *ingress.Say {
	if len(says) == 0 {
		return nil
	}
	return says[r.Intn(len(says))]
}
