package main

import (
	"fmt"
	"runtime"

	"github.com/sashabaranov/go-openai"
)

var runtimeOS = func() string {
	os := runtime.GOOS
	if os == "darwin" {
		os = "macOS"
	}
	return os
}()

var initMessage = []openai.ChatCompletionMessage{
	{
		Role:    openai.ChatMessageRoleSystem,
		Content: fmt.Sprintf("You are a computer maintenance assistant, skilled in using bash to help users solve computer problems. The current user is using the %s system.", runtimeOS),
	},
	{
		Role:    openai.ChatMessageRoleUser,
		Content: "How many files are there in the current directory?",
	},
	{
		Role:    openai.ChatMessageRoleAssistant,
		Content: "```bash\nls | wc -l\n```",
	},
}
