package main

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/sashabaranov/go-openai"
)

type OpenChat struct {
	client   *openai.Client
	messages []openai.ChatCompletionMessage
	ctx      context.Context
}

func NewChat(key string, initMessage []openai.ChatCompletionMessage) *OpenChat {
	client := openai.NewClient(key)

	return &OpenChat{
		client:   client,
		messages: initMessage,
		ctx:      context.Background(),
	}
}

func (oc *OpenChat) Completion(content string) (string, error) {
	oc.messages = append(oc.messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: content,
	})
	resp, err := oc.client.CreateChatCompletion(
		oc.ctx,
		openai.ChatCompletionRequest{
			Model:       openai.GPT3Dot5Turbo,
			Messages:    oc.messages,
			Temperature: 0.1,
			N:           1,
		},
	)
	if err == nil {
		respText := resp.Choices[0].Message.Content
		oc.messages = append(oc.messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleAssistant,
			Content: respText,
		})
		return respText, nil
	}
	return "", err
}

func (oc *OpenChat) CompletionStream(content string, chars chan<- string, errs chan<- error) {
	oc.messages = append(oc.messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: content,
	})
	req := openai.ChatCompletionRequest{
		Model:    openai.GPT3Dot5Turbo,
		Messages: oc.messages,
		Stream:   true,
		N:        1,
	}
	stream, err := oc.client.CreateChatCompletionStream(oc.ctx, req)
	if err != nil {
		errs <- err
		return
	}
	defer stream.Close()

	var assistantContent string
	for {
		response, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			close(chars)
			oc.messages = append(oc.messages, openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleAssistant,
				Content: assistantContent,
			})
			return
		}
		if err != nil {
			errs <- err
			return
		}
		respContent := response.Choices[0].Delta.Content
		chars <- respContent
		assistantContent += respContent
	}
}

func (oc *OpenChat) GetConversation() string {
	var conversation string
	for _, message := range oc.messages {
		conversation += fmt.Sprintf("<%s>: %s\n", message.Role, message.Content)
	}
	return conversation
}
