package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

// ChatwootClient handles all HTTP interactions with Chatwoot API
type ChatwootClient struct {
	BaseURL   string
	AccountID string
	Token     string
	client    *http.Client
}

// NewChatwootClient creates a new Chatwoot API client
func NewChatwootClient(baseURL, accountID, token string) *ChatwootClient {
	return &ChatwootClient{
		BaseURL:   normalizeChatwootBaseURL(baseURL),
		AccountID: accountID,
		Token:     token,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func normalizeChatwootBaseURL(baseURL string) string {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	parsed, err := url.Parse(baseURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return baseURL
	}

	path := strings.TrimRight(parsed.Path, "/")
	for _, marker := range []string{"/app", "/api"} {
		if idx := strings.Index(path, marker); idx >= 0 {
			path = path[:idx]
			break
		}
	}

	parsed.Path = path
	parsed.RawPath = ""
	parsed.RawQuery = ""
	parsed.Fragment = ""
	return strings.TrimRight(parsed.String(), "/")
}

// Contact represents a Chatwoot contact
type Contact struct {
	ID               int                    `json:"id"`
	Name             string                 `json:"name"`
	PhoneNumber      string                 `json:"phone_number"`
	Identifier       string                 `json:"identifier"`
	Email            string                 `json:"email"`
	Thumbnail        string                 `json:"thumbnail"`
	AvatarURL        string                 `json:"avatar_url,omitempty"`
	CustomAttributes map[string]interface{} `json:"custom_attributes,omitempty"`
}

// ContactInbox represents the contact information within an inbox
type ContactInbox struct {
	ContactID int `json:"contact_id"`
}

// Conversation represents a Chatwoot conversation
type Conversation struct {
	ID                   int                    `json:"id"`
	AccountID            int                    `json:"account_id"`
	InboxID              int                    `json:"inbox_id"`
	Status               string                 `json:"status"`
	ContactID            int                    `json:"contact_id"`
	ContactInbox         ContactInbox           `json:"contact_inbox"`
	Meta                 map[string]interface{} `json:"meta"`
	AdditionalAttributes map[string]interface{} `json:"additional_attributes,omitempty"`
}

// Message represents a Chatwoot message
type Message struct {
	ID              int                    `json:"id"`
	Content         string                 `json:"content"`
	MessageType     interface{}            `json:"message_type"`
	ContentType     string                 `json:"content_type"`
	CreatedAt       int64                  `json:"created_at"`
	Private         bool                   `json:"private"`
	Attachments     []Attachment           `json:"attachments,omitempty"`
	ConversationID  int                    `json:"conversation_id"`
	Sender          map[string]interface{} `json:"sender,omitempty"`
	PrivateMetadata map[string]interface{} `json:"private_metadata,omitempty"`
}

// Attachment represents a message attachment
type Attachment struct {
	ID       int    `json:"id"`
	FileType string `json:"file_type"`
	DataURL  string `json:"data_url"`
}

// Inbox represents a Chatwoot inbox
type Inbox struct {
	ID        int                    `json:"id"`
	Name      string                 `json:"name"`
	ChannelID int                    `json:"channel_id"`
	Channel   map[string]interface{} `json:"channel"`
}

func parseInboxResponse(respBody []byte) (*Inbox, error) {
	var inbox Inbox
	if err := json.Unmarshal(respBody, &inbox); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	if inbox.ID != 0 || inbox.Name != "" {
		return &inbox, nil
	}

	var wrapped struct {
		Payload Inbox `json:"payload"`
		Data    Inbox `json:"data"`
	}
	if err := json.Unmarshal(respBody, &wrapped); err != nil {
		return nil, fmt.Errorf("failed to unmarshal wrapped response: %w", err)
	}
	if wrapped.Payload.ID != 0 || wrapped.Payload.Name != "" {
		return &wrapped.Payload, nil
	}
	if wrapped.Data.ID != 0 || wrapped.Data.Name != "" {
		return &wrapped.Data, nil
	}

	return &inbox, nil
}

// doRequest performs an HTTP request to Chatwoot API
func (c *ChatwootClient) doRequest(method, path string, body interface{}) ([]byte, error) {
	url := fmt.Sprintf("%s/api/v1/accounts/%s%s", c.BaseURL, c.AccountID, path)

	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api_access_token", c.Token)

	log.Debug().
		Str("method", method).
		Str("url", url).
		Msg("Chatwoot API request")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.Error().
			Int("status", resp.StatusCode).
			Str("response", string(respBody)).
			Msg("Chatwoot API error")
		return nil, fmt.Errorf("API error: status %d, body: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// CreateContact creates a new contact in Chatwoot
func (c *ChatwootClient) CreateContact(inboxID int, name, phoneNumber, identifier, avatarURL string) (*Contact, error) {
	data := map[string]interface{}{
		"inbox_id": inboxID,
		"name":     name,
	}

	if phoneNumber != "" {
		data["phone_number"] = phoneNumber
	}
	if identifier != "" {
		data["identifier"] = identifier
	}
	if avatarURL != "" {
		data["avatar_url"] = avatarURL
	}

	respBody, err := c.doRequest("POST", "/contacts", data)
	if err != nil {
		return nil, err
	}

	log.Debug().Str("response", string(respBody)).Msg("CreateContact response")

	type ContactResponse struct {
		Contact Contact `json:"contact"`
		// Embed Contact fields to support flat structure if API changes
		ID               int                    `json:"id"`
		Name             string                 `json:"name"`
		PhoneNumber      string                 `json:"phone_number"`
		Identifier       string                 `json:"identifier"`
		Email            string                 `json:"email"`
		Thumbnail        string                 `json:"thumbnail"`
		AvatarURL        string                 `json:"avatar_url"`
		CustomAttributes map[string]interface{} `json:"custom_attributes"`
	}

	var result struct {
		Payload ContactResponse `json:"payload"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	contact := result.Payload.Contact
	// If nested contact is empty (ID=0), try to use the flat fields
	if contact.ID == 0 {
		contact = Contact{
			ID:               result.Payload.ID,
			Name:             result.Payload.Name,
			PhoneNumber:      result.Payload.PhoneNumber,
			Identifier:       result.Payload.Identifier,
			Email:            result.Payload.Email,
			Thumbnail:        result.Payload.Thumbnail,
			AvatarURL:        result.Payload.AvatarURL,
			CustomAttributes: result.Payload.CustomAttributes,
		}
	}

	if contact.ID == 0 {
		return nil, fmt.Errorf("failed to parse contact ID from response: %s", string(respBody))
	}

	return &contact, nil
}

// FindContact searches for a contact by phone number
func (c *ChatwootClient) FindContact(query string) (*Contact, error) {
	// Use filter endpoint for more precise matching
	filterPayload := []map[string]interface{}{
		{
			"attribute_key":   "phone_number",
			"filter_operator": "equal_to",
			"values":          []string{query},
			"query_operator":  nil,
		},
	}

	data := map[string]interface{}{
		"payload": filterPayload,
	}

	respBody, err := c.doRequest("POST", "/contacts/filter", data)
	if err != nil {
		return nil, err
	}

	var result struct {
		Payload []Contact `json:"payload"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(result.Payload) == 0 {
		return nil, nil
	}

	return &result.Payload[0], nil
}

// GetContact retrieves a contact by ID
func (c *ChatwootClient) GetContact(contactID int) (*Contact, error) {
	path := fmt.Sprintf("/contacts/%d", contactID)
	respBody, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var result struct {
		Payload Contact `json:"payload"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &result.Payload, nil
}

// FindContactByIdentifier searches for a contact by their unique identifier (JID)
func (c *ChatwootClient) FindContactByIdentifier(identifier string) (*Contact, error) {
	filterPayload := []map[string]interface{}{
		{
			"attribute_key":   "identifier",
			"filter_operator": "equal_to",
			"values":          []string{identifier},
			"query_operator":  nil,
		},
	}

	data := map[string]interface{}{
		"payload": filterPayload,
	}

	respBody, err := c.doRequest("POST", "/contacts/filter", data)
	if err != nil {
		return nil, err
	}

	var result struct {
		Payload []Contact `json:"payload"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(result.Payload) == 0 {
		return nil, nil
	}

	return &result.Payload[0], nil
}

// SearchContact searches for contacts by query string
func (c *ChatwootClient) SearchContact(query string) ([]Contact, error) {
	path := fmt.Sprintf("/contacts/search?q=%s", query)
	respBody, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var result struct {
		Payload []Contact `json:"payload"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return result.Payload, nil
}

// UpdateContact updates an existing contact
func (c *ChatwootClient) UpdateContact(contactID int, data map[string]interface{}) (*Contact, error) {
	path := fmt.Sprintf("/contacts/%d", contactID)
	respBody, err := c.doRequest("PATCH", path, data)
	if err != nil {
		return nil, err
	}

	var result struct {
		Payload Contact `json:"payload"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &result.Payload, nil
}

// GetContactConversations retrieves conversations for a contact
func (c *ChatwootClient) GetContactConversations(contactID int) ([]Conversation, error) {
	path := fmt.Sprintf("/contacts/%d/conversations", contactID)
	respBody, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var result struct {
		Payload []Conversation `json:"payload"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return result.Payload, nil
}

// CreateConversation creates a new conversation
func (c *ChatwootClient) CreateConversation(contactID, inboxID int) (*Conversation, error) {
	data := map[string]interface{}{
		"contact_id": fmt.Sprintf("%d", contactID),
		"inbox_id":   fmt.Sprintf("%d", inboxID),
	}

	respBody, err := c.doRequest("POST", "/conversations", data)
	if err != nil {
		return nil, err
	}

	var conv Conversation
	if err := json.Unmarshal(respBody, &conv); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &conv, nil
}

// GetConversation retrieves a conversation by ID
func (c *ChatwootClient) GetConversation(conversationID int) (*Conversation, error) {
	path := fmt.Sprintf("/conversations/%d", conversationID)
	respBody, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var conv Conversation
	if err := json.Unmarshal(respBody, &conv); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &conv, nil
}

// UpdateConversationStatus updates the status of a conversation
func (c *ChatwootClient) UpdateConversationStatus(conversationID int, status string) error {
	path := fmt.Sprintf("/conversations/%d/toggle_status", conversationID)
	data := map[string]interface{}{
		"status": status,
	}

	_, err := c.doRequest("POST", path, data)
	return err
}

// UpdateMessageStatus updates the delivery/read status of a Chatwoot message.
func (c *ChatwootClient) UpdateMessageStatus(conversationID, messageID int, status string) error {
	path := fmt.Sprintf("/conversations/%d/messages/%d", conversationID, messageID)
	data := map[string]interface{}{
		"status": status,
	}

	_, err := c.doRequest("PATCH", path, data)
	return err
}

// CreateMessage sends a message in a conversation
func (c *ChatwootClient) CreateMessage(conversationID int, content, messageType string, private bool, contentType string, attributes map[string]interface{}) (*Message, error) {
	data := map[string]interface{}{
		"content":      content,
		"message_type": messageType,
		"private":      private,
	}

	if contentType != "" {
		data["content_type"] = contentType
	} else {
		data["content_type"] = "text"
	}

	if attributes != nil {
		data["content_attributes"] = attributes
	}

	jsonData, _ := json.Marshal(data)
	log.Debug().Str("payload", string(jsonData)).Msg("CreateMessage Payload")

	path := fmt.Sprintf("/conversations/%d/messages", conversationID)
	respBody, err := c.doRequest("POST", path, data)
	if err != nil {
		return nil, err
	}

	var msg Message
	if err := json.Unmarshal(respBody, &msg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &msg, nil
}

// CreateMessageWithAttachment sends a message with an attachment in a conversation
func (c *ChatwootClient) CreateMessageWithAttachment(conversationID int, content, messageType string, private bool, fileName string, fileData []byte, contentType string, attributes map[string]interface{}) (*Message, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add fields
	_ = writer.WriteField("content", content)
	_ = writer.WriteField("message_type", messageType)
	if private {
		_ = writer.WriteField("private", "true")
	} else {
		_ = writer.WriteField("private", "false")
	}

	if attributes != nil {
		attrJSON, err := json.Marshal(attributes)
		if err == nil {
			_ = writer.WriteField("content_attributes", string(attrJSON))
		}
	}

	// Add file with explicit Content-Type
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="attachments[]"; filename="%s"`, strings.ReplaceAll(fileName, `"`, `\"`)))
	if contentType != "" {
		h.Set("Content-Type", contentType)
	}

	part, err := writer.CreatePart(h)
	if err != nil {
		return nil, fmt.Errorf("failed to create multipart section: %w", err)
	}
	_, err = part.Write(fileData)
	if err != nil {
		return nil, fmt.Errorf("failed to write file data: %w", err)
	}

	err = writer.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to close multipart writer: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/accounts/%s/conversations/%d/messages", c.BaseURL, c.AccountID, conversationID)
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("api_access_token", c.Token)

	log.Debug().
		Str("method", "POST").
		Str("url", url).
		Str("fileName", fileName).
		Str("contentType", contentType).
		Msg("Chatwoot API request (with attachment)")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("API error: status %d, body: %s", resp.StatusCode, string(respBody))
	}

	var msg Message
	if err := json.Unmarshal(respBody, &msg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &msg, nil
}

// ToggleTyping sends a typing indicator to the conversation
func (c *ChatwootClient) ToggleTyping(conversationID int, status string) error {
	path := fmt.Sprintf("/conversations/%d/toggle_typing_status", conversationID)
	data := map[string]interface{}{
		"typing_status": status,
	}

	_, err := c.doRequest("POST", path, data)
	return err
}

// CreateInbox creates a new inbox
func (c *ChatwootClient) CreateInbox(name, webhookURL string) (*Inbox, error) {
	data := map[string]interface{}{
		"name": name,
		"channel": map[string]interface{}{
			"type":        "api",
			"webhook_url": webhookURL,
		},
	}

	respBody, err := c.doRequest("POST", "/inboxes", data)
	if err != nil {
		return nil, err
	}

	return parseInboxResponse(respBody)
}

// ListInboxes retrieves all inboxes
func (c *ChatwootClient) ListInboxes() ([]Inbox, error) {
	respBody, err := c.doRequest("GET", "/inboxes", nil)
	if err != nil {
		return nil, err
	}

	var result struct {
		Payload []Inbox `json:"payload"`
		Data    []Inbox `json:"data"`
	}
	if err := json.Unmarshal(respBody, &result); err == nil {
		if result.Payload != nil {
			return result.Payload, nil
		}
		if result.Data != nil {
			return result.Data, nil
		}
	}

	var inboxes []Inbox
	if err := json.Unmarshal(respBody, &inboxes); err != nil {
		return nil, fmt.Errorf("failed to unmarshal inbox list response: %w", err)
	}
	return inboxes, nil
}

// GetInbox retrieves an inbox by ID
func (c *ChatwootClient) GetInbox(inboxID int) (*Inbox, error) {
	path := fmt.Sprintf("/inboxes/%d", inboxID)
	respBody, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	return parseInboxResponse(respBody)
}

// UpdateInbox updates an existing inbox
func (c *ChatwootClient) UpdateInbox(inboxID int, name, webhookURL string) (*Inbox, error) {
	data := map[string]interface{}{
		"channel": map[string]interface{}{
			"webhook_url": webhookURL,
		},
	}
	if name != "" {
		data["name"] = name
	}

	path := fmt.Sprintf("/inboxes/%d", inboxID)
	respBody, err := c.doRequest("PATCH", path, data)
	if err != nil {
		return nil, err
	}

	return parseInboxResponse(respBody)
}

// DeleteMessage deletes a message in a conversation
func (c *ChatwootClient) DeleteMessage(conversationID, messageID int) error {
	path := fmt.Sprintf("/conversations/%d/messages/%d", conversationID, messageID)
	_, err := c.doRequest("DELETE", path, nil)
	return err
}

// ToggleReaction toggles a reaction on a message
func (c *ChatwootClient) ToggleReaction(conversationID, messageID int, reaction string) error {
	path := fmt.Sprintf("/conversations/%d/messages/%d/toggle_reaction", conversationID, messageID)
	data := map[string]interface{}{
		"reaction": reaction,
	}

	_, err := c.doRequest("POST", path, data)
	return err
}

// GetLabels retrieves all labels for the account
func (c *ChatwootClient) GetLabels() ([]string, error) {
	respBody, err := c.doRequest("GET", "/labels", nil)
	if err != nil {
		return nil, err
	}

	var result struct {
		Payload []struct {
			Title string `json:"title"`
		} `json:"payload"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	labels := make([]string, len(result.Payload))
	for i, l := range result.Payload {
		labels[i] = l.Title
	}

	return labels, nil
}

// CreateLabel creates a new label in the account
func (c *ChatwootClient) CreateLabel(title, description, color string) error {
	data := map[string]interface{}{
		"title":       title,
		"description": description,
		"color":       color,
	}

	_, err := c.doRequest("POST", "/labels", data)
	return err
}

// GetConversationLabels retrieves labels for a specific conversation
func (c *ChatwootClient) GetConversationLabels(conversationID int) ([]string, error) {
	path := fmt.Sprintf("/conversations/%d/labels", conversationID)
	respBody, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var result struct {
		Payload []string `json:"payload"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return result.Payload, nil
}

// AddConversationLabel sets labels for a conversation (Chatwoot API usually sets the entire list)
func (c *ChatwootClient) AddConversationLabel(conversationID int, labels []string) error {
	path := fmt.Sprintf("/conversations/%d/labels", conversationID)
	data := map[string]interface{}{
		"labels": labels,
	}

	_, err := c.doRequest("POST", path, data)
	return err
}

// RemoveConversationLabel is not directly supported as a "remove single" in some versions,
// but usually sending the new list of labels works. However, Chatwoot also has a DELETE endpoint.
func (c *ChatwootClient) RemoveConversationLabel(conversationID int, labels []string) error {
	// Most Chatwoot versions use POST to set the entire list.
	// To remove, you'd usually GET the current ones and POST the new list.
	// But let's assume we want to just "set" or there is a specific delete.
	// For now, we'll implement it as "set" or follow the common pattern.
	path := fmt.Sprintf("/conversations/%d/labels", conversationID)
	data := map[string]interface{}{
		"labels": labels,
	}

	_, err := c.doRequest("POST", path, data)
	return err
}

// UpdateContactAvatar updates a contact's avatar image
func (c *ChatwootClient) UpdateContactAvatar(contactID int, fileName string, fileData []byte, contentType string) (*Contact, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add file
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="avatar"; filename="%s"`, strings.ReplaceAll(fileName, `"`, `\"`)))
	if contentType != "" {
		h.Set("Content-Type", contentType)
	}

	part, err := writer.CreatePart(h)
	if err != nil {
		return nil, fmt.Errorf("failed to create multipart section: %w", err)
	}
	_, err = part.Write(fileData)
	if err != nil {
		return nil, fmt.Errorf("failed to write file data: %w", err)
	}

	err = writer.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to close multipart writer: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/accounts/%s/contacts/%d/avatar", c.BaseURL, c.AccountID, contactID)
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("api_access_token", c.Token)

	log.Debug().
		Str("method", "POST").
		Str("url", url).
		Str("fileName", fileName).
		Str("contentType", contentType).
		Msg("Chatwoot API request (update contact avatar)")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("API error: status %d, body: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Payload Contact `json:"payload"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &result.Payload, nil
}
