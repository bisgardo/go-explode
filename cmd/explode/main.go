package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/bisgardo/go-explode"
)

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		log("usage: explode EXPRESSION...")
		return
	}

	for _, input := range args {
		firstErr := true
		res := explode.Explode(input, func(pos int, msg string) {
			if firstErr {
				log(input)
				firstErr = false
			}
			log(indent(pos) + "^ " + msg)
		})
		for _, r := range res {
			fmt.Println(r)
		}
	}
}

func indent(l int) string {
	if l <= 0 {
		return ""
	}
	return strings.Repeat(" ", l)
}

func log(s string) {
	_, _ = fmt.Fprintln(os.Stderr, s)
}
