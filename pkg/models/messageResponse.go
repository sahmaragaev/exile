package models

type TextContent struct {
	Value       string        `json:"value"`
	Annotations []interface{} `json:"annotations"`
}

type Content struct {
	Type string      `json:"type"`
	Text TextContent `json:"text"`
}

type ThreadMessage struct {
	ID          string      `json:"id"`
	Object      string      `json:"object"`
	CreatedAt   int64       `json:"created_at"`
	ThreadID    string      `json:"thread_id"`
	Role        string      `json:"role"`
	Content     []Content   `json:"content"`
	FileIDs     []string    `json:"file_ids"`
	AssistantID string      `json:"assistant_id"`
	RunID       string      `json:"run_id"`
	Metadata    interface{} `json:"metadata"`
}

type MessagesResponse struct {
	Object  string          `json:"object"`
	Data    []ThreadMessage `json:"data"`
	FirstID string          `json:"first_id"`
	LastID  string          `json:"last_id"`
	HasMore bool            `json:"has_more"`
}
