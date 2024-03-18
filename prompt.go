package autog

type PromptItem {
	Name string
	GenMessages func (query string) []ChatMessage
}

func (pi *PromptItem) doGenMessages(query string) []ChatMessage {
	return pi.GenMessages(query)
}