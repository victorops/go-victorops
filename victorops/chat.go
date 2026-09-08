package victorops

import (
	"bytes"
	"context"
	"encoding/json"
	"strconv"
)

// ChatMessage represents a chat message to send to VictorOps
type ChatMessage struct {
	Username         string   `json:"username"`
	ExternalUsername string   `json:"externalUsername"`
	Text             string   `json:"text"`
	MonitoringTool   string   `json:"monitoringTool"`
	IncidentID       int      `json:"incidentId,omitempty"`
	Tags             []string `json:"tags,omitempty"`
}

// ChatMessageItem represents a chat message returned by the get-chat endpoint.
type ChatMessageItem struct {
	Username    string   `json:"username,omitempty"`
	Text        string   `json:"text,omitempty"`
	ServiceTime int64    `json:"serviceTime,omitempty"` // epoch milliseconds
	Sequence    int64    `json:"sequence,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

// ChatMessageList is the response from getting chat messages.
type ChatMessageList struct {
	Messages []ChatMessageItem `json:"messages,omitempty"`
	HasMore  bool              `json:"hasMore,omitempty"`
}

// GetChatMessagesOptions holds optional filters for GetChatMessages. A zero
// IncidentID means no incident filter; zero Limit/Offset use API defaults.
type GetChatMessagesOptions struct {
	IncidentID int64
	Limit      int
	Offset     int
}

// GetChatMessages retrieves chat messages for the organization, optionally
// filtered by incident ID and paged via limit/offset.
func (c *Client) GetChatMessages(ctx context.Context, options *GetChatMessagesOptions) (*ChatMessageList, *RequestDetails, error) {
	queryParams := make(map[string]string)
	if options != nil {
		if options.IncidentID > 0 {
			queryParams["incidentId"] = strconv.FormatInt(options.IncidentID, 10)
		}
		if options.Limit > 0 {
			queryParams["limit"] = strconv.Itoa(options.Limit)
		}
		if options.Offset > 0 {
			queryParams["offset"] = strconv.Itoa(options.Offset)
		}
	}

	details, err := c.makePublicAPICall(ctx, "GET", "v1/chat", bytes.NewBufferString("{}"), queryParams)
	if err != nil {
		return nil, details, err
	}

	var list ChatMessageList
	if err := json.Unmarshal([]byte(details.ResponseBody), &list); err != nil {
		return nil, details, err
	}

	return &list, details, nil
}

// SendChatMessage sends a chat message into VictorOps
func (c *Client) SendChatMessage(ctx context.Context, message *ChatMessage) (*RequestDetails, error) {
	jsonMessage, err := json.Marshal(message)
	if err != nil {
		return nil, err
	}

	details, err := c.makePublicAPICall(ctx, "POST", "v1/chat", bytes.NewBuffer(jsonMessage), nil)
	return details, err
}
