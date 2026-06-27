package main

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"hash/fnv"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/gorilla/mux"
	gocache "github.com/patrickmn/go-cache"
	"github.com/rs/zerolog/log"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/appstate"
	"go.mau.fi/whatsmeow/proto/waCommon"
	"go.mau.fi/whatsmeow/proto/waE2E"
	watypes "go.mau.fi/whatsmeow/types"
	"google.golang.org/protobuf/proto"
)

// SetupChatwoot configures Chatwoot integration for a user
func (s *server) SetupChatwoot() http.HandlerFunc {
	type setupRequest struct {
		Enabled             bool   `json:"enabled"`
		AccountID           string `json:"accountId"`
		Token               string `json:"token"`
		URL                 string `json:"url"`
		InboxID             int    `json:"inboxId,omitempty"`
		InboxName           string `json:"nameInbox,omitempty"`
		WebhookURL          string `json:"webhookUrl,omitempty"`
		WebhookSecret       string `json:"webhookSecret,omitempty"`
		SignMsg             bool   `json:"signMsg"`
		SignDelimiter       string `json:"signDelimiter,omitempty"`
		ReopenConversation  bool   `json:"reopenConversation"`
		ConversationPending bool   `json:"conversationPending"`
		MergeBrazilContacts bool   `json:"mergeBrazilContacts"`
		ImportGroups        bool   `json:"importGroups"`
		SendStatusStories   bool   `json:"sendStatusStories"`
		SendTyping          bool   `json:"sendTyping"`
		SendReadReceipts    bool   `json:"sendReadReceipts"`
		AutoCreate          bool   `json:"autoCreate"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		// Safely retrieve user info
		val := r.Context().Value("userinfo")
		if val == nil {
			log.Error().Msg("userinfo missing from context")
			s.Respond(w, r, http.StatusUnauthorized, errors.New("unauthorized"))
			return
		}

		values, ok := val.(Values)
		if !ok {
			log.Error().Msgf("userinfo is not Values type, got %T", val)
			s.Respond(w, r, http.StatusInternalServerError, errors.New("internal server error"))
			return
		}

		txtid := values.Get("Id")
		token := values.Get("Token")

		var err error
		var req setupRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			s.Respond(w, r, http.StatusBadRequest, errors.New("invalid request body"))
			return
		}

		// Trim whitespace from inputs
		req.URL = strings.TrimSpace(req.URL)
		req.URL = strings.TrimRight(req.URL, "/")
		req.URL = normalizeChatwootBaseURL(req.URL)
		req.AccountID = strings.TrimSpace(req.AccountID)
		req.Token = strings.TrimSpace(req.Token)
		req.WebhookURL = strings.TrimSpace(req.WebhookURL)
		req.WebhookSecret = strings.TrimSpace(req.WebhookSecret)

		// Handle masked token (********)
		if req.Token == "********" {
			var storedToken string
			query := s.db.Rebind("SELECT token FROM chatwoot_config WHERE user_id = ?")
			err := s.db.QueryRow(query, txtid).Scan(&storedToken)
			if err == nil && storedToken != "" {
				req.Token = storedToken
			} else {
				// If we can't find the stored token, we can't validate.
				// However, the user might be trying to save for the first time with a dummy value (unlikely but possible)
				// or the DB fetch failed.
				// In this specific case, if it fails, we let it proceed to validation which will likely fail
				// if the token is truly "********".
				log.Warn().Err(err).Str("userID", txtid).Msg("Failed to retrieve stored token for masked update or no token found")
			}
		}

		if req.WebhookSecret == "" || req.WebhookSecret == "********" {
			var storedSecret string
			query := s.db.Rebind("SELECT COALESCE(webhook_secret, '') FROM chatwoot_config WHERE user_id = ?")
			if err := s.db.QueryRow(query, txtid).Scan(&storedSecret); err == nil {
				req.WebhookSecret = storedSecret
			}
		}

		// Set default inbox name
		inboxName := req.InboxName
		if inboxName == "" {
			name := values.Get("Name")
			inboxName = name
		}

		// Set default sign delimiter
		signDelimiter := req.SignDelimiter
		if req.SignMsg && signDelimiter == "" {
			signDelimiter = "\n\n---\n"
		}

		var inboxID int
		setupWarning := ""
		defaultWebhookURL := fmt.Sprintf("%s/chatwoot/webhook/%s", getServerURL(), txtid)
		effectiveWebhookURL := req.WebhookURL
		if effectiveWebhookURL == "" {
			effectiveWebhookURL = defaultWebhookURL
		}

		// Validate required fields if enabled
		if req.Enabled {
			if req.URL == "" || req.AccountID == "" || req.Token == "" {
				s.Respond(w, r, http.StatusBadRequest, errors.New("url, accountId, and token are required when enabled"))
				return
			}

			if req.InboxID > 0 {
				inboxID = req.InboxID
				inboxName = strings.TrimSpace(inboxName)
				setupWarning = fmt.Sprintf("Configuracao salva usando a caixa existente ID %d. A Miauzap vai reutilizar este webhook da caixa: %s", inboxID, effectiveWebhookURL)
				log.Info().Int("inboxID", inboxID).Str("name", inboxName).Msg("Using existing Chatwoot inbox by informed ID without remote validation")
			} else {
				// Validate Connection
				client := NewChatwootClient(req.URL, req.AccountID, req.Token)
				inboxes, err := client.ListInboxes()
				if err != nil {
					log.Error().Err(err).Str("url", req.URL).Str("accountID", req.AccountID).Msg("Failed to validate Chatwoot credentials")
					s.Respond(w, r, http.StatusBadRequest, fmt.Errorf("nao foi possivel conectar ao Chatwoot. Confirme se a URL e a raiz do Chatwoot, sem /app nem /api: %v", err))
					return
				}

				// Find inbox if exists. Compare case-insensitive and trimmed to avoid creating duplicates.
				var existingInbox *Inbox
				for i := range inboxes {
					if strings.EqualFold(strings.TrimSpace(inboxes[i].Name), strings.TrimSpace(inboxName)) {
						existingInbox = &inboxes[i]
						break
					}
				}

				if existingInbox != nil {
					inboxID = existingInbox.ID
					inboxName = strings.TrimSpace(existingInbox.Name)
					log.Info().Int("inboxID", inboxID).Str("name", inboxName).Msg("Using existing inbox")
					if _, err := client.UpdateInbox(inboxID, "", effectiveWebhookURL); err != nil {
						log.Warn().Err(err).Int("inboxID", inboxID).Msg("Failed to update existing inbox webhook; continuing with saved config")
						setupWarning = fmt.Sprintf("Configuracao salva, mas nao foi possivel atualizar o webhook automaticamente. Configure manualmente este webhook na caixa do Chatwoot: %s", effectiveWebhookURL)
					}
				} else if req.AutoCreate {
					log.Info().Str("webhookURL", effectiveWebhookURL).Str("name", inboxName).Msg("Creating new Chatwoot inbox")

					inbox, err := client.CreateInbox(inboxName, effectiveWebhookURL)
					if err != nil {
						log.Error().Err(err).Str("url", req.URL).Str("accountID", req.AccountID).Msg("Failed to create Chatwoot inbox")
						s.Respond(w, r, http.StatusBadRequest, fmt.Errorf("nao foi possivel criar a caixa de entrada '%s' no Chatwoot: %v", inboxName, err))
						return
					}
					inboxID = inbox.ID
					log.Info().Int("inboxID", inboxID).Str("name", inboxName).Msg("Created new inbox")
				} else {
					s.Respond(w, r, http.StatusBadRequest, fmt.Errorf("caixa de entrada '%s' nao encontrada no Chatwoot; informe o ID da caixa existente ou ative a criacao automatica", inboxName))
					return
				}
			}

			// Update the request object with the resolved InboxName
			req.InboxName = inboxName
		}

		// Get current configuration to check if enablement is changing
		var currentEnabled bool
		err = s.db.QueryRow(s.db.Rebind("SELECT enabled FROM chatwoot_config WHERE user_id = ?"), txtid).Scan(&currentEnabled)
		if err != nil && err != sql.ErrNoRows {
			log.Error().Err(err).Str("userID", txtid).Msg("Failed to check current Chatwoot status in DB")
		}

		enabledAtUpdate := "chatwoot_config.enabled_at"
		if req.Enabled && !currentEnabled {
			// Integration being enabled now
			enabledAtUpdate = "CURRENT_TIMESTAMP"
		}

		// Store configuration in database
		// Note: Using CURRENT_TIMESTAMP for the INSERT part is safe as a starting point
		query := fmt.Sprintf(`
			INSERT INTO chatwoot_config (
				user_id, enabled, account_id, token, url, inbox_id, inbox_name, webhook_url, webhook_secret,
				sign_msg, sign_delimiter, reopen_conversation, conversation_pending, merge_brazil_contacts, import_messages, send_status_stories, send_typing, send_read_receipts,
				enabled_at, updated_at
			) VALUES (:user_id, :enabled, :account_id, :token, :url, :inbox_id, :inbox_name, :webhook_url, :webhook_secret,
				:sign_msg, :sign_delimiter, :reopen_conversation, :conversation_pending, :merge_brazil_contacts, :import_messages, :send_status_stories, :send_typing, :send_read_receipts,
				CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
			ON CONFLICT (user_id) DO UPDATE SET
				enabled = :enabled, account_id = :account_id, token = :token, url = :url, 
				inbox_id = :inbox_id, inbox_name = :inbox_name,
				webhook_url = :webhook_url,
				webhook_secret = :webhook_secret,
				sign_msg = :sign_msg, sign_delimiter = :sign_delimiter, 
				reopen_conversation = :reopen_conversation, 
				conversation_pending = :conversation_pending, 
				merge_brazil_contacts = :merge_brazil_contacts, 
				import_messages = :import_messages,
				send_status_stories = :send_status_stories,
				send_typing = :send_typing,
				send_read_receipts = :send_read_receipts,
				enabled_at = %s,
				updated_at = CURRENT_TIMESTAMP
		`, enabledAtUpdate)

		params := map[string]interface{}{
			"user_id":               txtid,
			"enabled":               req.Enabled,
			"account_id":            req.AccountID,
			"token":                 req.Token,
			"url":                   req.URL,
			"inbox_id":              inboxID,
			"inbox_name":            inboxName,
			"webhook_url":           effectiveWebhookURL,
			"webhook_secret":        req.WebhookSecret,
			"sign_msg":              req.SignMsg,
			"sign_delimiter":        signDelimiter,
			"reopen_conversation":   req.ReopenConversation,
			"conversation_pending":  req.ConversationPending,
			"merge_brazil_contacts": req.MergeBrazilContacts,
			"import_messages":       req.ImportGroups,
			"send_status_stories":   req.SendStatusStories,
			"send_typing":           req.SendTyping,
			"send_read_receipts":    req.SendReadReceipts,
		}

		_, err = s.db.NamedExec(query, params)

		if err != nil {
			log.Error().Err(err).Msg("Failed to save Chatwoot configuration")
			s.Respond(w, r, http.StatusInternalServerError, errors.New("failed to save configuration"))
			return
		}

		// Clear cache
		userinfocache.Delete(token)

		response := map[string]interface{}{
			"enabled":             req.Enabled,
			"accountId":           req.AccountID,
			"token":               req.Token,
			"url":                 req.URL,
			"inboxId":             inboxID,
			"inboxName":           inboxName,
			"webhookUrl":          effectiveWebhookURL,
			"webhookSecret":       "",
			"signMsg":             req.SignMsg,
			"signDelimiter":       signDelimiter,
			"reopenConversation":  req.ReopenConversation,
			"conversationPending": req.ConversationPending,
			"mergeBrazilContacts": req.MergeBrazilContacts,
			"importGroups":        req.ImportGroups,
			"sendStatusStories":   req.SendStatusStories,
			"sendTyping":          req.SendTyping,
			"sendReadReceipts":    req.SendReadReceipts,
		}
		if req.WebhookSecret != "" {
			response["webhookSecret"] = "********"
		}
		if setupWarning != "" {
			response["setupWarning"] = setupWarning
		}

		s.Respond(w, r, http.StatusOK, response)
	}
}

// GetChatwootConfig retrieves the Chatwoot configuration for a user
func (s *server) GetChatwootConfig() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Error().Interface("panic", rec).Msg("Panic in GetChatwootConfig")
				fmt.Fprintf(os.Stderr, "Panic in GetChatwootConfig: %v\n", rec)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()

		log.Info().Msg("GetChatwootConfig called")
		fmt.Println("GetChatwootConfig called via fmt")

		val := r.Context().Value("userinfo")
		if val == nil {
			log.Error().Msg("userinfo missing from context")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		values, ok := val.(Values)
		if !ok {
			log.Error().Msgf("userinfo is not Values type, got %T", val)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		txtid := values.Get("Id")

		log.Info().Str("userID", txtid).Msg("Querying Chatwoot config")

		var config ChatwootConfig
		err := s.db.QueryRow(s.db.Rebind(`
			SELECT enabled, account_id, token, url, inbox_id, inbox_name, COALESCE(webhook_url, ''), COALESCE(webhook_secret, ''),
				   sign_msg, sign_delimiter, reopen_conversation, conversation_pending, merge_brazil_contacts, import_messages, send_status_stories, send_typing, send_read_receipts, enabled_at
			FROM chatwoot_config
			WHERE user_id = ?
		`), txtid).Scan(
			&config.Enabled, &config.AccountID, &config.Token, &config.URL,
			&config.InboxID, &config.InboxName, &config.WebhookURL, &config.WebhookSecret, &config.SignMsg, &config.SignDelimiter,
			&config.ReopenConversation, &config.ConversationPending, &config.MergeBrazilContacts, &config.ImportGroups, &config.SendStatusStories, &config.SendTyping, &config.SendReadReceipts, &config.EnabledAt,
		)

		if err == sql.ErrNoRows {
			response := map[string]interface{}{
				"enabled": false,
				"message": "Chatwoot not configured",
			}
			s.Respond(w, r, http.StatusOK, response)
			return
		}

		if err != nil {
			log.Error().Err(err).Msg("Failed to get Chatwoot configuration")
			s.Respond(w, r, http.StatusInternalServerError, errors.New("failed to get configuration"))
			return
		}

		response := map[string]interface{}{
			"enabled":             config.Enabled,
			"accountId":           config.AccountID,
			"token":               config.Token,
			"url":                 config.URL,
			"inboxId":             config.InboxID,
			"inboxName":           config.InboxName,
			"webhookUrl":          config.WebhookURL,
			"webhookSecret":       "",
			"signMsg":             config.SignMsg,
			"signDelimiter":       config.SignDelimiter,
			"reopenConversation":  config.ReopenConversation,
			"conversationPending": config.ConversationPending,
			"mergeBrazilContacts": config.MergeBrazilContacts,
			"importGroups":        config.ImportGroups,
			"sendStatusStories":   config.SendStatusStories,
			"sendTyping":          config.SendTyping,
			"sendReadReceipts":    config.SendReadReceipts,
		}
		if response["webhookUrl"] == "" {
			response["webhookUrl"] = fmt.Sprintf("%s/chatwoot/webhook/%s", getServerURL(), txtid)
		}
		if config.WebhookSecret != "" {
			response["webhookSecret"] = "********"
		}

		s.Respond(w, r, http.StatusOK, response)
	}
}

type labelQueueItem struct {
	client      *whatsmeow.Client
	jid         watypes.JID
	labelID     string
	labelName   string
	ensureLabel bool
	add         bool
}

var (
	labelQueueMap = make(map[string]chan labelQueueItem)
	labelQueueMut sync.Mutex
)

// getOrCreateLabelQueue ensures a dedicated label sync queue exists for a user and starts its worker.
func getOrCreateLabelQueue(userID string) chan labelQueueItem {
	labelQueueMut.Lock()
	defer labelQueueMut.Unlock()

	q, exists := labelQueueMap[userID]
	if !exists {
		// Buffered channel to prevent blocking the webhook handler entirely
		q = make(chan labelQueueItem, 1000)
		labelQueueMap[userID] = q

		go func(ch chan labelQueueItem) {
			for item := range ch {
				if item.ensureLabel && item.labelName != "" {
					err := item.client.SendAppState(context.Background(), appstate.BuildLabelEdit(item.labelID, item.labelName, 0, false))
					if err != nil {
						log.Error().Err(err).Str("labelID", item.labelID).Str("labelName", item.labelName).Msg("Failed to create WhatsApp label via queue")
						time.Sleep(1 * time.Second)
						continue
					}
					time.Sleep(1 * time.Second)
				}
				err := item.client.SendAppState(context.Background(), appstate.BuildLabelChat(item.jid, item.labelID, item.add))
				if err != nil {
					log.Error().Err(err).Str("labelID", item.labelID).Str("jid", item.jid.String()).Msg("Failed to sync label to WhatsApp via queue")
				} else {
					log.Info().Str("labelID", item.labelID).Str("jid", item.jid.String()).Msg("Synced label to WhatsApp via queue")
				}
				// 1 second delay between operations to prevent freezing WhatsApp Web state sync
				time.Sleep(1 * time.Second)
			}
		}(q)
	}
	return q
}

func normalizeChatwootLabelForWhatsApp(labelName string) string {
	labelName = strings.TrimSpace(strings.ReplaceAll(labelName, "_", " "))
	if labelName == "" {
		return ""
	}
	return labelName
}

func deterministicWhatsAppLabelID(userID, labelName string) string {
	hash := fnv.New32a()
	_, _ = hash.Write([]byte(strings.ToLower(userID + ":" + labelName)))
	return fmt.Sprintf("9%d", hash.Sum32())
}

func (s *server) getOrCreateWhatsAppLabelID(userID, labelName string) (string, bool, error) {
	if labelName == "" {
		return "", false, errors.New("empty label name")
	}

	var labelID string
	err := s.db.QueryRow(s.db.Rebind(`
		SELECT label_id
		FROM chatwoot_labels
		WHERE user_id = ? AND lower(name) = lower(?)
		LIMIT 1
	`), userID, labelName).Scan(&labelID)
	if err == nil {
		return labelID, false, nil
	}
	if err != sql.ErrNoRows {
		return "", false, err
	}

	labelID = deterministicWhatsAppLabelID(userID, labelName)
	_, err = s.db.Exec(s.db.Rebind(`
		INSERT INTO chatwoot_labels (user_id, label_id, name)
		VALUES (?, ?, ?)
		ON CONFLICT (user_id, label_id) DO UPDATE SET name = EXCLUDED.name
	`), userID, labelID, labelName)
	if err != nil {
		return "", false, err
	}
	return labelID, true, nil
}

func (s *server) setLocalConversationLabel(userID, remoteJID, labelID string, add bool) {
	var err error
	if add {
		_, err = s.db.Exec(s.db.Rebind(`
			INSERT INTO chatwoot_conversation_labels (user_id, remote_jid, label_id)
			VALUES (?, ?, ?)
			ON CONFLICT DO NOTHING
		`), userID, remoteJID, labelID)
	} else {
		_, err = s.db.Exec(s.db.Rebind(`
			DELETE FROM chatwoot_conversation_labels
			WHERE user_id = ? AND remote_jid = ? AND label_id = ?
		`), userID, remoteJID, labelID)
	}
	if err != nil {
		log.Warn().Err(err).Str("labelID", labelID).Str("jid", remoteJID).Bool("add", add).Msg("Failed to update local label association")
	}
}

func (s *server) storeChatwootOutgoingMapping(userID, remoteJID string, conversationID int, waMsgID string, chatwootMessageID int) {
	if waMsgID == "" {
		return
	}

	dedupKey := fmt.Sprintf("%s:%s:%s", userID, remoteJID, waMsgID)
	messageDedupCache.Set(dedupKey, true, gocache.DefaultExpiration)

	_, err := s.db.Exec(s.db.Rebind(`
		INSERT INTO chatwoot_messages (user_id, wa_message_id, chatwoot_message_id, direction, synced, chat_jid, chatwoot_conversation_id)
		VALUES (?, ?, ?, 'outgoing', true, ?, ?)
		ON CONFLICT (user_id, wa_message_id) DO NOTHING
	`), userID, waMsgID, chatwootMessageID, remoteJID, conversationID)
	if err != nil {
		log.Error().Err(err).Str("waMsgID", waMsgID).Int("cwMsgID", chatwootMessageID).Msg("Failed to store outgoing Chatwoot message mapping")
	} else {
		log.Debug().Str("waMsgID", waMsgID).Int("cwMsgID", chatwootMessageID).Str("jid", remoteJID).Msg("Stored outgoing Chatwoot message mapping")
	}
}

func verifyChatwootWebhookSignature(secret string, body []byte, signatureHeader, timestampHeader string) bool {
	secret = strings.TrimSpace(secret)
	signatureHeader = strings.TrimSpace(signatureHeader)
	timestampHeader = strings.TrimSpace(timestampHeader)
	if secret == "" || signatureHeader == "" || timestampHeader == "" {
		return false
	}

	payload := timestampHeader + "." + string(body)
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(payload))
	expected := "sha256=" + hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(expected), []byte(signatureHeader))
}

func (s *server) resolveChatwootWebhookUserID(pathUserID string, inboxID int) string {
	if inboxID == 0 {
		return pathUserID
	}

	rows, err := s.db.Query(s.db.Rebind(`
		SELECT user_id
		FROM chatwoot_config
		WHERE enabled = TRUE
		  AND inbox_id = ?
		ORDER BY
		  CASE
		    WHEN COALESCE(webhook_url, '') LIKE ? THEN 0
		    WHEN user_id = ? THEN 1
		    ELSE 2
		  END,
		  updated_at DESC
	`), inboxID, "%/chatwoot/webhook/"+pathUserID+"%", pathUserID)
	if err != nil {
		return pathUserID
	}
	defer rows.Close()

	fallbackUserID := ""
	for rows.Next() {
		var candidateUserID string
		if err := rows.Scan(&candidateUserID); err != nil || candidateUserID == "" {
			continue
		}
		if fallbackUserID == "" {
			fallbackUserID = candidateUserID
		}

		if client := clientManager.GetWhatsmeowClient(candidateUserID); client != nil && client.IsLoggedIn() {
			if candidateUserID != pathUserID {
				log.Info().
					Str("pathUserID", pathUserID).
					Str("resolvedUserID", candidateUserID).
					Int("inboxID", inboxID).
					Msg("Resolved reused Chatwoot webhook to connected instance")
			}
			return candidateUserID
		}
	}

	resolvedUserID := fallbackUserID
	if resolvedUserID == "" {
		resolvedUserID = pathUserID
	}

	if resolvedUserID != pathUserID {
		log.Info().
			Str("pathUserID", pathUserID).
			Str("resolvedUserID", resolvedUserID).
			Int("inboxID", inboxID).
			Msg("Resolved reused Chatwoot webhook to configured instance")
	}

	return resolvedUserID
}

// ChatwootWebhook handles incoming webhooks from Chatwoot
func (s *server) ChatwootWebhook() http.HandlerFunc {
	type webhookPayload struct {
		Event        string `json:"event"`
		MessageType  string `json:"message_type"`
		ID           int    `json:"id"`
		Content      string `json:"content"`
		Conversation struct {
			ID        int    `json:"id"`
			ContactID int    `json:"contact_id"`
			InboxID   int    `json:"inbox_id"`
			Status    string `json:"status"`
			Meta      struct {
				Sender struct {
					ID          int    `json:"id"`
					PhoneNumber string `json:"phone_number"`
					Identifier  string `json:"identifier"`
				} `json:"sender"`
			} `json:"meta"`
			ContactInbox struct {
				ContactID int `json:"contact_id"`
			} `json:"contact_inbox"`
		} `json:"conversation"`
		Sender struct {
			Type string `json:"type"`
		} `json:"sender"`
		User *struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		} `json:"user"`
		Private     bool `json:"private"`
		Attachments []struct {
			DataURL  string `json:"data_url"`
			FileType string `json:"file_type"`
		} `json:"attachments"`
		PrivateMetadata   map[string]interface{}   `json:"private_metadata"`
		ContentAttributes map[string]interface{}   `json:"content_attributes"`
		SourceID          string                   `json:"source_id"`
		ChangedAttributes []map[string]interface{} `json:"changed_attributes"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		userID := vars["instance"]

		// Read body for debugging
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			log.Error().Err(err).Msg("Failed to read request body")
			s.Respond(w, r, http.StatusBadRequest, errors.New("read body failed"))
			return
		}
		// Restore body
		r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		log.Debug().Str("payload", string(bodyBytes)).Msg("Chatwoot Webhook Raw Payload")

		var payload webhookPayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			log.Error().Err(err).Msg("Failed to decode webhook payload")
			s.Respond(w, r, http.StatusBadRequest, errors.New("invalid payload"))
			return
		}

		userID = s.resolveChatwootWebhookUserID(userID, payload.Conversation.InboxID)

		var webhookSecret string
		if err := s.db.QueryRow(s.db.Rebind(`
			SELECT COALESCE(webhook_secret, '')
			FROM chatwoot_config
			WHERE user_id = ?
		`), userID).Scan(&webhookSecret); err == nil && strings.TrimSpace(webhookSecret) != "" {
			if !verifyChatwootWebhookSignature(webhookSecret, bodyBytes, r.Header.Get("X-Chatwoot-Signature"), r.Header.Get("X-Chatwoot-Timestamp")) {
				log.Warn().Str("userID", userID).Msg("Rejected Chatwoot webhook with invalid signature")
				s.Respond(w, r, http.StatusUnauthorized, errors.New("invalid chatwoot webhook signature"))
				return
			}
		} else if err != nil && err != sql.ErrNoRows {
			log.Warn().Err(err).Str("userID", userID).Msg("Failed to load Chatwoot webhook secret")
		}

		log.Info().
			Str("event", payload.Event).
			Str("messageType", payload.MessageType).
			Int("conversationID", payload.Conversation.ID).
			Str("userID", userID).
			Interface("privateMetadata", payload.PrivateMetadata).
			Interface("contentAttributes", payload.ContentAttributes).
			Str("sourceID", payload.SourceID).
			Msg("Received Chatwoot webhook")

		isDeleteEvent := payload.Event == "message_deleted" || (payload.Event == "message_updated" && payload.Content == "" && !payload.Private)

		// Check for SourceID to prevent loop (messages originating from WA have SourceID starting with WAID)
		if payload.SourceID != "" && strings.HasPrefix(payload.SourceID, "WAID") && !isDeleteEvent && payload.Event != "conversation_updated" {
			log.Info().Str("sourceID", payload.SourceID).Msg("Ignoring message from WhatsApp source to prevent loop")
			s.Respond(w, r, http.StatusOK, map[string]string{"status": "ignored", "reason": "loop_prevention"})
			return
		}

		// Handle typing events
		if payload.Event == "conversation_typing_on" || payload.Event == "conversation_typing_off" {
			// Only sync typing events *from agents* back to WhatsApp
			if payload.User == nil {
				log.Debug().Msg("Ignoring typing event that does not originate from an agent")
				s.Respond(w, r, http.StatusOK, map[string]string{"status": "ignored_contact_typing"})
				return
			}

			// Get WhatsApp client
			client := clientManager.GetWhatsmeowClient(userID)
			if client == nil || !client.IsLoggedIn() {
				s.Respond(w, r, http.StatusOK, map[string]string{"status": "ignored_no_client"})
				return
			}

			// Validate if SendTyping is enabled in ChatwootConfig
			var sendTypingEnabled bool
			err := s.db.QueryRow(s.db.Rebind(`SELECT COALESCE(send_typing, true) FROM chatwoot_config WHERE user_id = ?`), userID).Scan(&sendTypingEnabled)
			if err == nil && !sendTypingEnabled {
				s.Respond(w, r, http.StatusOK, map[string]string{"status": "ignored_typing_disabled"})
				return
			}

			// Get remote JID
			var remoteJID string
			err = s.db.QueryRow(s.db.Rebind(`
				SELECT remote_jid 
				FROM chatwoot_conversations 
				WHERE user_id = ? AND chatwoot_conversation_id = ?
			`), userID, payload.Conversation.ID).Scan(&remoteJID)

			if err != nil {
				// Conversation not found possibly
				s.Respond(w, r, http.StatusOK, map[string]string{"status": "ignored_no_conv"})
				return
			}

			jid, err := watypes.ParseJID(remoteJID)
			if err != nil {
				return
			}
			// Ensure it's a bare JID (no device part) for typing status
			jid = jid.ToNonAD()

			state := watypes.ChatPresenceComposing
			if payload.Event == "conversation_typing_off" {
				state = watypes.ChatPresencePaused
			}

			client.SendChatPresence(r.Context(), jid, state, watypes.ChatPresenceMediaText)
			s.Respond(w, r, http.StatusOK, map[string]string{"status": "success"})
			return
		}

		// Handle message deletion
		if isDeleteEvent {
			log.Info().Int("cwMsgID", payload.ID).Str("event", payload.Event).Msg("Processing message deletion/update sync to WhatsApp")

			// Get WhatsApp client
			client := clientManager.GetWhatsmeowClient(userID)
			if client == nil || !client.IsLoggedIn() {
				log.Warn().Str("userID", userID).Msg("WhatsApp client not available for deletion sync")
				s.Respond(w, r, http.StatusOK, map[string]string{"status": "ignored_no_client"})
				return
			}

			// Find the WhatsApp Message ID
			var waMsgID string
			var remoteJID string
			var senderJID string

			// First get the WA Message ID from chatwoot_messages
			// Try finding by Chatwoot ID
			err := s.db.QueryRow(s.db.Rebind(`
				SELECT wa_message_id, COALESCE(chat_jid, ''), COALESCE(sender_jid, '')
				FROM chatwoot_messages
				WHERE user_id = ? AND chatwoot_message_id = ?
				LIMIT 1
			`), userID, payload.ID).Scan(&waMsgID, &remoteJID, &senderJID)

			// Fallback: Check SourceID if available (sometimes Chatwoot sends WAID:... as source_id)
			if err != nil && payload.SourceID != "" && strings.HasPrefix(payload.SourceID, "WAID") {
				waMsgID = strings.TrimPrefix(payload.SourceID, "WAID:")
				log.Info().Str("waMsgID", waMsgID).Msg("Resolved WA message ID from SourceID fallback")
				err = nil
			}

			if err != nil {
				log.Warn().Int("cwMsgID", payload.ID).Err(err).Msg("Failed to find WA message ID for deletion in database")
				s.Respond(w, r, http.StatusOK, map[string]string{"status": "ignored_unknown_message"})
				return
			}

			if remoteJID == "" {
				err = s.db.QueryRow(s.db.Rebind("SELECT remote_jid FROM chatwoot_conversations WHERE user_id = ? AND chatwoot_conversation_id = ?"), userID, payload.Conversation.ID).Scan(&remoteJID)
			}
			if err != nil || remoteJID == "" {
				log.Warn().Int("convID", payload.Conversation.ID).Err(err).Msg("Failed to find remote JID for deletion mapping")
				s.Respond(w, r, http.StatusOK, map[string]string{"status": "ignored_unknown_conversation"})
				return
			}

			jid, err := watypes.ParseJID(remoteJID)
			if err != nil {
				log.Error().Err(err).Str("remoteJID", remoteJID).Msg("Invalid JID during deletion")
				return
			}

			// Revoke the message
			log.Info().Str("waMsgID", waMsgID).Str("jid", jid.String()).Msg("Revoking message on WhatsApp")
			if jid.Server == watypes.GroupServer && senderJID != "" {
				if parsedSender, parseErr := watypes.ParseJID(senderJID); parseErr == nil {
					_, err = client.SendMessage(r.Context(), jid, client.BuildRevoke(jid, parsedSender, waMsgID))
				} else {
					log.Warn().Err(parseErr).Str("senderJID", senderJID).Msg("Invalid sender JID for group revoke, falling back to own revoke")
					_, err = client.RevokeMessage(r.Context(), jid, waMsgID)
				}
			} else {
				_, err = client.RevokeMessage(r.Context(), jid, waMsgID)
			}
			if err != nil {
				log.Error().Err(err).Str("waMsgID", waMsgID).Msg("Failed to revoke message in WhatsApp")
				s.Respond(w, r, http.StatusOK, map[string]string{"status": "failed_revoke", "error": err.Error()})
			} else {
				log.Info().Str("waMsgID", waMsgID).Msg("Revoked message in WhatsApp successfully")

				// Optional: Remove from mapping or mark as deleted
				s.db.Exec(s.db.Rebind("DELETE FROM chatwoot_messages WHERE user_id = ? AND wa_message_id = ?"), userID, waMsgID)

				s.Respond(w, r, http.StatusOK, map[string]string{"status": "revoked"})
			}
			return
		}

		// Handle label synchronization from Chatwoot to WhatsApp
		if payload.Event == "conversation_updated" {
			for _, changedAttr := range payload.ChangedAttributes {
				labelsChange, ok := changedAttr["label_list"].(map[string]interface{})
				if !ok {
					continue
				}

				previousArr, _ := labelsChange["previous"].([]interface{})
				currentArr, _ := labelsChange["current"].([]interface{})

				// Convert to string slices
				prevLabels := make([]string, 0, len(previousArr))
				for _, l := range previousArr {
					if ls, ok := l.(string); ok {
						prevLabels = append(prevLabels, ls)
					}
				}
				currLabels := make([]string, 0, len(currentArr))
				for _, l := range currentArr {
					if ls, ok := l.(string); ok {
						currLabels = append(currLabels, ls)
					}
				}

				// Find added and removed labels
				added := []string{}
				for _, c := range currLabels {
					found := false
					for _, p := range prevLabels {
						if c == p {
							found = true
							break
						}
					}
					if !found {
						added = append(added, c)
					}
				}

				removed := []string{}
				for _, p := range prevLabels {
					found := false
					for _, c := range currLabels {
						if p == c {
							found = true
							break
						}
					}
					if !found {
						removed = append(removed, p)
					}
				}

				if len(added) > 0 || len(removed) > 0 {
					log.Info().
						Interface("added", added).
						Interface("removed", removed).
						Int("conversationID", payload.Conversation.ID).
						Msg("Labels changed in Chatwoot, syncing to WhatsApp")

					// Get WhatsApp client
					client := clientManager.GetWhatsmeowClient(userID)
					if client == nil || !client.IsLoggedIn() {
						continue
					}

					// Get remote JID
					var remoteJID string
					err := s.db.QueryRow(s.db.Rebind(`
						SELECT remote_jid 
						FROM chatwoot_conversations 
						WHERE user_id = ? AND chatwoot_conversation_id = ?
					`), userID, payload.Conversation.ID).Scan(&remoteJID)

					if err != nil {
						continue
					}

					jid, err := watypes.ParseJID(remoteJID)
					if err != nil {
						continue
					}

					// Process additions
					for _, labelName := range added {
						waLabelName := normalizeChatwootLabelForWhatsApp(labelName)
						labelID, created, err := s.getOrCreateWhatsAppLabelID(userID, waLabelName)
						if err != nil {
							log.Warn().Err(err).Str("labelName", labelName).Msg("Failed to resolve WhatsApp label")
							continue
						}

						var exists bool
						errCheck := s.db.QueryRow(s.db.Rebind("SELECT EXISTS(SELECT 1 FROM chatwoot_conversation_labels WHERE user_id = ? AND remote_jid = ? AND label_id = ?)"), userID, remoteJID, labelID).Scan(&exists)
						if errCheck == nil && exists && !created {
							log.Debug().Str("labelName", labelName).Msg("Label addition echo from Chatwoot ignored (already in DB)")
							continue
						}

						log.Info().Str("labelID", labelID).Str("jid", jid.String()).Msg("Queueing label addition for WhatsApp")
						q := getOrCreateLabelQueue(userID)
						q <- labelQueueItem{
							client:      client,
							jid:         jid,
							labelID:     labelID,
							labelName:   waLabelName,
							ensureLabel: created,
							add:         true,
						}
						s.setLocalConversationLabel(userID, remoteJID, labelID, true)
					}

					// Process removals
					for _, labelName := range removed {
						waLabelName := normalizeChatwootLabelForWhatsApp(labelName)
						labelID, _, err := s.getOrCreateWhatsAppLabelID(userID, waLabelName)
						if err != nil {
							log.Warn().Err(err).Str("labelName", labelName).Msg("Failed to resolve WhatsApp label for removal")
							continue
						}

						var exists bool
						errCheck := s.db.QueryRow(s.db.Rebind("SELECT EXISTS(SELECT 1 FROM chatwoot_conversation_labels WHERE user_id = ? AND remote_jid = ? AND label_id = ?)"), userID, remoteJID, labelID).Scan(&exists)
						if errCheck == nil && !exists {
							log.Debug().Str("labelName", labelName).Msg("Label removal echo from Chatwoot ignored (already removed from DB)")
							continue
						}

						log.Info().Str("labelID", labelID).Str("jid", jid.String()).Msg("Queueing label removal for WhatsApp")
						q := getOrCreateLabelQueue(userID)
						q <- labelQueueItem{
							client:  client,
							jid:     jid,
							labelID: labelID,
							add:     false,
						}
						s.setLocalConversationLabel(userID, remoteJID, labelID, false)
					}
				}
			}
			s.Respond(w, r, http.StatusOK, map[string]string{"status": "success"})
			return
		}

		// Only process outgoing messages from agents (not from contact)
		if payload.Event != "message_created" || payload.MessageType != "outgoing" {
			s.Respond(w, r, http.StatusOK, map[string]string{"status": "ignored"})
			return
		}

		// Ignore private messages
		if payload.Private {
			s.Respond(w, r, http.StatusOK, map[string]string{"status": "ignored_private"})
			return
		}

		// Ignore messages from contact (only process agent messages)
		if payload.Sender.Type == "contact" {
			s.Respond(w, r, http.StatusOK, map[string]string{"status": "ignored_contact"})
			return
		}

		// Ignore imported messages (synced from WhatsApp)
		if payload.ContentAttributes != nil {
			if imported, ok := payload.ContentAttributes["imported"]; ok {
				// Check if boolean true
				if val, ok := imported.(bool); ok && val {
					s.Respond(w, r, http.StatusOK, map[string]string{"status": "ignored_imported"})
					return
				}
				// Check if string "true"
				if val, ok := imported.(string); ok && val == "true" {
					s.Respond(w, r, http.StatusOK, map[string]string{"status": "ignored_imported"})
					return
				}
				log.Debug().Interface("imported", imported).Msg("Checked imported attribute")
			}
		}

		// Get WhatsApp client
		client := clientManager.GetWhatsmeowClient(userID)
		if client == nil {
			log.Error().Str("userID", userID).Msg("WhatsApp client not found")
			s.Respond(w, r, http.StatusNotFound, errors.New("client not found"))
			return
		}

		if !client.IsLoggedIn() {
			log.Error().Str("userID", userID).Msg("WhatsApp client not logged in")
			s.Respond(w, r, http.StatusBadRequest, errors.New("client not logged in"))
			return
		}

		// Get remote JID from conversation mapping
		var remoteJID string
		err = s.db.QueryRow(s.db.Rebind(`
			SELECT remote_jid 
			FROM chatwoot_conversations 
			WHERE user_id = ? AND chatwoot_conversation_id = ?
		`), userID, payload.Conversation.ID).Scan(&remoteJID)

		if err != nil {
			log.Warn().Err(err).Int("conversationID", payload.Conversation.ID).Msg("Conversation mapping not found, attempting to resolve")

			// If mapping not found, we attempt to resolve it by fetching details from Chatwoot
			// 1. Get Chatwoot Config for credentials
			var cwConfig struct {
				URL              string
				AccountID        string
				Token            string
				InboxID          int
				SendReadReceipts bool
			}
			errConfig := s.db.QueryRow(s.db.Rebind(`
				SELECT url, account_id, token, inbox_id, COALESCE(send_read_receipts, false)
				FROM chatwoot_config WHERE user_id = ?
			`), userID).Scan(&cwConfig.URL, &cwConfig.AccountID, &cwConfig.Token, &cwConfig.InboxID, &cwConfig.SendReadReceipts)

			if errConfig != nil {
				log.Error().Err(errConfig).Msg("Failed to get Chatwoot config for resolution")
				s.Respond(w, r, http.StatusNotFound, errors.New("conversation not set up"))
				return
			}

			// 2. Initialize temporary client
			cwClient := NewChatwootClient(cwConfig.URL, cwConfig.AccountID, cwConfig.Token)

			// 3. Resolve Contact ID from payload or API
			contactID := payload.Conversation.ContactID
			if contactID == 0 {
				contactID = payload.Conversation.ContactInbox.ContactID
			}
			if contactID == 0 {
				contactID = payload.Conversation.Meta.Sender.ID
			}

			inboxID := payload.Conversation.InboxID
			status := payload.Conversation.Status

			// If still 0, fetch conversation from API
			if contactID == 0 {
				log.Info().Int("convID", payload.Conversation.ID).Msg("Fetching conversation details from Chatwoot API")
				cwConv, errConv := cwClient.GetConversation(payload.Conversation.ID)
				if errConv == nil {
					contactID = cwConv.ContactID
					if contactID == 0 && cwConv.ContactInbox.ContactID != 0 {
						contactID = cwConv.ContactInbox.ContactID
					}
					if inboxID == 0 {
						inboxID = cwConv.InboxID
					}
					if status == "" {
						status = cwConv.Status
					}
				} else {
					log.Error().Err(errConv).Msg("Failed to fetch conversation from Chatwoot")
					s.Respond(w, r, http.StatusNotFound, errors.New("conversation not found in chatwoot"))
					return
				}
			}

			// 4. Fetch Contact
			cwContact, errContact := cwClient.GetContact(contactID)
			if errContact != nil {
				log.Error().Err(errContact).Int("contactID", contactID).Msg("Failed to fetch contact from Chatwoot")
				s.Respond(w, r, http.StatusNotFound, errors.New("contact not found"))
				return
			}

			if cwContact.PhoneNumber == "" && cwContact.Identifier == "" {
				log.Error().Msg("Chatwoot contact has no phone number or identifier, cannot resolve JID")
				s.Respond(w, r, http.StatusBadRequest, errors.New("contact missing identifier"))
				return
			}

			// Resolve remoteJID from identifier or phone
			remoteJID = cwContact.Identifier
			if remoteJID == "" {
				remoteJID = cwContact.PhoneNumber + "@s.whatsapp.net"
			}

			// Save mapping for future use
			if inboxID == 0 {
				inboxID = cwConfig.InboxID
			}
			if status == "" {
				status = "open"
			}

			_, err = s.db.Exec(s.db.Rebind(`
				INSERT INTO chatwoot_conversations (user_id, remote_jid, chatwoot_conversation_id, chatwoot_inbox_id, status)
				VALUES (?, ?, ?, ?, ?)
				ON CONFLICT (user_id, remote_jid) DO UPDATE SET chatwoot_conversation_id = EXCLUDED.chatwoot_conversation_id
			`), userID, remoteJID, payload.Conversation.ID, inboxID, status)
			if err != nil {
				log.Warn().Err(err).Msg("Failed to save resolved conversation mapping")
			}
			// 5. Resolve WhatsApp JID
			log.Info().Str("phone", cwContact.PhoneNumber).Msg("Checking if contact is on WhatsApp")

			// Clean phone number for check (remove + and spaces)
			cleanPhone := strings.ReplaceAll(cwContact.PhoneNumber, "+", "")
			cleanPhone = strings.ReplaceAll(cleanPhone, " ", "")
			cleanPhone = strings.ReplaceAll(cleanPhone, "-", "")

			// Check existence
			checkResp, errCheck := client.IsOnWhatsApp(r.Context(), []string{cleanPhone})

			var targetJID watypes.JID
			found := false

			if errCheck == nil && len(checkResp) > 0 {
				for _, res := range checkResp {
					if res.IsIn {
						targetJID = res.JID
						found = true
						break
					}
				}
			}

			if !found {
				// Fallback: Construct JID manually
				targetJID = watypes.NewJID(cleanPhone, watypes.DefaultUserServer)
				log.Info().Str("jid", targetJID.String()).Msg("Contact check failed or not found, using constructed JID")
			} else {
				log.Info().Str("jid", targetJID.String()).Msg("Contact found on WhatsApp")
			}

			// Ensure we never use AD (LID) JID for sending, always use Phone JID
			// user requested to avoid @lid
			if targetJID.Server == watypes.HiddenUserServer {
				// If we somehow got an LID, try to convert or warn?
				// Usually IsOnWhatsApp returns the phone JID.
				// If matched by LID, we might need to be careful.
				// But let's enforce non-AD just in case.
				targetJID.Server = watypes.DefaultUserServer // Force @s.whatsapp.net
			}
			targetJID = targetJID.ToNonAD()

			remoteJID = targetJID.String()

			// 6. Update Chatwoot Contact Identifier if needed
			if cwContact.Identifier != remoteJID {
				_, errUpdate := cwClient.UpdateContact(cwContact.ID, map[string]interface{}{
					"identifier": remoteJID,
				})
				if errUpdate != nil {
					log.Warn().Err(errUpdate).Msg("Failed to update Chatwoot contact identifier")
				} else {
					log.Info().Str("identifier", remoteJID).Msg("Updated Chatwoot contact identifier")
				}
			}

			// 7. Store Mappings (Contacts and Conversations)

			// Insert Contact Mapping
			go func() {
				_, err := s.db.Exec(s.db.Rebind(`
					INSERT INTO chatwoot_contacts (user_id, phone_number, chatwoot_contact_id, jid, is_group)
					VALUES (?, ?, ?, ?, ?)
					ON CONFLICT (user_id, phone_number) 
					DO UPDATE SET chatwoot_contact_id = EXCLUDED.chatwoot_contact_id, jid = EXCLUDED.jid
				`), userID, cleanPhone, cwContact.ID, remoteJID, false)
				if err != nil {
					log.Warn().Err(err).Msg("Failed to store resolved contact mapping")
				}
			}()

			// Insert Conversation Mapping
			// We need to know which inbox it belongs to, we have cwConfig.InboxID
			// But the conversation might be in a different inbox if we support multiple?
			// The webhook payload has `Conversation.ID`. Use that.
			_, errMap := s.db.Exec(s.db.Rebind(`
				INSERT INTO chatwoot_conversations (user_id, remote_jid, chatwoot_conversation_id, chatwoot_inbox_id, status, updated_at)
				VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
				ON CONFLICT (user_id, remote_jid) 
				DO UPDATE SET chatwoot_conversation_id = EXCLUDED.chatwoot_conversation_id, chatwoot_inbox_id = EXCLUDED.chatwoot_inbox_id, status = EXCLUDED.status, updated_at = CURRENT_TIMESTAMP
			`), userID, remoteJID, payload.Conversation.ID, cwConfig.InboxID, "open")

			if errMap != nil {
				log.Error().Err(errMap).Msg("Failed to store resolved conversation mapping")
				// We proceed anyway as we have the JID now
			} else {
				log.Info().Str("remoteJID", remoteJID).Int("convID", payload.Conversation.ID).Msg("Stored new conversation mapping")
			}

		}

		// Parse JID
		jid, err := watypes.ParseJID(remoteJID)
		if err != nil {
			log.Error().Err(err).Str("remoteJID", remoteJID).Msg("Failed to parse JID")
			s.Respond(w, r, http.StatusBadRequest, errors.New("invalid JID"))
			return
		}
		// Ensure it's a bare JID (no device part) for sending messages
		jid = jid.ToNonAD()

		// Send message to WhatsApp
		content := payload.Content

		// Handle attachments
		if len(payload.Attachments) > 0 {
			for _, attachment := range payload.Attachments {
				// Download media
				resp, err := http.Get(attachment.DataURL)
				if err != nil {
					log.Error().Err(err).Str("url", attachment.DataURL).Msg("Failed to download media")
					continue
				}
				defer resp.Body.Close()

				if resp.StatusCode != http.StatusOK {
					log.Error().Int("status", resp.StatusCode).Str("url", attachment.DataURL).Msg("Failed to download media")
					continue
				}

				contentType := resp.Header.Get("Content-Type")
				mediaFileType := attachment.FileType
				// If the fileType is generic or empty, use the actual Content-Type from headers
				if contentType != "" && (mediaFileType == "file" || mediaFileType == "" || mediaFileType == "image" || mediaFileType == "video" || mediaFileType == "audio") {
					if strings.Contains(contentType, "/") {
						mediaFileType = contentType
					}
				}

				// Read media data
				mediaData, err := io.ReadAll(resp.Body)
				if err != nil {
					log.Error().Err(err).Str("url", attachment.DataURL).Msg("Failed to read media data")
					continue
				}

				// log.Debug().Str("fileType", mediaFileType).Str("contentType", contentType).Str("url", attachment.DataURL).Msg("Preparing media for WhatsApp")

				// Determine message type based on file type
				var msg *waE2E.Message
				caption := content

				if mediaFileType == "image/webp" {
					// Handle as Sticker
					uploaded, err := client.Upload(context.Background(), mediaData, whatsmeow.MediaImage)
					if err != nil {
						log.Error().Err(err).Msg("Failed to upload sticker")
						continue
					}

					msg = &waE2E.Message{
						StickerMessage: &waE2E.StickerMessage{
							URL:           &uploaded.URL,
							DirectPath:    &uploaded.DirectPath,
							MediaKey:      uploaded.MediaKey,
							Mimetype:      &mediaFileType,
							FileEncSHA256: uploaded.FileEncSHA256,
							FileSHA256:    uploaded.FileSHA256,
							FileLength:    &uploaded.FileLength,
						},
					}
				} else if strings.HasPrefix(mediaFileType, "image") {
					// Ensure mimetype is valid
					if !strings.Contains(mediaFileType, "/") {
						mediaFileType = contentType
						if mediaFileType == "" {
							mediaFileType = http.DetectContentType(mediaData)
						}
					}
					uploaded, err := client.Upload(context.Background(), mediaData, whatsmeow.MediaImage)
					if err != nil {
						log.Error().Err(err).Msg("Failed to upload image")
						continue
					}

					msg = &waE2E.Message{
						ImageMessage: &waE2E.ImageMessage{
							Caption:       &caption,
							URL:           &uploaded.URL,
							DirectPath:    &uploaded.DirectPath,
							MediaKey:      uploaded.MediaKey,
							Mimetype:      &mediaFileType,
							FileEncSHA256: uploaded.FileEncSHA256,
							FileSHA256:    uploaded.FileSHA256,
							FileLength:    &uploaded.FileLength,
						},
					}
				} else if strings.HasPrefix(mediaFileType, "video") {
					uploaded, err := client.Upload(context.Background(), mediaData, whatsmeow.MediaVideo)
					if err != nil {
						log.Error().Err(err).Msg("Failed to upload video")
						continue
					}

					msg = &waE2E.Message{
						VideoMessage: &waE2E.VideoMessage{
							Caption:       &caption,
							URL:           &uploaded.URL,
							DirectPath:    &uploaded.DirectPath,
							MediaKey:      uploaded.MediaKey,
							Mimetype:      &mediaFileType,
							FileEncSHA256: uploaded.FileEncSHA256,
							FileSHA256:    uploaded.FileSHA256,
							FileLength:    &uploaded.FileLength,
							Seconds:       proto.Uint32(5),
						},
					}
				} else if strings.HasPrefix(mediaFileType, "audio") {
					uploaded, err := client.Upload(context.Background(), mediaData, whatsmeow.MediaAudio)
					if err != nil {
						log.Error().Err(err).Msg("Failed to upload audio")
						continue
					}
					isPTT := true
					msg = &waE2E.Message{
						AudioMessage: &waE2E.AudioMessage{
							URL:           &uploaded.URL,
							DirectPath:    &uploaded.DirectPath,
							MediaKey:      uploaded.MediaKey,
							Mimetype:      &mediaFileType,
							FileEncSHA256: uploaded.FileEncSHA256,
							FileSHA256:    uploaded.FileSHA256,
							FileLength:    &uploaded.FileLength,
							PTT:           &isPTT,
						},
					}
				} else {
					// Document
					uploaded, err := client.Upload(context.Background(), mediaData, whatsmeow.MediaDocument)
					if err != nil {
						log.Error().Err(err).Msg("Failed to upload document")
						continue
					}
					fileName := "file"
					if parts := strings.Split(attachment.DataURL, "/"); len(parts) > 0 {
						possibleName := parts[len(parts)-1]
						if possibleName != "" && !strings.Contains(possibleName, "?") {
							fileName = possibleName
						}
					}
					msg = &waE2E.Message{
						DocumentMessage: &waE2E.DocumentMessage{
							Caption:       &caption,
							URL:           &uploaded.URL,
							DirectPath:    &uploaded.DirectPath,
							MediaKey:      uploaded.MediaKey,
							Mimetype:      &mediaFileType,
							FileEncSHA256: uploaded.FileEncSHA256,
							FileSHA256:    uploaded.FileSHA256,
							FileLength:    &uploaded.FileLength,
							FileName:      &fileName,
						},
					}
				}

				if msg != nil {
					sendResp, err := client.SendMessage(r.Context(), jid, msg)
					if err != nil {
						log.Error().Err(err).Msg("Failed to send media message")
					} else {
						s.storeChatwootOutgoingMapping(userID, remoteJID, payload.Conversation.ID, sendResp.ID, payload.ID)
					}
				}

				content = "" // Don't send text separately if we sent media with caption
			}
		}

		// Send text message if there's content
		// Send text message if there's content
		if content != "" {
			// Check for reply (in_reply_to)
			var stanzaID *string
			var isReaction bool
			var quotedFromMe bool
			var quotedParticipant string
			var cwMsgID int

			if payload.ContentAttributes != nil {
				if inReplyTo, ok := payload.ContentAttributes["in_reply_to"]; ok {
					// JSON numbers can be float64
					if idFloat, ok := inReplyTo.(float64); ok {
						cwMsgID = int(idFloat)
					} else if idInt, ok := inReplyTo.(int); ok {
						cwMsgID = idInt
					}

					if cwMsgID != 0 {
						var waMsgID string
						var direction string
						var senderJID sql.NullString
						err := s.db.QueryRow(s.db.Rebind("SELECT wa_message_id, direction, sender_jid FROM chatwoot_messages WHERE chatwoot_message_id = ? LIMIT 1"), cwMsgID).Scan(&waMsgID, &direction, &senderJID)
						if err == nil && waMsgID != "" {
							stanzaID = proto.String(waMsgID)
							quotedFromMe = (direction == "outgoing")
							if senderJID.Valid {
								quotedParticipant = senderJID.String
							}

							// Check if content is an emoji
							contentTrimmed := strings.TrimSpace(content)
							runes := []rune(contentTrimmed)
							// An emoji can be 1-4 runes typically, no alphanumeric
							if len(runes) > 0 && len(runes) <= 5 {
								hasAlphaNum := false
								for _, r := range runes {
									if unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.IsPunct(r) || unicode.IsSpace(r) {
										hasAlphaNum = true
										break
									}
								}
								if !hasAlphaNum {
									isReaction = true
									content = contentTrimmed // use exact trim for emoji
								}
							}
						}
					}
				}
			}

			// Check for commands in private notes (e.g. for reactions or delete if webhook supports it)
			if payload.Private {
				// We generally ignore private messages, but we could use them for commands
			}

			var msg *waE2E.Message
			if isReaction {
				key := &waCommon.MessageKey{
					RemoteJID: proto.String(remoteJID),
					FromMe:    proto.Bool(quotedFromMe),
					ID:        stanzaID,
				}
				if !quotedFromMe && quotedParticipant != "" {
					key.Participant = proto.String(quotedParticipant)
				}

				msg = &waE2E.Message{
					ReactionMessage: &waE2E.ReactionMessage{
						Key:               key,
						Text:              proto.String(content),
						GroupingKey:       proto.String(content),
						SenderTimestampMS: proto.Int64(time.Now().UnixMilli()),
					},
				}
			} else {
				msg = &waE2E.Message{
					Conversation: proto.String(content),
				}

				if stanzaID != nil {
					msg = &waE2E.Message{
						ExtendedTextMessage: &waE2E.ExtendedTextMessage{
							Text: proto.String(content),
							ContextInfo: &waE2E.ContextInfo{
								StanzaID:    stanzaID,
								Participant: proto.String(remoteJID), // Quote the sender roughly
								QuotedMessage: &waE2E.Message{
									Conversation: proto.String("..."),
								},
							},
						},
					}
				}
			}

			resp, err := client.SendMessage(r.Context(), jid, msg)

			if err != nil {
				log.Error().Err(err).Msg("Failed to send WhatsApp message")
				s.Respond(w, r, http.StatusInternalServerError, errors.New("failed to send message"))
				return
			}

			s.storeChatwootOutgoingMapping(userID, remoteJID, payload.Conversation.ID, resp.ID, payload.ID)
		}

		log.Info().
			Str("remoteJID", remoteJID).
			Int("conversationID", payload.Conversation.ID).
			Msg("Sent message to WhatsApp from Chatwoot")

		s.Respond(w, r, http.StatusOK, map[string]string{"status": "sent"})
	}
}

// getServerURL returns the server URL from environment or default
func getServerURL() string {
	// Read from environment variable SERVER_URL
	if serverURL := os.Getenv("SERVER_URL"); serverURL != "" {
		return serverURL
	}
	// Fallback to localhost for development
	return "http://localhost:8080"
}

// SyncChatwootMessage helper to sync outgoing messages from API to Chatwoot
func (s *server) SyncChatwootMessage(userID, remoteJID, messageID, content, fileName string, fileData []byte, contentType string) {
	// Fetch Chatwoot Config for User
	var cwConfig ChatwootConfig
	err := s.db.QueryRow(s.db.Rebind(`
		SELECT user_id, enabled, account_id, token, url, inbox_id, inbox_name, 
			   sign_msg, sign_delimiter, reopen_conversation, conversation_pending, merge_brazil_contacts,
			   import_messages, send_status_stories, send_typing, send_read_receipts, enabled_at
		FROM chatwoot_config WHERE user_id = ?
	`), userID).Scan(
		&cwConfig.UserID, &cwConfig.Enabled, &cwConfig.AccountID, &cwConfig.Token, &cwConfig.URL,
		&cwConfig.InboxID, &cwConfig.InboxName, &cwConfig.SignMsg, &cwConfig.SignDelimiter,
		&cwConfig.ReopenConversation, &cwConfig.ConversationPending, &cwConfig.MergeBrazilContacts,
		&cwConfig.ImportGroups, &cwConfig.SendStatusStories, &cwConfig.SendTyping, &cwConfig.SendReadReceipts, &cwConfig.EnabledAt,
	)

	if err != nil {
		// Chatwoot not configured or error
		return
	}

	if !cwConfig.Enabled {
		return
	}

	// Create service
	svc := NewChatwootService(s.db, cwConfig, clientManager.GetWhatsmeowClient(userID))

	// Determine JID
	jid, err := watypes.ParseJID(remoteJID)
	if err != nil {
		log.Error().Err(err).Str("jid", remoteJID).Msg("Failed to parse JID for Chatwoot sync")
		return
	}

	phoneNumber := jid.User
	name := phoneNumber // Default name
	isGroup := jid.Server == "g.us"

	// Append source info
	content = fmt.Sprintf("%s\n\n(Enviado via API)", content)

	// Call SendMessageToChatwoot with fromMe=true
	go func() {
		err := svc.SendMessageToChatwoot(
			remoteJID,
			phoneNumber,
			name,
			content,
			"", // AvatarURL - let ChatwootService fetch if needed
			isGroup,
			true, // fromMe = TRUE to mark as Private Note
			messageID,
			fileName,
			fileData,
			contentType,
			"", // quotedMessageID
			"", // senderJID
		)
		if err != nil {
			log.Error().Err(err).Str("msgID", messageID).Msg("Failed to manual sync API message to Chatwoot")
		} else {
			log.Debug().Str("msgID", messageID).Msg("Manually synced API message to Chatwoot")
		}
	}()
}
