package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
)

type Migration struct {
	ID      int
	Name    string
	UpSQL   string
	DownSQL string
}

var migrations = []Migration{
	{
		ID:    1,
		Name:  "initial_schema",
		UpSQL: initialSchemaSQL,
	},
	{
		ID:   2,
		Name: "add_proxy_url",
		UpSQL: `
            -- PostgreSQL version
            DO $$
            BEGIN
                IF NOT EXISTS (
                    SELECT 1 FROM information_schema.columns 
                    WHERE table_name = 'users' AND column_name = 'proxy_url'
                ) THEN
                    ALTER TABLE users ADD COLUMN proxy_url TEXT DEFAULT '';
                END IF;
            END $$;
            
            -- SQLite version (handled in code)
            `,
	},
	{
		ID:    3,
		Name:  "change_id_to_string",
		UpSQL: changeIDToStringSQL,
	},
	{
		ID:    4,
		Name:  "add_s3_support",
		UpSQL: addS3SupportSQL,
	},
	{
		ID:    5,
		Name:  "add_message_history",
		UpSQL: addMessageHistorySQL,
	},
	{
		ID:    6,
		Name:  "add_quoted_message_id",
		UpSQL: addQuotedMessageIDSQL,
	},
	{
		ID:    7,
		Name:  "add_hmac_key",
		UpSQL: addHmacKeySQL,
	},
	{
		ID:    8,
		Name:  "add_data_json",
		UpSQL: addDataJsonSQL,
	},
	{
		ID:    9,
		Name:  "add_chatwoot_support",
		UpSQL: addChatwootSupportSQL,
	},
	{
		ID:    10,
		Name:  "add_chatwoot_import_groups",
		UpSQL: addChatwootImportGroupsSQL,
	},
	{
		ID:    11,
		Name:  "fix_chatwoot_import_groups",
		UpSQL: "-- Handled in code",
	},
	{
		ID:    12,
		Name:  "add_chatwoot_enabled_at",
		UpSQL: "-- Handled in code",
	},
	{
		ID:    13,
		Name:  "force_fix_chatwoot_columns_v2",
		UpSQL: "-- Handled in code",
	},
	{
		ID:    14,
		Name:  "add_chatwoot_auto_create",
		UpSQL: "-- Handled in code",
	},
	{
		ID:    15,
		Name:  "force_fix_chatwoot_columns_v3",
		UpSQL: "-- Handled in code",
	},
	{
		ID:    16,
		Name:  "add_sender_jid_to_chatwoot_messages",
		UpSQL: "-- Handled in code",
	},
	{
		ID:    17,
		Name:  "add_chat_jid_and_conversation_id_to_chatwoot_messages",
		UpSQL: "-- Handled in code",
	},
	{
		ID:    18,
		Name:  "add_language_to_users",
		UpSQL: "-- Handled in code",
	},
	{
		ID:    19,
		Name:  "add_chatwoot_labels_table",
		UpSQL: "-- Handled in code",
	},
	{
		ID:    20,
		Name:  "add_chatwoot_conversation_labels_table",
		UpSQL: "-- Handled in code",
	},
	{
		ID:    21,
		Name:  "add_chatwoot_send_typing_column",
		UpSQL: "-- Handled in code",
	},
	{
		ID:    22,
		Name:  "add_chatwoot_send_status_stories_column",
		UpSQL: "-- Handled in code",
	},
	{
		ID:    23,
		Name:  "add_chatwoot_send_read_receipts_column",
		UpSQL: "-- Handled in code",
	},
	{
		ID:    24,
		Name:  "add_chatwoot_existing_webhook_fields",
		UpSQL: "-- Handled in code",
	},
}

const changeIDToStringSQL = `
-- Migration to change ID from integer to random string
DO $$
BEGIN
    -- Only execute if the column is currently integer type
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'users' AND column_name = 'id' AND data_type = 'integer'
    ) THEN
        -- For PostgreSQL
        ALTER TABLE users ADD COLUMN new_id TEXT;
		UPDATE users SET new_id = md5(random()::text || id::text || clock_timestamp()::text);
		ALTER TABLE users DROP CONSTRAINT users_pkey;
        ALTER TABLE users DROP COLUMN id CASCADE;
        ALTER TABLE users RENAME COLUMN new_id TO id;
        ALTER TABLE users ALTER COLUMN id SET NOT NULL;
        ALTER TABLE users ADD PRIMARY KEY (id);
    END IF;
END $$;
`

const initialSchemaSQL = `
-- PostgreSQL version
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'users') THEN
        CREATE TABLE users (
            id TEXT PRIMARY KEY,
            name TEXT NOT NULL,
            token TEXT NOT NULL,
            webhook TEXT NOT NULL DEFAULT '',
            jid TEXT NOT NULL DEFAULT '',
            qrcode TEXT NOT NULL DEFAULT '',
            connected INTEGER,
            expiration INTEGER,
            events TEXT NOT NULL DEFAULT '',
            proxy_url TEXT DEFAULT '',
            language TEXT NOT NULL DEFAULT 'pt'
        );
    END IF;
END $$;

-- SQLite version (handled in code)
`

const addS3SupportSQL = `
-- PostgreSQL version
DO $$
BEGIN
    -- Add S3 configuration columns if they don't exist
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'users' AND column_name = 's3_enabled') THEN
        ALTER TABLE users ADD COLUMN s3_enabled BOOLEAN DEFAULT FALSE;
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'users' AND column_name = 's3_endpoint') THEN
        ALTER TABLE users ADD COLUMN s3_endpoint TEXT DEFAULT '';
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'users' AND column_name = 's3_region') THEN
        ALTER TABLE users ADD COLUMN s3_region TEXT DEFAULT '';
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'users' AND column_name = 's3_bucket') THEN
        ALTER TABLE users ADD COLUMN s3_bucket TEXT DEFAULT '';
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'users' AND column_name = 's3_access_key') THEN
        ALTER TABLE users ADD COLUMN s3_access_key TEXT DEFAULT '';
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'users' AND column_name = 's3_secret_key') THEN
        ALTER TABLE users ADD COLUMN s3_secret_key TEXT DEFAULT '';
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'users' AND column_name = 's3_path_style') THEN
        ALTER TABLE users ADD COLUMN s3_path_style BOOLEAN DEFAULT TRUE;
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'users' AND column_name = 's3_public_url') THEN
        ALTER TABLE users ADD COLUMN s3_public_url TEXT DEFAULT '';
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'users' AND column_name = 'media_delivery') THEN
        ALTER TABLE users ADD COLUMN media_delivery TEXT DEFAULT 'base64';
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'users' AND column_name = 's3_retention_days') THEN
        ALTER TABLE users ADD COLUMN s3_retention_days INTEGER DEFAULT 30;
    END IF;
END $$;
`

const addMessageHistorySQL = `
-- PostgreSQL version
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'message_history') THEN
        CREATE TABLE message_history (
            id SERIAL PRIMARY KEY,
            user_id TEXT NOT NULL,
            chat_jid TEXT NOT NULL,
            sender_jid TEXT NOT NULL,
            message_id TEXT NOT NULL,
            timestamp TIMESTAMP NOT NULL,
            message_type TEXT NOT NULL,
            text_content TEXT,
            media_link TEXT,
            UNIQUE(user_id, message_id)
        );
        CREATE INDEX idx_message_history_user_chat_timestamp ON message_history (user_id, chat_jid, timestamp DESC);
    END IF;
    
    -- Add history column to users table if it doesn't exist
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'users' AND column_name = 'history') THEN
        ALTER TABLE users ADD COLUMN history INTEGER DEFAULT 0;
    END IF;
END $$;
`

const addQuotedMessageIDSQL = `
-- PostgreSQL version
DO $$
BEGIN
    -- Add quoted_message_id column to message_history table if it doesn't exist
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'message_history' AND column_name = 'quoted_message_id') THEN
        ALTER TABLE message_history ADD COLUMN quoted_message_id TEXT;
    END IF;
END $$;
`

const addDataJsonSQL = `
-- PostgreSQL version
DO $$
BEGIN
    -- Add dataJson column to message_history table if it doesn't exist
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'message_history' AND column_name = 'datajson') THEN
        ALTER TABLE message_history ADD COLUMN datajson TEXT;
    END IF;
END $$;

-- SQLite version (handled in code)
`

// GenerateRandomID creates a random string ID
func GenerateRandomID() (string, error) {
	bytes := make([]byte, 16) // 128 bits
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random ID: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

// Initialize the database with migrations
func initializeSchema(db *sqlx.DB) error {
	// Create migrations table if it doesn't exist
	if err := createMigrationsTable(db); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get already applied migrations
	applied, err := getAppliedMigrations(db)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Apply missing migrations
	for _, migration := range migrations {
		if _, ok := applied[migration.ID]; !ok {
			log.Info().Int("id", migration.ID).Str("name", migration.Name).Msg("Applying database migration")
			if err := applyMigration(db, migration); err != nil {
				return fmt.Errorf("failed to apply migration %d: %w", migration.ID, err)
			}
			log.Info().Int("id", migration.ID).Msg("Migration applied successfully")
		}
	}

	return nil
}

func createMigrationsTable(db *sqlx.DB) error {
	var tableExists bool
	var err error

	switch db.DriverName() {
	case "postgres":
		err = db.Get(&tableExists, `
			SELECT EXISTS (
				SELECT 1 FROM information_schema.tables 
				WHERE table_name = 'migrations'
			)`)
	case "sqlite":
		err = db.Get(&tableExists, `
			SELECT EXISTS (
				SELECT 1 FROM sqlite_master 
				WHERE type='table' AND name='migrations'
			)`)
	default:
		return fmt.Errorf("unsupported database driver: %s", db.DriverName())
	}

	if err != nil {
		return fmt.Errorf("failed to check migrations table existence: %w", err)
	}

	if tableExists {
		return nil
	}

	_, err = db.Exec(`
		CREATE TABLE migrations (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`)
	if err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	return nil
}

func getAppliedMigrations(db *sqlx.DB) (map[int]struct{}, error) {
	applied := make(map[int]struct{})
	var rows []struct {
		ID   int    `db:"id"`
		Name string `db:"name"`
	}

	err := db.Select(&rows, "SELECT id, name FROM migrations ORDER BY id ASC")
	if err != nil {
		return nil, fmt.Errorf("failed to query applied migrations: %w", err)
	}

	for _, row := range rows {
		applied[row.ID] = struct{}{}
	}

	return applied, nil
}

func applyMigration(db *sqlx.DB, migration Migration) error {
	tx, err := db.Beginx()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	if migration.ID == 1 {
		// Handle initial schema creation differently per database
		if db.DriverName() == "sqlite" {
			err = createTableIfNotExistsSQLite(tx, "users", `
                CREATE TABLE users (
                    id TEXT PRIMARY KEY,
                    name TEXT NOT NULL,
                    token TEXT NOT NULL,
                    webhook TEXT NOT NULL DEFAULT '',
                    jid TEXT NOT NULL DEFAULT '',
                    qrcode TEXT NOT NULL DEFAULT '',
                    connected INTEGER,
                    expiration INTEGER,
                    events TEXT NOT NULL DEFAULT '',
                    proxy_url TEXT DEFAULT '',
                    language TEXT NOT NULL DEFAULT 'pt'
                )`)
		} else {
			_, err = tx.Exec(migration.UpSQL)
		}
	} else if migration.ID == 2 {
		if db.DriverName() == "sqlite" {
			err = addColumnIfNotExistsSQLite(tx, "users", "proxy_url", "TEXT DEFAULT ''")
		} else {
			_, err = tx.Exec(migration.UpSQL)
		}
	} else if migration.ID == 3 {
		if db.DriverName() == "sqlite" {
			err = migrateSQLiteIDToString(tx)
		} else {
			_, err = tx.Exec(migration.UpSQL)
		}
	} else if migration.ID == 4 {
		if db.DriverName() == "sqlite" {
			// Handle S3 columns for SQLite
			err = addColumnIfNotExistsSQLite(tx, "users", "s3_enabled", "BOOLEAN DEFAULT 0")
			if err == nil {
				err = addColumnIfNotExistsSQLite(tx, "users", "s3_endpoint", "TEXT DEFAULT ''")
			}
			if err == nil {
				err = addColumnIfNotExistsSQLite(tx, "users", "s3_region", "TEXT DEFAULT ''")
			}
			if err == nil {
				err = addColumnIfNotExistsSQLite(tx, "users", "s3_bucket", "TEXT DEFAULT ''")
			}
			if err == nil {
				err = addColumnIfNotExistsSQLite(tx, "users", "s3_access_key", "TEXT DEFAULT ''")
			}
			if err == nil {
				err = addColumnIfNotExistsSQLite(tx, "users", "s3_secret_key", "TEXT DEFAULT ''")
			}
			if err == nil {
				err = addColumnIfNotExistsSQLite(tx, "users", "s3_path_style", "BOOLEAN DEFAULT 1")
			}
			if err == nil {
				err = addColumnIfNotExistsSQLite(tx, "users", "s3_public_url", "TEXT DEFAULT ''")
			}
			if err == nil {
				err = addColumnIfNotExistsSQLite(tx, "users", "media_delivery", "TEXT DEFAULT 'base64'")
			}
			if err == nil {
				err = addColumnIfNotExistsSQLite(tx, "users", "s3_retention_days", "INTEGER DEFAULT 30")
			}
		} else {
			_, err = tx.Exec(migration.UpSQL)
		}
	} else if migration.ID == 5 {
		if db.DriverName() == "sqlite" {
			// Handle message_history table creation for SQLite
			err = createTableIfNotExistsSQLite(tx, "message_history", `
				CREATE TABLE message_history (
					id INTEGER PRIMARY KEY AUTOINCREMENT,
					user_id TEXT NOT NULL,
					chat_jid TEXT NOT NULL,
					sender_jid TEXT NOT NULL,
					message_id TEXT NOT NULL,
					timestamp DATETIME NOT NULL,
					message_type TEXT NOT NULL,
					text_content TEXT,
					media_link TEXT,
					UNIQUE(user_id, message_id)
				)`)
			if err == nil {
				// Create index for SQLite
				_, err = tx.Exec(`
					CREATE INDEX IF NOT EXISTS idx_message_history_user_chat_timestamp 
					ON message_history (user_id, chat_jid, timestamp DESC)`)
			}
			if err == nil {
				// Add history column to users table
				err = addColumnIfNotExistsSQLite(tx, "users", "history", "INTEGER DEFAULT 0")
			}
		} else {
			_, err = tx.Exec(migration.UpSQL)
		}
	} else if migration.ID == 6 {
		if db.DriverName() == "sqlite" {
			// Add quoted_message_id column to message_history table for SQLite
			err = addColumnIfNotExistsSQLite(tx, "message_history", "quoted_message_id", "TEXT")
		} else {
			_, err = tx.Exec(migration.UpSQL)
		}
	} else if migration.ID == 7 {
		if db.DriverName() == "sqlite" {
			// Add hmac_key column as BLOB for encrypted data in SQLite
			err = addColumnIfNotExistsSQLite(tx, "users", "hmac_key", "BLOB")
		} else {
			_, err = tx.Exec(migration.UpSQL)
		}
	} else if migration.ID == 8 {
		if db.DriverName() == "sqlite" {
			// Add dataJson column to message_history table for SQLite
			err = addColumnIfNotExistsSQLite(tx, "message_history", "datajson", "TEXT")
		} else {
			_, err = tx.Exec(migration.UpSQL)
		}
	} else if migration.ID == 9 {
		if db.DriverName() == "sqlite" {
			// Create Chatwoot tables for SQLite
			err = createChatwootTablesSQLite(tx)
		} else {
			_, err = tx.Exec(migration.UpSQL)
		}
	} else if migration.ID >= 10 && migration.ID <= 15 {
		if db.DriverName() == "sqlite" {
			err = addColumnIfNotExistsSQLite(tx, "chatwoot_config", "import_messages", "BOOLEAN DEFAULT 0")
			if err == nil {
				err = addColumnIfNotExistsSQLite(tx, "chatwoot_config", "enabled_at", "TIMESTAMP")
			}
		} else {
			// PostgreSQL - Use robust DO block
			query := `
			DO $$
			BEGIN
				IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'chatwoot_config' AND column_name = 'import_messages') THEN
					ALTER TABLE chatwoot_config ADD COLUMN import_messages BOOLEAN DEFAULT FALSE;
				END IF;
				
				IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'chatwoot_config' AND column_name = 'enabled_at') THEN
					ALTER TABLE chatwoot_config ADD COLUMN enabled_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP;
				END IF;
			END $$;`
			_, err = tx.Exec(query)
		}
	} else if migration.ID == 16 {
		if db.DriverName() == "sqlite" {
			err = addColumnIfNotExistsSQLite(tx, "chatwoot_messages", "sender_jid", "TEXT")
		} else {
			// PostgreSQL
			query := `
			DO $$
			BEGIN
				IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'chatwoot_messages' AND column_name = 'sender_jid') THEN
					ALTER TABLE chatwoot_messages ADD COLUMN sender_jid TEXT;
				END IF;
			END $$;`
			_, err = tx.Exec(query)
		}
	} else if migration.ID == 17 {
		if db.DriverName() == "sqlite" {
			err = addColumnIfNotExistsSQLite(tx, "chatwoot_messages", "chat_jid", "TEXT")
			if err == nil {
				err = addColumnIfNotExistsSQLite(tx, "chatwoot_messages", "chatwoot_conversation_id", "INTEGER")
			}
		} else {
			// PostgreSQL
			query := `
			DO $$
			BEGIN
				IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'chatwoot_messages' AND column_name = 'chat_jid') THEN
					ALTER TABLE chatwoot_messages ADD COLUMN chat_jid TEXT;
				END IF;
				IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'chatwoot_messages' AND column_name = 'chatwoot_conversation_id') THEN
					ALTER TABLE chatwoot_messages ADD COLUMN chatwoot_conversation_id INTEGER;
				END IF;
			END $$;`
			_, err = tx.Exec(query)
		}
	} else if migration.ID == 18 {
		if db.DriverName() == "sqlite" {
			err = addColumnIfNotExistsSQLite(tx, "users", "language", "TEXT NOT NULL DEFAULT 'pt'")
		} else {
			// PostgreSQL
			query := `
			DO $$
			BEGIN
				IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'users' AND column_name = 'language') THEN
					ALTER TABLE users ADD COLUMN language TEXT NOT NULL DEFAULT 'pt';
				END IF;
			END $$;`
			_, err = tx.Exec(query)
		}
	} else if migration.ID == 19 {
		query := `
			CREATE TABLE IF NOT EXISTS chatwoot_labels (
				user_id TEXT,
				label_id TEXT,
				name TEXT,
				PRIMARY KEY (user_id, label_id),
				FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
			)`
		_, err = tx.Exec(query)
	} else if migration.ID == 20 {
		query := `
			CREATE TABLE IF NOT EXISTS chatwoot_conversation_labels (
				user_id TEXT,
				remote_jid TEXT,
				label_id TEXT,
				PRIMARY KEY (user_id, remote_jid, label_id),
				FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
			)`
		_, err = tx.Exec(query)
	} else if migration.ID == 21 {
		if db.DriverName() == "sqlite" {
			err = addColumnIfNotExistsSQLite(tx, "chatwoot_config", "send_typing", "BOOLEAN DEFAULT 1")
		} else {
			// PostgreSQL
			query := `
			DO $$
			BEGIN
				IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'chatwoot_config' AND column_name = 'send_typing') THEN
					ALTER TABLE chatwoot_config ADD COLUMN send_typing BOOLEAN DEFAULT TRUE;
				END IF;
			END $$;`
			_, err = tx.Exec(query)
		}
	} else if migration.ID == 22 {
		if db.DriverName() == "sqlite" {
			err = addColumnIfNotExistsSQLite(tx, "chatwoot_config", "send_status_stories", "BOOLEAN DEFAULT 0")
		} else {
			// PostgreSQL
			query := `
			DO $$
			BEGIN
				IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'chatwoot_config' AND column_name = 'send_status_stories') THEN
					ALTER TABLE chatwoot_config ADD COLUMN send_status_stories BOOLEAN DEFAULT FALSE;
				END IF;
			END $$;`
			_, err = tx.Exec(query)
		}
	} else if migration.ID == 23 {
		if db.DriverName() == "sqlite" {
			err = addColumnIfNotExistsSQLite(tx, "chatwoot_config", "send_read_receipts", "BOOLEAN DEFAULT 0")
		} else {
			// PostgreSQL
			query := `
			DO $$
			BEGIN
				IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'chatwoot_config' AND column_name = 'send_read_receipts') THEN
					ALTER TABLE chatwoot_config ADD COLUMN send_read_receipts BOOLEAN DEFAULT FALSE;
				END IF;
			END $$;`
			_, err = tx.Exec(query)
		}
	} else if migration.ID == 24 {
		if db.DriverName() == "sqlite" {
			err = addColumnIfNotExistsSQLite(tx, "chatwoot_config", "webhook_url", "TEXT DEFAULT ''")
			if err == nil {
				err = addColumnIfNotExistsSQLite(tx, "chatwoot_config", "webhook_secret", "TEXT DEFAULT ''")
			}
		} else {
			query := `
			DO $$
			BEGIN
				IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'chatwoot_config' AND column_name = 'webhook_url') THEN
					ALTER TABLE chatwoot_config ADD COLUMN webhook_url TEXT DEFAULT '';
				END IF;
				IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'chatwoot_config' AND column_name = 'webhook_secret') THEN
					ALTER TABLE chatwoot_config ADD COLUMN webhook_secret TEXT DEFAULT '';
				END IF;
			END $$;`
			_, err = tx.Exec(query)
		}
	} else {
		_, err = tx.Exec(migration.UpSQL)
	}

	if err != nil {
		return fmt.Errorf("failed to execute migration SQL: %w", err)
	}

	// Record the migration
	placeholder := "$1, $2"
	if db.DriverName() == "sqlite" {
		placeholder = "?, ?"
	}
	if _, err = tx.Exec(fmt.Sprintf(`
        INSERT INTO migrations (id, name) 
        VALUES (%s)`, placeholder), migration.ID, migration.Name); err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}

	return tx.Commit()
}

func createTableIfNotExistsSQLite(tx *sqlx.Tx, tableName, createSQL string) error {
	var exists int
	err := tx.Get(&exists, `
        SELECT COUNT(*) FROM sqlite_master
        WHERE type='table' AND name=?`, tableName)
	if err != nil {
		return err
	}

	if exists == 0 {
		_, err = tx.Exec(createSQL)
		return err
	}
	return nil
}
func sqliteChangeIDType(tx *sqlx.Tx) error {
	// SQLite requires a more complex approach:
	// 1. Create new table with string ID
	// 2. Copy data with new UUIDs
	// 3. Drop old table
	// 4. Rename new table

	// Step 1: Get the current schema
	var tableInfo string
	err := tx.Get(&tableInfo, `
        SELECT sql FROM sqlite_master
        WHERE type='table' AND name='users'`)
	if err != nil {
		return fmt.Errorf("failed to get table info: %w", err)
	}

	// Step 2: Create new table with string ID
	newTableSQL := strings.Replace(tableInfo,
		"CREATE TABLE users (",
		"CREATE TABLE users_new (id TEXT PRIMARY KEY, ", 1)
	newTableSQL = strings.Replace(newTableSQL,
		"id INTEGER PRIMARY KEY AUTOINCREMENT,", "", 1)

	if _, err = tx.Exec(newTableSQL); err != nil {
		return fmt.Errorf("failed to create new table: %w", err)
	}

	// Step 3: Copy data with new UUIDs
	columns, err := getTableColumns(tx, "users")
	if err != nil {
		return fmt.Errorf("failed to get table columns: %w", err)
	}

	// Remove 'id' from columns list
	var filteredColumns []string
	for _, col := range columns {
		if col != "id" {
			filteredColumns = append(filteredColumns, col)
		}
	}

	columnList := strings.Join(filteredColumns, ", ")
	if _, err = tx.Exec(fmt.Sprintf(`
        INSERT INTO users_new (id, %s)
        SELECT gen_random_uuid(), %s FROM users`,
		columnList, columnList)); err != nil {
		return fmt.Errorf("failed to copy data: %w", err)
	}

	// Step 4: Drop old table
	if _, err = tx.Exec("DROP TABLE users"); err != nil {
		return fmt.Errorf("failed to drop old table: %w", err)
	}

	// Step 5: Rename new table
	if _, err = tx.Exec("ALTER TABLE users_new RENAME TO users"); err != nil {
		return fmt.Errorf("failed to rename table: %w", err)
	}

	return nil
}

func getTableColumns(tx *sqlx.Tx, tableName string) ([]string, error) {
	var columns []string
	rows, err := tx.Query(fmt.Sprintf("PRAGMA table_info(%s)", tableName))
	if err != nil {
		return nil, fmt.Errorf("failed to get table info: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var name, typ string
		var notnull int
		var dfltValue interface{}
		var pk int

		if err := rows.Scan(&cid, &name, &typ, &notnull, &dfltValue, &pk); err != nil {
			return nil, fmt.Errorf("failed to scan column info: %w", err)
		}
		columns = append(columns, name)
	}

	return columns, nil
}

func migrateSQLiteIDToString(tx *sqlx.Tx) error {
	// 1. Check if we need to do the migration
	var currentType string
	err := tx.QueryRow(`
        SELECT type FROM pragma_table_info('users')
        WHERE name = 'id'`).Scan(&currentType)
	if err != nil {
		return fmt.Errorf("failed to check column type: %w", err)
	}

	if currentType != "INTEGER" {
		// No migration needed
		return nil
	}

	// 2. Create new table with string ID
	_, err = tx.Exec(`
        CREATE TABLE users_new (
            id TEXT PRIMARY KEY,
            name TEXT NOT NULL,
            token TEXT NOT NULL,
            webhook TEXT NOT NULL DEFAULT '',
            jid TEXT NOT NULL DEFAULT '',
            qrcode TEXT NOT NULL DEFAULT '',
            connected INTEGER,
            expiration INTEGER,
            events TEXT NOT NULL DEFAULT '',
            proxy_url TEXT DEFAULT '',
            language TEXT NOT NULL DEFAULT 'pt'
        )`)
	if err != nil {
		return fmt.Errorf("failed to create new table: %w", err)
	}

	// 3. Copy data with new UUIDs
	_, err = tx.Exec(`
        INSERT INTO users_new
        SELECT
            hex(randomblob(16)),
            name, token, webhook, jid, qrcode,
            connected, expiration, events, proxy_url, 'pt'
        FROM users`)
	if err != nil {
		return fmt.Errorf("failed to copy data: %w", err)
	}

	// 4. Drop old table
	_, err = tx.Exec(`DROP TABLE users`)
	if err != nil {
		return fmt.Errorf("failed to drop old table: %w", err)
	}

	// 5. Rename new table
	_, err = tx.Exec(`ALTER TABLE users_new RENAME TO users`)
	if err != nil {
		return fmt.Errorf("failed to rename table: %w", err)
	}

	return nil
}

func addColumnIfNotExistsSQLite(tx *sqlx.Tx, tableName, columnName, columnDef string) error {
	var exists int
	err := tx.Get(&exists, `
        SELECT COUNT(*) FROM pragma_table_info(?)
        WHERE name = ?`, tableName, columnName)
	if err != nil {
		return fmt.Errorf("failed to check column existence: %w", err)
	}

	if exists == 0 {
		_, err = tx.Exec(fmt.Sprintf(
			"ALTER TABLE %s ADD COLUMN %s %s",
			tableName, columnName, columnDef))
		if err != nil {
			return fmt.Errorf("failed to add column: %w", err)
		}
	}
	return nil
}

func createChatwootTablesSQLite(tx *sqlx.Tx) error {
	// Create chatwoot_config table
	err := createTableIfNotExistsSQLite(tx, "chatwoot_config", `
		CREATE TABLE chatwoot_config (
			user_id TEXT PRIMARY KEY,
			enabled BOOLEAN DEFAULT 0,
			account_id TEXT,
			token TEXT,
			url TEXT,
			inbox_id INTEGER,
			inbox_name TEXT,
			webhook_url TEXT DEFAULT '',
			webhook_secret TEXT DEFAULT '',
			sign_msg BOOLEAN DEFAULT 0,
			sign_delimiter TEXT,
			reopen_conversation BOOLEAN DEFAULT 0,
			conversation_pending BOOLEAN DEFAULT 0,
			merge_brazil_contacts BOOLEAN DEFAULT 0,
			send_status_stories BOOLEAN DEFAULT 0,
			send_typing BOOLEAN DEFAULT 1,
			send_read_receipts BOOLEAN DEFAULT 0,
			enabled_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`)
	if err != nil {
		return err
	}

	// Create chatwoot_contacts table
	err = createTableIfNotExistsSQLite(tx, "chatwoot_contacts", `
		CREATE TABLE chatwoot_contacts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id TEXT NOT NULL,
			phone_number TEXT NOT NULL,
			chatwoot_contact_id INTEGER NOT NULL,
			jid TEXT,
			is_group BOOLEAN DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(user_id, phone_number),
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`CREATE INDEX IF NOT EXISTS idx_chatwoot_contacts_user_phone 
		ON chatwoot_contacts(user_id, phone_number)`)
	if err != nil {
		return err
	}

	// Create chatwoot_conversations table
	err = createTableIfNotExistsSQLite(tx, "chatwoot_conversations", `
		CREATE TABLE chatwoot_conversations (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id TEXT NOT NULL,
			remote_jid TEXT NOT NULL,
			chatwoot_conversation_id INTEGER NOT NULL,
			chatwoot_inbox_id INTEGER NOT NULL,
			status TEXT DEFAULT 'open',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(user_id, remote_jid),
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`CREATE INDEX IF NOT EXISTS idx_chatwoot_conversations_user_jid 
		ON chatwoot_conversations(user_id, remote_jid)`)
	if err != nil {
		return err
	}

	// Create chatwoot_messages table
	err = createTableIfNotExistsSQLite(tx, "chatwoot_messages", `
		CREATE TABLE chatwoot_messages (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id TEXT NOT NULL,
			wa_message_id TEXT NOT NULL,
			chatwoot_message_id INTEGER,
			direction TEXT NOT NULL,
			synced BOOLEAN DEFAULT 0,
			sender_jid TEXT,
			chat_jid TEXT,
			chatwoot_conversation_id INTEGER,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(user_id, wa_message_id),
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`CREATE INDEX IF NOT EXISTS idx_chatwoot_messages_user_wa_id 
		ON chatwoot_messages(user_id, wa_message_id)`)

	return err
}

const addHmacKeySQL = `
-- PostgreSQL version - Add encrypted HMAC key column
DO $$
BEGIN
    -- Add hmac_key column as BYTEA for encrypted data
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'users' AND column_name = 'hmac_key') THEN
        ALTER TABLE users ADD COLUMN hmac_key BYTEA;
    END IF;
END $$;

-- SQLite version (handled in code)
`

const addChatwootSupportSQL = `
-- PostgreSQL version - Add Chatwoot integration tables
DO $$
BEGIN
    -- Create chatwoot_config table
    IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'chatwoot_config') THEN
        CREATE TABLE chatwoot_config (
            user_id TEXT PRIMARY KEY,
            enabled BOOLEAN DEFAULT FALSE,
            account_id TEXT,
            token TEXT,
            url TEXT,
            inbox_id INTEGER,
            inbox_name TEXT,
            webhook_url TEXT DEFAULT '',
            webhook_secret TEXT DEFAULT '',
            sign_msg BOOLEAN DEFAULT FALSE,
            sign_delimiter TEXT,
            reopen_conversation BOOLEAN DEFAULT FALSE,
            conversation_pending BOOLEAN DEFAULT FALSE,
            merge_brazil_contacts BOOLEAN DEFAULT FALSE,
            send_status_stories BOOLEAN DEFAULT FALSE,
            send_typing BOOLEAN DEFAULT TRUE,
            send_read_receipts BOOLEAN DEFAULT FALSE,
            enabled_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
        );
    END IF;

    -- Create chatwoot_contacts table
    IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'chatwoot_contacts') THEN
        CREATE TABLE chatwoot_contacts (
            id SERIAL PRIMARY KEY,
            user_id TEXT NOT NULL,
            phone_number TEXT NOT NULL,
            chatwoot_contact_id INTEGER NOT NULL,
            jid TEXT,
            is_group BOOLEAN DEFAULT FALSE,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            UNIQUE(user_id, phone_number),
            FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
        );
        CREATE INDEX idx_chatwoot_contacts_user_phone ON chatwoot_contacts(user_id, phone_number);
    END IF;

    -- Create chatwoot_conversations table
    IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'chatwoot_conversations') THEN
        CREATE TABLE chatwoot_conversations (
            id SERIAL PRIMARY KEY,
            user_id TEXT NOT NULL,
            remote_jid TEXT NOT NULL,
            chatwoot_conversation_id INTEGER NOT NULL,
            chatwoot_inbox_id INTEGER NOT NULL,
            status TEXT DEFAULT 'open',
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            UNIQUE(user_id, remote_jid),
            FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
        );
        CREATE INDEX idx_chatwoot_conversations_user_jid ON chatwoot_conversations(user_id, remote_jid);
    END IF;

    -- Create chatwoot_messages table
    IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'chatwoot_messages') THEN
        CREATE TABLE chatwoot_messages (
            id SERIAL PRIMARY KEY,
            user_id TEXT NOT NULL,
            wa_message_id TEXT NOT NULL,
            chatwoot_message_id INTEGER,
            direction TEXT NOT NULL,
            synced BOOLEAN DEFAULT FALSE,
            sender_jid TEXT,
            chat_jid TEXT,
            chatwoot_conversation_id INTEGER,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            UNIQUE(user_id, wa_message_id),
            FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
        );
        CREATE INDEX idx_chatwoot_messages_user_wa_id ON chatwoot_messages(user_id, wa_message_id);
    END IF;
END $$;

-- SQLite version (handled in code)
`

const addChatwootImportGroupsSQL = `
-- PostgreSQL version - Add import_messages column to chatwoot_config table
ALTER TABLE chatwoot_config ADD COLUMN IF NOT EXISTS import_messages BOOLEAN DEFAULT FALSE;
`

const addChatwootEnabledAtSQL = `
-- PostgreSQL version - Add enabled_at column to chatwoot_config table
ALTER TABLE chatwoot_config ADD COLUMN IF NOT EXISTS enabled_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP;
`
