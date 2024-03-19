package autog

type PromptItem struct {
	Name string
	GetMessages func (query string) []ChatMessage
	GetPrompt func (query string) (role string, prompt string)
}

func (pi *PromptItem) doGetMessages(query string) []ChatMessage {
	if pi.GetMessages == nil {
		return []ChatMessage{}
	}
	return pi.GetMessages(query)
}

func (pi *PromptItem) doGetPrompt(query string) (role string, prompt string) {
	if pi.GetPrompt == nil {
		return "", ""
	}
	return pi.GetPrompt(query)
}