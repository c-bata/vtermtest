package main

import (
	"fmt"
	"strings"

	"github.com/c-bata/go-prompt"
)

func completer(d prompt.Document) []prompt.Suggest {
	s := []prompt.Suggest{
		{Text: "users", Description: "user table"},
		{Text: "users-1", Description: "user-1 table"},
		{Text: "users-2", Description: "user-2 table"},
		{Text: "articles", Description: "articles table"},
	}
	return prompt.FilterHasPrefix(s, d.GetWordBeforeCursor(), true)
}

func executor(in string) {
	in = strings.TrimSpace(in)
	if in == "" {
		return
	}
	fmt.Println("You selected:", in)
}

func main() {
	p := prompt.New(
		executor,
		completer,
		prompt.OptionPrefix(">>> "),
		prompt.OptionTitle("SQL prompt"),
	)
	p.Run()
}
