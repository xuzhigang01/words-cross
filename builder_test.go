package main

import (
	"log"
	"testing"
)

func TestBuild(t *testing.T) {
	words := `ghost, hint, hoist, host, might, mint, 
		mist, moist, month, most, night, omit, shot,
		sigh, sight, sign, sting, thin, thing, tongs`
	_, err := Build(words)
	if err != nil {
		log.Println(err)
	}
}
