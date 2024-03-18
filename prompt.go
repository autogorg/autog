package autog

type PromptItem {
	Name string
	GetMessages func (query string) []ChatMessage
}

func (pi *PromptItem) doGetMessages(query string) []ChatMessage {
	if pi.GetMessages == nil {
		return []ChatMessage{}
	}
	return pi.GetMessages(query)
}