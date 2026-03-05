package graph

import "github.com/cloudwego/eino/schema"

func buildMessagesWithHistory(history []*schema.Message, currentMessage string) []*schema.Message {
	messages := make([]*schema.Message, 0, len(history)+1)
	maxHistory := 10
	start := 0
	if len(history) > maxHistory {
		start = len(history) - maxHistory
	}
	for i := start; i < len(history); i++ {
		if history[i] != nil {
			messages = append(messages, history[i])
		}
	}
	messages = append(messages, schema.UserMessage(currentMessage))
	return messages
}
