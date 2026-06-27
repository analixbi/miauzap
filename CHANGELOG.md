# Changelog

All notable changes to the "Miauzap" project will be documented in this file.


## [v1.0.12] - 2026-02-03
### Fixed
- **Chatwoot**: Resolved "message recipient must be a user JID with no device part" error by sanitizing the JID before sending messages or typing indicators.
- **Chatwoot**: Added source labeling for private notes: "(Enviado via API)" for API messages and "(Enviado via Celular)" for messages sent from the phone.

## [v1.0.11] - 2026-02-01
### Fixed
- **Chatwoot**: API messages sync to Chatwoot: Explicitly mark outgoing messages sent via HTTP API as private notes to prevent duplication and ensure consistency.
- **Chatwoot**: Updated SendMessage, SendImage, SendAudio, SendVideo, SendDocument, SendSticker handlers to trigger Chatwoot sync with `fromMe=true`.

## [v1.0.10] - 2026-02-01

### Changed
- **Chatwoot**: Outgoing messages synced from WhatsApp (via phone) are now created as **Private Notes** in Chatwoot. This ensures they are visible to agents but flagged as private, preventing the Chatwoot webhook from re-sending them to the customer and causing duplication.

## [v1.0.9] - 2026-02-01

### Fixed
- **Chatwoot**: Switched from `private_metadata` to `content_attributes` for flagging imported messages, as `content_attributes` is correctly propagated in Chatwoot webhooks.

## [v1.0.8] - 2026-02-01

### Fixed
- **Chatwoot**: Improved robustness of webhook handler to correctly identify and ignore `imported` messages whether the flag is passed as a boolean or string.
- **Logging**: Added detailed debug logging for Chatwoot webhook payloads to assist in diagnosing integration issues.

## [v1.0.7] - 2026-02-01

### Fixed
- **Chatwoot**: Resolved critical loop issue where outgoing messages synced from the mobile device were being re-sent to WhatsApp by the Chatwoot webhook handler, causing duplicates. Now uses `private_metadata` to tag and ignore synced messages.

## [v1.0.6] - 2026-02-01

### Fixed
- **Chatwoot**: Resolved issue where outgoing messages were creating new conversations with unknown IDs (LIDs). Now correctly resolves LIDs to phone numbers using `RecipientAlt` to ensure messages are linked to the correct contact.

## [v1.0.5] - 2026-02-01

### Fixed
- **Chatwoot**: Fixed outgoing message synchronization. Messages sent from the mobile device are now correctly attributed to the recipient contact in Chatwoot instead of creating a conversation with the sender (agent).

## [v1.0.4] - 2026-01-29

### Added
- **API**: Updated `SendList` (Buttons) to support simplified `choices` JSON format for easier menu creation.
- **API**: Updated `SendCarousel` to support media (`image`, `video`, `document`) directly in cards via URL or base64.
- **Documentation**: Updated API documentation (`API.md`) and Frontend documentation (`/docs`) for new endpoints.

## [v1.2.0] - 2026-01-26

### Added
- **Chatwoot**: Improved Group Chat support. Group messages now arrive in Chatwoot with participant names prepended (e.g. `[Name]: message`).
- **Chatwoot**: Group contacts are now correctly created using their JID as identifier, fixing the "Phone number should be in e164 format" validation error.

## [v1.1.6] - 2026-01-26

### Fixed
- **Backend**: Fixed a critical SQL syntax error in Chatwoot setup handler that prevented saving configuration if correctly migrated.
- **Backend**: Improved cross-database compatibility for Chatwoot status check.

## [v1.1.5] - 2026-01-26

### Fixed
- **Database**: Re-implemented migrations using robust Postgres `DO` blocks to handle "column exists" checks safely across all PG versions.
- **Database**: Fixed migration recording in SQLite (correct parameter placeholders).
- **Backend**: Resolved a variable shadowing bug in Chatwoot setup handler.

## [v1.1.4] - 2026-01-26

### Fixed
- **Database**: Added robust, forced migration logic for Chatwoot columns (`import_messages`, `enabled_at`) to resolve issues in certain PostgreSQL environments.
- **Logging**: Added detailed migration progress logging during startup.

## [v1.1.3] - 2026-01-26

### Fixed
- **Chatwoot**: Fixed bug where group message synchronization setting was ignored in the main event loop.
- **Chatwoot**: Implemented "No History" logic; only messages received *after* the integration is enabled will be forwarded to Chatwoot.

## [v1.1.2] - 2026-01-26

### Fixed
- **Database**: Added migration ID 11 to force creation of `import_messages` column, resolving issue where previous migration might have silently failed.

## [v1.1.1] - 2026-01-26

### Fixed
- **Database**: Fixed missing migration ID 9 and simplified migration ID 10 SQL to ensure `import_messages` column is created.

## [v1.1.0] - 2026-01-26

### Added
- **Interactive Messages**: Added full support for WhatsApp interactive messages including Buttons, Lists, and Carousels via `whatsmeow`.
- **Chatwoot Integration**:
    - Added "Import Group Messages" toggle to the configuration dashboard.
    - Implemented persistence for Chatwoot Account ID in the browser.
- **API Documentation**: Added `cURL` examples and "Copy to Clipboard" functionality to the API dashboard (`/api/docs`).

### Changed
- **Database**: Added `import_messages` column to `chatwoot_config` table (Migration #10).
- **Backend**: Updated `ChatwootService` to conditionally sync group messages based on configuration.

### Fixed
- **Codebase**: Removed duplicate struct fields (`reopenConv`, `convPending`) in `ChatwootService` to satisfy linters.

## [v1.0.0] - 2026-01-24

### Added
- Initial release of Miauzap (rebranded from WuzAPI).
- Docker support with `analixbi/miauzap` image.
