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
		res, err := explode.String(input)
		if err != nil {
			if explodeErr, ok := err.(explode.Error); ok {
				log(input)
				log(indent(explodeErr.Pos) + "^ " + explodeErr.Error())
			} else {
				log(err.Error())
			}
		}
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
