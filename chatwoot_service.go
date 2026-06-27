package main

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"mime"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/patrickmn/go-cache"
	"github.com/rs/zerolog/log"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
)

var (
	labelSyncMutexes sync.Map // map[string]*sync.Mutex
	globalLabelCache *cache.Cache
	cacheOnce        sync.Once
)

func getLabelSyncMutex(key string) *sync.Mutex {
	mut, _ := labelSyncMutexes.LoadOrStore(key, &sync.Mutex{})
	return mut.(*sync.Mutex)
}

func getGlobalLabelCache() *cache.Cache {
	cacheOnce.Do(func() {
		globalLabelCache = cache.New(20*time.Second, 1*time.Minute)
	})
	return globalLabelCache
}

// ChatwootService handles business logic for Chatwoot integration
type ChatwootService struct {
	client            *ChatwootClient
	waClient          *whatsmeow.Client
	db                *sqlx.DB
	cache             *cache.Cache
	userID            string
	inboxID           int
	inboxName         string
	signMsg           bool
	signDelimiter     string
	reopenConv        bool
	convPending       bool
	mergeBrazil       bool
	importGroups      bool
	sendStatusStories bool
	sendTyping        bool
	sendReadReceipts  bool
	enabledAt         time.Time
}

// ChatwootConfig represents Chatwoot configuration for a user
type ChatwootConfig struct {
	UserID              string    `db:"user_id"`
	Enabled             bool      `db:"enabled"`
	AccountID           string    `db:"account_id"`
	Token               string    `db:"token"`
	URL                 string    `db:"url"`
	InboxID             int       `db:"inbox_id"`
	InboxName           string    `db:"inbox_name"`
	WebhookURL          string    `db:"webhook_url"`
	WebhookSecret       string    `db:"webhook_secret"`
	SignMsg             bool      `db:"sign_msg"`
	SignDelimiter       string    `db:"sign_delimiter"`
	ReopenConversation  bool      `db:"reopen_conversation"`
	ConversationPending bool      `db:"conversation_pending"`
	MergeBrazilContacts bool      `db:"merge_brazil_contacts"`
	ImportGroups        bool      `db:"import_messages"`
	SendStatusStories   bool      `db:"send_status_stories"`
	SendTyping          bool      `db:"send_typing"`
	SendReadReceipts    bool      `db:"send_read_receipts"`
	EnabledAt           time.Time `db:"enabled_at"`
	CreatedAt           time.Time `db:"created_at"`
	UpdatedAt           time.Time `db:"updated_at"`
}

// NewChatwootService creates a new instance of ChatwootService
func NewChatwootService(db *sqlx.DB, config ChatwootConfig, waClient *whatsmeow.Client) *ChatwootService {
	client := NewChatwootClient(config.URL, config.AccountID, config.Token)

	return &ChatwootService{
		client:            client,
		waClient:          waClient,
		db:                db,
		cache:             cache.New(5*time.Minute, 10*time.Minute),
		userID:            config.UserID,
		inboxID:           config.InboxID,
		inboxName:         config.InboxName,
		signMsg:           config.SignMsg,
		signDelimiter:     config.SignDelimiter,
		reopenConv:        config.ReopenConversation,
		convPending:       config.ConversationPending,
		mergeBrazil:       config.MergeBrazilContacts,
		importGroups:      config.ImportGroups,
		sendStatusStories: config.SendStatusStories,
		sendTyping:        config.SendTyping,
		sendReadReceipts:  config.SendReadReceipts,
		enabledAt:         config.EnabledAt,
	}
}

// GetOrCreateContact finds or creates a contact in Chatwoot
func (s *ChatwootService) GetOrCreateContact(phoneNumber, name, avatarURL, jid string, isGroup bool) (*Contact, error) {
	// Calculate avatar hash (ignoring query parameters like tokens)
	avatarHash := avatarURL
	if avatarURL != "" {
		parts := strings.Split(avatarURL, "?")
		avatarHash = parts[0]
	}

	identifier := phoneNumber
	if isGroup {
		identifier = jid
	}
	cacheKey := fmt.Sprintf("contact:%s:%s", s.userID, identifier)

	var contact *Contact

	// Check cache first
	if cached, found := s.cache.Get(cacheKey); found {
		contact = cached.(*Contact)
	}

	if contact == nil {
		// Check database mapping
		var chatwootContactID int
		var dbQuery string
		var dbArgs []interface{}

		if isGroup {
			dbQuery = s.db.Rebind(`SELECT chatwoot_contact_id FROM chatwoot_contacts WHERE user_id = ? AND jid = ?`)
			dbArgs = []interface{}{s.userID, jid}
		} else {
			dbQuery = s.db.Rebind(`SELECT chatwoot_contact_id FROM chatwoot_contacts WHERE user_id = ? AND phone_number = ?`)
			dbArgs = []interface{}{s.userID, phoneNumber}
		}

		err := s.db.QueryRow(dbQuery, dbArgs...).Scan(&chatwootContactID)

		if err == nil {
			// Contact exists in our mapping, verify it exists in Chatwoot
			contact, err = s.getContactByID(chatwootContactID)
			if err != nil || contact == nil {
				contact = nil
				// If not found in Chatwoot, remove from our mapping
				if isGroup {
					_, _ = s.db.Exec(s.db.Rebind("DELETE FROM chatwoot_contacts WHERE user_id = ? AND jid = ?"), s.userID, jid)
				} else {
					_, _ = s.db.Exec(s.db.Rebind("DELETE FROM chatwoot_contacts WHERE user_id = ? AND phone_number = ?"), s.userID, phoneNumber)
				}
			}
		}
	}

	if contact == nil {
		// Search in Chatwoot
		var searchErr error
		if isGroup {
			contact, searchErr = s.client.FindContactByIdentifier(jid)
		} else {
			contact, searchErr = s.client.FindContact("+" + phoneNumber)
		}

		if searchErr != nil {
			log.Error().Err(searchErr).Str("phone", phoneNumber).Str("jid", jid).Msg("Error finding contact in Chatwoot")
		}
	}

	// Create if not found
	if contact == nil {
		displayName := name
		if displayName == "" {
			displayName = phoneNumber
		}
		if isGroup {
			displayName = displayName + " (GROUP)"
		}

		// For groups, we pass empty phone number to Chatwoot to avoid e164 validation errors
		cwPhone := "+" + phoneNumber
		if isGroup {
			cwPhone = ""
		}

		var err error
		contact, err = s.client.CreateContact(s.inboxID, displayName, cwPhone, jid, avatarURL)
		if err != nil {
			return nil, fmt.Errorf("failed to create contact: %w", err)
		}

		if contact.ID == 0 {
			return nil, fmt.Errorf("created contact has ID 0")
		}

		if avatarHash != "" {
			// Update the new contact instantly with the avatar hash to prevent future syncing
			customAttrs := make(map[string]interface{})
			if contact.CustomAttributes != nil {
				for k, v := range contact.CustomAttributes {
					customAttrs[k] = v
				}
			}
			customAttrs["wa_avatar_hash"] = avatarHash
			updated, errUpdate := s.client.UpdateContact(contact.ID, map[string]interface{}{"custom_attributes": customAttrs})
			if errUpdate == nil {
				contact = updated
			}

			// Also upload the avatar file directly to Chatwoot
			if avatarURL != "" {
				go func(contactID int, url string) {
					avatarData, contentType, err := downloadWebFile(url)
					if err == nil {
						_, _ = s.client.UpdateContactAvatar(contactID, "avatar.jpg", avatarData, contentType)
					}
				}(contact.ID, avatarURL)
			}
		}

		log.Info().
			Int("contactID", contact.ID).
			Str("name", displayName).
			Str("phone", phoneNumber).
			Msg("Created new Chatwoot contact")
	} else {
		// Update contact if needed (applies to cache hit, DB hit, and API search hit)
		needsUpdate := false
		updateData := make(map[string]interface{})
		customAttrs := make(map[string]interface{})

		if contact.CustomAttributes != nil {
			for k, v := range contact.CustomAttributes {
				customAttrs[k] = v
			}
		}

		// Check name update
		if contact.Name == "" || contact.Name == phoneNumber {
			if name != "" {
				updateData["name"] = name
				needsUpdate = true
			}
		} else if name != "" && contact.Name != name && contact.Name != name+" (GROUP)" {
			// Update name if changed
			displayName := name
			if isGroup {
				displayName = displayName + " (GROUP)"
			}
			updateData["name"] = displayName
			needsUpdate = true
		}

		// Check identifier update
		if jid != "" && contact.Identifier != jid {
			updateData["identifier"] = jid
			needsUpdate = true
		}

		// Check avatar update
		if avatarHash != "" {
			var currentHash string
			if h, ok := customAttrs["wa_avatar_hash"].(string); ok {
				currentHash = h
			}
			if currentHash != avatarHash {
				updateData["avatar_url"] = avatarURL
				customAttrs["wa_avatar_hash"] = avatarHash
				updateData["custom_attributes"] = customAttrs
				needsUpdate = true

				// Upload the avatar file directly to Chatwoot
				go func(contactID int, url string) {
					avatarData, contentType, err := downloadWebFile(url)
					if err == nil {
						_, _ = s.client.UpdateContactAvatar(contactID, "avatar.jpg", avatarData, contentType)
					}
				}(contact.ID, avatarURL)
			}
		}

		if needsUpdate {
			updatedContact, err := s.client.UpdateContact(contact.ID, updateData)
			if err != nil {
				log.Warn().Err(err).Int("contactID", contact.ID).Msg("Failed to update contact details")
			} else {
				log.Info().Int("contactID", contact.ID).Msg("Updated Chatwoot contact (Avatar/Name)")
				contact = updatedContact
			}
		}
	}

	// Store mapping in database
	dbPhone := phoneNumber
	if isGroup {
		dbPhone = jid // For groups, use JID as unique key in the phone_number column
	}

	_, err := s.db.Exec(s.db.Rebind(`
		INSERT INTO chatwoot_contacts (user_id, phone_number, chatwoot_contact_id, jid, is_group)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT (user_id, phone_number) 
		DO UPDATE SET chatwoot_contact_id = EXCLUDED.chatwoot_contact_id, jid = EXCLUDED.jid, is_group = EXCLUDED.is_group
	`), s.userID, dbPhone, contact.ID, jid, isGroup)

	if err != nil {
		log.Warn().Err(err).Msg("Failed to store contact mapping")
	}

	// Cache the result
	s.cache.Set(cacheKey, contact, cache.DefaultExpiration)

	return contact, nil
}

// getContactByID retrieves a contact by Chatwoot ID (helper method)
func (s *ChatwootService) getContactByID(contactID int) (*Contact, error) {
	contacts, err := s.client.SearchContact(fmt.Sprintf("%d", contactID))
	if err != nil {
		return nil, err
	}

	for _, contact := range contacts {
		if contact.ID == contactID {
			return &contact, nil
		}
	}

	return nil, nil
}

// NormalizeJID ensures we use the standard phone-based server for individual contacts
// It uses GetAltJID to resolve Linked Identity (LID) to the actual phone JID
func (s *ChatwootService) NormalizeJID(cli *whatsmeow.Client, jid types.JID) types.JID {
	if jid.Server == types.HiddenUserServer && cli != nil {
		altJID, err := cli.Store.GetAltJID(context.Background(), jid)
		if err == nil && !altJID.IsEmpty() {
			return altJID.ToNonAD()
		}
	}

	if jid.Server == types.HiddenUserServer {
		jid.Server = types.DefaultUserServer
	}
	return jid.ToNonAD()
}

// getLabelColor returns a distinct hex color based on the label name
func (s *ChatwootService) getLabelColor(name string) string {
	colors := []string{
		"#32CD32", // LimeGreen
		"#1E90FF", // DodgerBlue
		"#FF4500", // OrangeRed
		"#9370DB", // MediumPurple
		"#3CB371", // MediumSeaGreen
		"#FFD700", // Gold
		"#00CED1", // DarkCyan
		"#FF69B4", // HotPink
		"#CD5C5C", // IndianRed
		"#4682B4", // SteelBlue
	}

	// Simple hash to select color
	hash := 0
	for _, char := range name {
		hash += int(char)
	}
	return colors[hash%len(colors)]
}

// GetOrCreateConversation finds or creates a conversation in Chatwoot
func (s *ChatwootService) GetOrCreateConversation(remoteJID, phoneNumber, name, avatarURL string, isGroup bool) (int, error) {
	if isGroup && !s.importGroups {
		return 0, fmt.Errorf("group sync is disabled")
	}
	// Check cache first
	cacheKey := fmt.Sprintf("conversation:%s:%s", s.userID, remoteJID)
	if cached, found := s.cache.Get(cacheKey); found {
		return cached.(int), nil
	}

	// Check database mapping
	var conversationID int
	var status string
	query := s.db.Rebind(`
		SELECT chatwoot_conversation_id, status 
		FROM chatwoot_conversations 
		WHERE user_id = ? AND remote_jid = ?
	`)
	err := s.db.QueryRow(query, s.userID, remoteJID).Scan(&conversationID, &status)

	if err == nil {
		// Verify conversation still exists in Chatwoot
		conv, err := s.client.GetConversation(conversationID)
		if err == nil && conv != nil {
			// Reopen if needed
			if s.reopenConv && conv.Status == "resolved" {
				err = s.client.UpdateConversationStatus(conversationID, "open")
				if err == nil {
					s.db.Exec(s.db.Rebind("UPDATE chatwoot_conversations SET status = 'open', updated_at = CURRENT_TIMESTAMP WHERE user_id = ? AND remote_jid = ?"), s.userID, remoteJID)
					log.Info().Int("conversationID", conversationID).Msg("Reopened Chatwoot conversation")
				}
			}

			s.cache.Set(cacheKey, conversationID, cache.DefaultExpiration)
			return conversationID, nil
		}
		// If not found in Chatwoot, remove from our mapping
		s.db.Exec(s.db.Rebind("DELETE FROM chatwoot_conversations WHERE user_id = ? AND remote_jid = ?"), s.userID, remoteJID)
	}

	// Get or create contact
	contact, err := s.GetOrCreateContact(phoneNumber, name, avatarURL, remoteJID, isGroup)
	if err != nil {
		return 0, fmt.Errorf("failed to get/create contact: %w", err)
	}

	// Check for existing open/pending conversations in Chatwoot for this contact before creating a new one
	conversations, err := s.client.GetContactConversations(contact.ID)
	var existingConv *Conversation
	if err == nil {
		for _, c := range conversations {
			if c.InboxID == s.inboxID && (c.Status == "open" || c.Status == "pending" || c.Status == "snoozed") {
				existingConv = &c
				break
			}
		}
	}

	var conv *Conversation
	if existingConv != nil {
		conv = existingConv
		log.Info().Int("conversationID", conv.ID).Msg("Found existing active conversation in Chatwoot, reusing it")
	} else {
		// Create conversation
		conv, err = s.client.CreateConversation(contact.ID, s.inboxID)
		if err != nil {
			return 0, fmt.Errorf("failed to create conversation: %w", err)
		}

		if s.convPending {
			s.client.UpdateConversationStatus(conv.ID, "pending")
			conv.Status = "pending"
		}
	}

	conversationID = conv.ID
	convStatus := conv.Status

	// Store mapping
	upsertQuery := s.db.Rebind(`
		INSERT INTO chatwoot_conversations (user_id, remote_jid, chatwoot_conversation_id, chatwoot_inbox_id, status)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT (user_id, remote_jid) 
		DO UPDATE SET chatwoot_conversation_id = EXCLUDED.chatwoot_conversation_id, chatwoot_inbox_id = EXCLUDED.chatwoot_inbox_id, status = EXCLUDED.status, updated_at = CURRENT_TIMESTAMP
	`)
	if s.db.DriverName() == "sqlite" {
		upsertQuery = `
			INSERT INTO chatwoot_conversations (user_id, remote_jid, chatwoot_conversation_id, chatwoot_inbox_id, status)
			VALUES (?, ?, ?, ?, ?)
			ON CONFLICT (user_id, remote_jid) 
			DO UPDATE SET chatwoot_conversation_id = ?, chatwoot_inbox_id = ?, status = ?, updated_at = CURRENT_TIMESTAMP
		`
		_, err = s.db.Exec(upsertQuery, s.userID, remoteJID, conversationID, s.inboxID, convStatus, conversationID, s.inboxID, convStatus)
	} else {
		_, err = s.db.Exec(upsertQuery, s.userID, remoteJID, conversationID, s.inboxID, convStatus)
	}

	if err != nil {
		log.Warn().Err(err).Msg("Failed to store conversation mapping")
	}

	// Cache the result
	s.cache.Set(cacheKey, conversationID, cache.DefaultExpiration)

	log.Info().
		Int("conversationID", conversationID).
		Str("remoteJID", remoteJID).
		Msg("Created new Chatwoot conversation")

	// Sync existing labels from local DB to the new conversation
	go func() {
		labelQuery := s.db.Rebind(`
			SELECT cl.name 
			FROM chatwoot_labels cl
			JOIN chatwoot_conversation_labels ccl ON cl.user_id = ccl.user_id AND cl.label_id = ccl.label_id
			WHERE ccl.user_id = ? AND ccl.remote_jid = ?
		`)
		rows, err := s.db.Query(labelQuery, s.userID, remoteJID)
		if err == nil {
			defer rows.Close()
			var labels []string
			for rows.Next() {
				var name string
				if err := rows.Scan(&name); err == nil {
					labels = append(labels, name)
				}
			}
			if len(labels) > 0 {
				log.Info().Int("count", len(labels)).Str("jid", remoteJID).Msg("Syncing existing labels to new Chatwoot conversation")
				for _, l := range labels {
					parsedJID, _ := types.ParseJID(remoteJID)
					// Passing nil as client to avoid recursive GetOrCreateConversation
					_ = s.SyncLabelToChatwoot(nil, parsedJID, l, "add")
				}
			}
		}
	}()

	return conversationID, nil
}

// SendMessageToChatwoot sends a WhatsApp message to Chatwoot
func (s *ChatwootService) SendMessageToChatwoot(remoteJID, phoneNumber, name, content, avatarURL string, isGroup, fromMe bool, waMessageID string, fileName string, fileData []byte, contentType string, quotedMessageID string, senderJID string) error {
	conversationID, err := s.GetOrCreateConversation(remoteJID, phoneNumber, name, avatarURL, isGroup)
	if err != nil {
		return fmt.Errorf("failed to get/create conversation: %w", err)
	}

	// Check if message already exists to avoid duplication
	if fromMe {
		// Race condition fix: Wait for Webhook to insert the message mapping
		// We removed the memory cache check here because wmiau.go populates it for ALL events,
		// preventing us from syncing valid messages sent from the phone.
		time.Sleep(1 * time.Second)
	}

	var existingID int
	err = s.db.QueryRow(s.db.Rebind("SELECT chatwoot_message_id FROM chatwoot_messages WHERE user_id=? AND wa_message_id=?"), s.userID, waMessageID).Scan(&existingID)
	if err == nil {
		return nil // Message already synced
	}

	// Determine message type
	messageType := "incoming"
	if fromMe {
		messageType = "outgoing"
	}

	// Add signature if enabled and it's an outgoing message
	if s.signMsg && messageType == "outgoing" && s.signDelimiter != "" {
		content = content + s.signDelimiter
	}

	// Determine attributes for outgoing messages
	var attributes map[string]interface{}
	if fromMe {
		attributes = map[string]interface{}{
			"imported": true,
		}
	}

	// Handle reply to message
	if quotedMessageID != "" {
		var quotedChatwootMessageID int
		err := s.db.QueryRow(s.db.Rebind("SELECT chatwoot_message_id FROM chatwoot_messages WHERE user_id=? AND wa_message_id=?"), s.userID, quotedMessageID).Scan(&quotedChatwootMessageID)
		if err == nil {
			if attributes == nil {
				attributes = make(map[string]interface{})
			}
			attributes["in_reply_to"] = quotedChatwootMessageID
		} else {
			log.Warn().Str("quotedMessageID", quotedMessageID).Msg("Could not find Chatwoot message ID for quoted message")
		}
	}

	var msg *Message
	// Send message with attachment if present
	if len(fileData) > 0 {
		msg, err = s.client.CreateMessageWithAttachment(conversationID, content, messageType, fromMe, fileName, fileData, contentType, attributes)
	} else {
		msg, err = s.client.CreateMessage(conversationID, content, messageType, fromMe, contentType, attributes)
	}

	if err != nil {
		return fmt.Errorf("failed to create message: %w", err)
	}

	// Store message mapping
	_, err = s.db.Exec(s.db.Rebind(`
		INSERT INTO chatwoot_messages (user_id, wa_message_id, chatwoot_message_id, direction, synced, sender_jid, chat_jid, chatwoot_conversation_id)
		VALUES (?, ?, ?, ?, true, ?, ?, ?)
		ON CONFLICT (user_id, wa_message_id) DO NOTHING
	`), s.userID, waMessageID, msg.ID, messageType, senderJID, remoteJID, conversationID)

	if err != nil {
		log.Warn().Err(err).Msg("Failed to store message mapping")
	}

	log.Debug().
		Int("conversationID", conversationID).
		Int("messageID", msg.ID).
		Str("type", messageType).
		Bool("hasAttachment", len(fileData) > 0).
		Str("contentType", contentType).
		Msg("Sent message to Chatwoot")

	return nil
}

// NormalizePhoneNumber normalizes a phone number for Brazilian contacts
func (s *ChatwootService) NormalizePhoneNumber(phoneNumber string) []string {
	numbers := []string{phoneNumber}

	// Brazilian phone number handling
	if s.mergeBrazil && strings.HasPrefix(phoneNumber, "55") {
		if len(phoneNumber) == 13 {
			// Add 9 digit version: 5511999999999 -> 55119999999999
			with9 := phoneNumber[:4] + "9" + phoneNumber[4:]
			numbers = append(numbers, with9)
		} else if len(phoneNumber) == 12 {
			// Remove 9 digit version: 55119999999999 -> 5511999999999
			without9 := phoneNumber[:4] + phoneNumber[5:]
			numbers = append(numbers, without9)
		}
	}

	return numbers
}

// DeleteMessage deletes a message in Chatwoot
func (s *ChatwootService) DeleteMessage(waMessageID string) error {
	var chatwootMessageID int
	var conversationID int // this is chatwoot_conversation_id

	err := s.db.QueryRow(s.db.Rebind(`
		SELECT chatwoot_message_id, chatwoot_conversation_id 
		FROM chatwoot_messages 
		WHERE user_id=? AND wa_message_id=?
	`), s.userID, waMessageID).Scan(&chatwootMessageID, &conversationID)

	if err != nil {
		return fmt.Errorf("message not found in mapping: %w", err)
	}

	if chatwootMessageID == 0 || conversationID == 0 {
		return fmt.Errorf("invalid message or conversation id found in mapping")
	}

	err = s.client.DeleteMessage(conversationID, chatwootMessageID)
	if err != nil {
		return fmt.Errorf("failed to delete message in Chatwoot: %w", err)
	}

	return nil
}

// UpdateMessageStatus updates Chatwoot status for an outgoing WhatsApp message.
func (s *ChatwootService) UpdateMessageStatus(waMessageID, status string) error {
	if !s.sendReadReceipts {
		return nil
	}

	var chatwootMessageID int
	var conversationID int

	err := s.db.QueryRow(s.db.Rebind(`
		SELECT cm.chatwoot_message_id,
		       COALESCE(cm.chatwoot_conversation_id, cc.chatwoot_conversation_id, 0)
		FROM chatwoot_messages cm
		LEFT JOIN chatwoot_conversations cc
			ON cc.user_id = cm.user_id
			AND cc.remote_jid = cm.chat_jid
		WHERE cm.user_id = ?
		  AND cm.wa_message_id = ?
		  AND cm.direction = 'outgoing'
		LIMIT 1
	`), s.userID, waMessageID).Scan(&chatwootMessageID, &conversationID)
	if err != nil {
		return fmt.Errorf("message not found in Chatwoot mapping: %w", err)
	}

	if chatwootMessageID == 0 || conversationID == 0 {
		return fmt.Errorf("invalid Chatwoot message mapping for status update")
	}

	if err := s.client.UpdateMessageStatus(conversationID, chatwootMessageID, status); err != nil {
		return fmt.Errorf("failed to update Chatwoot message status: %w", err)
	}

	return nil
}

// SendReactionToChatwoot sends a reaction to Chatwoot as a private note
func (s *ChatwootService) SendReactionToChatwoot(remoteJID, reactorJID, waMessageID, reaction string) error {
	var chatwootMessageID int
	var conversationID int

	err := s.db.QueryRow(`
		SELECT cm.chatwoot_message_id, cc.chatwoot_conversation_id 
		FROM chatwoot_messages cm
		JOIN chatwoot_conversations cc ON cm.user_id = cc.user_id AND cm.chatwoot_conversation_id = cc.chatwoot_conversation_id
		WHERE cm.user_id=$1 AND cm.wa_message_id=$2
		 LIMIT 1
	`, s.userID, waMessageID).Scan(&chatwootMessageID, &conversationID)

	if err != nil {
		return fmt.Errorf("message not found for reaction: %w", err)
	}

	// Send as private note
	content := fmt.Sprintf("Reaction: %s", reaction)
	_, err = s.client.CreateMessage(conversationID, content, "incoming", true, "text", map[string]interface{}{
		"in_reply_to": chatwootMessageID,
	})

	if err != nil {
		return fmt.Errorf("failed to send reaction note: %w", err)
	}

	return nil
}

// HandleIncomingWhatsAppMessage processes an incoming WhatsApp message and forwards to Chatwoot
func (s *ChatwootService) HandleIncomingWhatsAppMessage(client *whatsmeow.Client, remoteJID types.JID, senderJID types.JID, message *waE2E.Message, messageID string, isGroup, fromMe bool, pushName string, senderAlt types.JID, recipientAlt types.JID, msgTimestamp time.Time) error {
	// Skip messages before integration was enabled
	if !msgTimestamp.IsZero() && !s.enabledAt.IsZero() && msgTimestamp.Before(s.enabledAt) {
		log.Debug().
			Str("messageID", messageID).
			Time("msgTime", msgTimestamp).
			Time("enabledAt", s.enabledAt).
			Msg("Skipping historical message")
		return nil
	}

	// Standardize JIDs
	remoteJID = s.NormalizeJID(client, remoteJID)
	senderJID = s.NormalizeJID(client, senderJID)

	phoneNumber := remoteJID.User
	senderPhone := senderJID.User
	contactIdentifier := senderJID.String()

	if isGroup {
		if !s.importGroups {
			return nil
		}
		phoneNumber = remoteJID.String()
	} else {
		// For DMs
		if fromMe {
			contactIdentifier = remoteJID.String()
			phoneNumber = remoteJID.User
		} else {
			contactIdentifier = senderJID.String()
			phoneNumber = senderJID.User
		}
	}

	// Get contact name and avatar
	name := pushName

	// TargetJID for Chatwoot message sender/recipient mapping
	targetJID := senderJID
	if !isGroup && fromMe {
		targetJID = remoteJID
	}

	if !isGroup && fromMe {
		name = ""
		targetJID = remoteJID

		if client != nil && client.Store != nil {
			contactInfo, err := client.Store.Contacts.GetContact(context.Background(), remoteJID)
			if err == nil && contactInfo.Found {
				if contactInfo.FullName != "" {
					name = contactInfo.FullName
				} else if contactInfo.PushName != "" {
					name = contactInfo.PushName
				} else if contactInfo.BusinessName != "" {
					name = contactInfo.BusinessName
				}
			}
		}
	}

	if name == "" {
		if !isGroup && fromMe {
			name = phoneNumber
		} else {
			name = senderPhone
		}
	}

	avatarURL := ""
	pic, err := client.GetProfilePictureInfo(context.Background(), targetJID, nil)
	if err == nil && pic != nil {
		avatarURL = pic.URL
	}

	if isGroup {
		participantPhone := senderPhone
		participantIdentifier := senderJID.String()

		if senderJID.Server == "lid" && !senderAlt.IsEmpty() {
			participantPhone = senderAlt.User
			participantIdentifier = senderAlt.String()
		}

		go func() {
			_, _ = s.GetOrCreateContact(participantPhone, pushName, avatarURL, participantIdentifier, false)
		}()

		groupName := remoteJID.String()
		groupPicURL := ""
		groupInfo, err := client.GetGroupInfo(context.Background(), remoteJID)
		if err == nil {
			groupName = groupInfo.Name
			groupPic, err := client.GetProfilePictureInfo(context.Background(), remoteJID, nil)
			if err == nil && groupPic != nil {
				groupPicURL = groupPic.URL
			}
		}

		phoneNumber = ""
		contactIdentifier = remoteJID.String()
		name = groupName
		avatarURL = groupPicURL
	}

	// Unwrap messages
	realMsg := message
	for realMsg != nil {
		if realMsg.GetDeviceSentMessage() != nil {
			realMsg = realMsg.GetDeviceSentMessage().GetMessage()
		} else if realMsg.GetViewOnceMessage() != nil {
			realMsg = realMsg.GetViewOnceMessage().GetMessage()
		} else if realMsg.GetViewOnceMessageV2() != nil {
			realMsg = realMsg.GetViewOnceMessageV2().GetMessage()
		} else if realMsg.GetViewOnceMessageV2Extension() != nil {
			realMsg = realMsg.GetViewOnceMessageV2Extension().GetMessage()
		} else if realMsg.GetEphemeralMessage() != nil {
			realMsg = realMsg.GetEphemeralMessage().GetMessage()
		} else if realMsg.GetDocumentWithCaptionMessage() != nil {
			realMsg = realMsg.GetDocumentWithCaptionMessage().GetMessage()
		} else {
			break
		}
	}
	if realMsg != nil {
		message = realMsg
	}

	content := ""
	var fileData []byte
	var fileName string
	var contentType string

	if reaction := message.GetReactionMessage(); reaction != nil {
		replyToMessageID := reaction.GetKey().GetID()
		textContent := reaction.GetText()

		var chatwootMessageID int
		var conversationID sql.NullInt64

		err := s.db.QueryRow(s.db.Rebind(`
			SELECT chatwoot_message_id, chatwoot_conversation_id 
			FROM chatwoot_messages 
			WHERE user_id=? AND wa_message_id=?
			LIMIT 1
		`), s.userID, replyToMessageID).Scan(&chatwootMessageID, &conversationID)

		if err != nil {
			log.Warn().Err(err).Str("replyToMessageID", replyToMessageID).Msg("Message not found for reaction sync")
			return nil
		}

		actualConvID := int(conversationID.Int64)
		if actualConvID == 0 {
			_ = s.db.QueryRow(s.db.Rebind(`SELECT chatwoot_conversation_id FROM chatwoot_conversations WHERE user_id=? AND remote_jid=?`), s.userID, remoteJID.String()).Scan(&actualConvID)
		}

		if actualConvID > 0 {
			var lang string
			err := s.db.QueryRow(s.db.Rebind("SELECT language FROM users WHERE id = ?"), s.userID).Scan(&lang)
			if err != nil || lang == "" {
				lang = "pt"
			}

			reactorName := name
			if isGroup && reaction.GetKey().GetParticipant() != "" {
				if pushName != "" {
					reactorName = pushName
				} else if senderPhone != "" && !strings.Contains(senderPhone, "@lid") {
					reactorName = senderPhone
				}
			}

			noteContent := fmt.Sprintf("%s: %s", reactorName, textContent)
			if fromMe {
				if lang == "en" {
					noteContent = fmt.Sprintf("You: %s", textContent)
				} else {
					noteContent = fmt.Sprintf("Você: %s", textContent)
				}
			}

			_, _ = s.client.CreateMessage(actualConvID, noteContent, "incoming", true, "text", map[string]interface{}{
				"in_reply_to": chatwootMessageID,
			})
		}

		return nil
	} else if protocolMsg := message.GetProtocolMessage(); protocolMsg != nil && protocolMsg.GetType() == 0 {
		msgIDToDelete := protocolMsg.GetKey().GetID()
		if msgIDToDelete != "" {
			_ = s.DeleteMessage(msgIDToDelete)
		}
		return nil
	} else if message.GetConversation() != "" {
		content = message.GetConversation()
	} else if message.GetExtendedTextMessage() != nil {
		extendedMsg := message.GetExtendedTextMessage()
		content = extendedMsg.GetText()

		if contextInfo := extendedMsg.GetContextInfo(); contextInfo != nil {
			if adReply := contextInfo.GetExternalAdReply(); adReply != nil {
				adTitle := adReply.GetTitle()
				adBody := adReply.GetBody()
				adURL := adReply.GetSourceURL()
				adThumbnail := adReply.GetThumbnail()

				adContext := "\n\n━━━━━━━━━━━━━━\n🎯 *Origem: Anúncio do WhatsApp*\n"
				if adTitle != "" {
					adContext += fmt.Sprintf("📦 *Produto:* %s\n", adTitle)
				}
				if adBody != "" {
					adContext += fmt.Sprintf("📝 *Descrição:* %s\n", adBody)
				}
				if adURL != "" {
					adContext += fmt.Sprintf("🔗 *Link:* %s\n", adURL)
				}
				content = content + adContext

				if len(adThumbnail) > 0 {
					fileData = adThumbnail
					fileName = "anuncio.jpg"
					contentType = "image/jpeg"
				}
			}
		}
	} else {
		var downloadErr error
		if img := message.GetImageMessage(); img != nil {
			caption := img.GetCaption()
			content = "imagem"
			if caption != "" {
				content = "imagem: " + caption
			}
			fileData, downloadErr = client.Download(context.Background(), img)
			contentType = "image/jpeg" // Force JPG
			exts, _ := mime.ExtensionsByType(contentType)
			if len(exts) > 0 {
				ext := exts[0]
				if ext == ".jpe" {
					ext = ".jpg"
				}
				fileName = "image" + ext
			} else {
				fileName = "image.jpg"
			}
		} else if video := message.GetVideoMessage(); video != nil {
			caption := video.GetCaption()
			content = "vídeo"
			if caption != "" {
				content = "vídeo: " + caption
			}
			fileData, downloadErr = client.Download(context.Background(), video)
			contentType = "video/mp4" // Force MP4
			fileName = "video.mp4"
		} else if audio := message.GetAudioMessage(); audio != nil {
			content = "áudio"
			fileData, downloadErr = client.Download(context.Background(), audio)
			contentType = "audio/mpeg" // Force MP3
			fileName = "audio.mp3"
		} else if doc := message.GetDocumentMessage(); doc != nil {
			fileName = doc.GetFileName()
			caption := doc.GetCaption()
			if caption != "" {
				content = fmt.Sprintf("arquivo: %s (%s)", fileName, caption)
			} else {
				content = "arquivo: " + fileName
			}
			fileData, downloadErr = client.Download(context.Background(), doc)
			contentType = doc.GetMimetype()
		} else if sticker := message.GetStickerMessage(); sticker != nil {
			content = "figurinha"
			fileData, downloadErr = client.Download(context.Background(), sticker)
			contentType = sticker.GetMimetype()
			fileName = "sticker.webp"
		} else if buttons := message.GetButtonsResponseMessage(); buttons != nil {
			content = buttons.GetSelectedButtonID()
		} else if list := message.GetListResponseMessage(); list != nil {
			content = list.GetSingleSelectReply().GetSelectedRowID()
		} else if interactive := message.GetInteractiveResponseMessage(); interactive != nil {
			if body := interactive.GetBody(); body != nil {
				content = body.GetText()
			}
		}

		if downloadErr != nil {
			log.Warn().Err(downloadErr).Msg("Failed to download media for Chatwoot")
		}
	}

	if content == "" && len(fileData) == 0 {
		return nil
	}

	if isGroup && !fromMe {
		senderTag := pushName
		if senderTag == "" {
			senderTag = senderJID.User
		}
		content = fmt.Sprintf("%s: %s", senderTag, content)
	}

	if fromMe {
		isAPI := false
		if _, found := apiMessageCache.Get(fmt.Sprintf("%s:%s", s.userID, messageID)); found {
			isAPI = true
		}

		if isAPI {
			content = fmt.Sprintf("%s\n\n(enviado via 💻)", content)
		} else {
			content = fmt.Sprintf("%s\n\n(enviado via 📱)", content)
		}
	}

	quotedMessageID := ""
	// Quoted extraction... (simplified for brevity)
	if message.GetExtendedTextMessage() != nil && message.GetExtendedTextMessage().GetContextInfo() != nil {
		quotedMessageID = message.GetExtendedTextMessage().GetContextInfo().GetStanzaID()
	}

	return s.SendMessageToChatwoot(contactIdentifier, phoneNumber, name, content, avatarURL, isGroup, fromMe, messageID, fileName, fileData, contentType, quotedMessageID, senderJID.String())
}

// GetConversationByRemoteJID retrieves the Chatwoot conversation ID
func (s *ChatwootService) GetConversationByRemoteJID(remoteJID string) (int, error) {
	var conversationID int
	err := s.db.QueryRow(s.db.Rebind(`
		SELECT chatwoot_conversation_id 
		FROM chatwoot_conversations 
		WHERE user_id = ? AND remote_jid = ?
	`), s.userID, remoteJID).Scan(&conversationID)
	return conversationID, err
}

// SendTypingEvent sends a typing indicator
func (s *ChatwootService) SendTypingEvent(remoteJID string, status string) error {
	conversationID, err := s.GetConversationByRemoteJID(remoteJID)
	if err != nil {
		return nil
	}

	cwStatus := "off"
	if status == "composing" {
		cwStatus = "on"
	}

	return s.client.ToggleTyping(conversationID, cwStatus)
}

// SyncLabelToChatwoot synchronizes WhatsApp labels
func (s *ChatwootService) SyncLabelToChatwoot(cli *whatsmeow.Client, remoteJID types.JID, labelName, action string) error {
	remoteJID = s.NormalizeJID(cli, remoteJID)
	remoteJIDStr := remoteJID.String()

	mutKey := fmt.Sprintf("%s:%s", s.userID, remoteJIDStr)
	mut := getLabelSyncMutex(mutKey)
	mut.Lock()
	defer mut.Unlock()

	performSync := func() error {
		conversationID, err := s.GetConversationByRemoteJID(remoteJIDStr)
		if err != nil {
			if action == "add" && cli != nil {
				conversationID, err = s.GetOrCreateConversation(remoteJIDStr, remoteJID.User, "", "", remoteJID.Server == types.GroupServer)
				if err != nil {
					return err
				}
			} else {
				return err
			}
		}

		cwLabelName := strings.ReplaceAll(strings.ToLower(labelName), " ", "_")
		currentLabels, err := s.client.GetConversationLabels(conversationID)
		if err != nil {
			return err
		}

		if action == "add" {
			found := false
			for _, l := range currentLabels {
				if l == cwLabelName {
					found = true
					break
				}
			}
			if !found {
				currentLabels = append(currentLabels, cwLabelName)
			}
		} else if action == "remove" {
			newLabels := []string{}
			for _, l := range currentLabels {
				if l != cwLabelName {
					newLabels = append(newLabels, l)
				}
			}
			currentLabels = newLabels
		}

		return s.client.AddConversationLabel(conversationID, currentLabels)
	}

	return performSync()
}

// downloadWebFile downloads a file from the web
func downloadWebFile(url string) ([]byte, string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("failed to download file: status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}

	contentType := resp.Header.Get("Content-Type")
	return data, contentType, nil
}
