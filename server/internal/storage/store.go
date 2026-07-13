package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

type Store struct {
	db *sql.DB
	mu sync.Mutex
}

type CacheEntry struct {
	Key       string
	Payload   string
	FetchedAt time.Time
	ExpiresAt time.Time
}

type Conversation struct {
	ID           string   `json:"id"`
	Mode         string   `json:"mode"`
	SourceTitle  string   `json:"sourceTitle,omitempty"`
	HotURL       string   `json:"hotUrl,omitempty"`
	CharacterIDs []string `json:"characterIds,omitempty"`
	Provider     string   `json:"provider,omitempty"`
	CreatedAt    int64    `json:"createdAt"`
	UpdatedAt    int64    `json:"updatedAt"`
}

type Message struct {
	ID             int64      `json:"id"`
	ConversationID string     `json:"conversationId"`
	Role           string     `json:"role"`
	CharacterID    string     `json:"characterId,omitempty"`
	CharacterName  string     `json:"characterName,omitempty"`
	Era            string     `json:"era,omitempty"`
	Round          int        `json:"round"`
	Content        string     `json:"content"`
	Provider       string     `json:"provider,omitempty"`
	Model          string     `json:"model,omitempty"`
	Citations      []Citation `json:"citations,omitempty"`
	CreatedAt      int64      `json:"createdAt"`
}

type Citation struct {
	Title   string `json:"title"`
	Source  string `json:"source,omitempty"`
	Excerpt string `json:"excerpt"`
}

func Open(dataDir string) (*Store, error) {
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return nil, fmt.Errorf("mkdir data dir: %w", err)
	}
	dbPath := filepath.Join(dataDir, "renwen.db")
	db, err := sql.Open("sqlite", dbPath+"?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_pragma=foreign_keys(ON)")
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := s.migrateMessageCitations(); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := s.importLegacyJSON(dataDir); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := s.seedHotListFixture(); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := s.migrateHotListMockSeeds(); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := s.migrateAggressiveCacheIntervals(); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := initSummaryMigration(s); err != nil {
		_ = db.Close()
		return nil, err
	}
	return s, nil
}

func (s *Store) Close() error {
	if s.db == nil {
		return nil
	}
	return s.db.Close()
}

func (s *Store) migrateMessageCitations() error {
	rows, err := s.db.Query(`PRAGMA table_info(messages)`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var id, notNull, pk int
		var name, typ string
		var def any
		if err := rows.Scan(&id, &name, &typ, &notNull, &def, &pk); err != nil {
			return err
		}
		if name == "citations_json" {
			return nil
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}
	_, err = s.db.Exec(`ALTER TABLE messages ADD COLUMN citations_json TEXT`)
	return err
}
func (s *Store) migrate() error {
	schema := `
CREATE TABLE IF NOT EXISTS api_cache (
  cache_key   TEXT PRIMARY KEY,
  payload     TEXT NOT NULL,
  fetched_at  INTEGER NOT NULL,
  expires_at  INTEGER NOT NULL
);
CREATE TABLE IF NOT EXISTS conversations (
  id            TEXT PRIMARY KEY,
  mode          TEXT NOT NULL,
  source_title  TEXT,
  hot_url       TEXT,
  character_ids TEXT,
  provider      TEXT,
  created_at    INTEGER NOT NULL,
  updated_at    INTEGER NOT NULL
);
CREATE TABLE IF NOT EXISTS messages (
  id              INTEGER PRIMARY KEY AUTOINCREMENT,
  conversation_id TEXT NOT NULL,
  role            TEXT NOT NULL,
  character_id    TEXT,
  character_name  TEXT,
  era             TEXT,
  round_num       INTEGER DEFAULT 1,
  content         TEXT NOT NULL,
  provider        TEXT,
  model           TEXT,
  created_at      INTEGER NOT NULL,
  FOREIGN KEY (conversation_id) REFERENCES conversations(id)
);
CREATE INDEX IF NOT EXISTS idx_messages_conv ON messages(conversation_id);
CREATE INDEX IF NOT EXISTS idx_conversations_updated ON conversations(updated_at DESC);
CREATE TABLE IF NOT EXISTS app_settings (
  setting_key   TEXT PRIMARY KEY,
  setting_value TEXT NOT NULL,
  updated_at    INTEGER NOT NULL
);
`
	_, err := s.db.Exec(schema)
	return err
}

type legacyFile struct {
	Cache         map[string]legacyCacheEntry `json:"cache"`
	Conversations map[string]Conversation     `json:"conversations"`
	Messages      map[string][]Message        `json:"messages"`
	NextMsgID     int64                       `json:"nextMsgId"`
}

type legacyCacheEntry struct {
	Payload   string `json:"payload"`
	FetchedAt int64  `json:"fetchedAt"`
	ExpiresAt int64  `json:"expiresAt"`
}

func (s *Store) importLegacyJSON(dataDir string) error {
	var n int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM api_cache`).Scan(&n); err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	raw, err := os.ReadFile(filepath.Join(dataDir, "renwen.json"))
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	var f legacyFile
	if err := json.Unmarshal(raw, &f); err != nil {
		return nil
	}
	for k, e := range f.Cache {
		_, _ = s.db.Exec(
			`INSERT OR REPLACE INTO api_cache (cache_key, payload, fetched_at, expires_at) VALUES (?, ?, ?, ?)`,
			k, e.Payload, e.FetchedAt, e.ExpiresAt,
		)
	}
	for _, c := range f.Conversations {
		ids, _ := json.Marshal(c.CharacterIDs)
		_, _ = s.db.Exec(
			`INSERT OR IGNORE INTO conversations (id, mode, source_title, hot_url, character_ids, provider, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			c.ID, c.Mode, c.SourceTitle, c.HotURL, string(ids), c.Provider, c.CreatedAt, c.UpdatedAt,
		)
	}
	for convID, msgs := range f.Messages {
		for _, m := range msgs {
			_, _ = s.db.Exec(
				`INSERT OR IGNORE INTO messages (id, conversation_id, role, character_id, character_name, era, round_num, content, provider, model, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
				m.ID, convID, m.Role, m.CharacterID, m.CharacterName, m.Era, m.Round, m.Content, m.Provider, m.Model, m.CreatedAt,
			)
		}
	}
	return nil
}

func (s *Store) GetAPICache(key string) (CacheEntry, bool) {
	entry, ok := s.getAPICacheRow(key)
	if !ok {
		return CacheEntry{}, false
	}
	if time.Now().After(entry.ExpiresAt) {
		return CacheEntry{}, false
	}
	return entry, true
}

func (s *Store) GetAPICacheAny(key string) (CacheEntry, bool) {
	return s.getAPICacheRow(key)
}

func (s *Store) getAPICacheRow(key string) (CacheEntry, bool) {
	var payload string
	var fetchedAt, expiresAt int64
	err := s.db.QueryRow(
		`SELECT payload, fetched_at, expires_at FROM api_cache WHERE cache_key = ?`,
		key,
	).Scan(&payload, &fetchedAt, &expiresAt)
	if err == sql.ErrNoRows {
		return CacheEntry{}, false
	}
	if err != nil {
		return CacheEntry{}, false
	}
	return CacheEntry{
		Key:       key,
		Payload:   payload,
		FetchedAt: time.Unix(fetchedAt, 0),
		ExpiresAt: time.Unix(expiresAt, 0),
	}, true
}

func (s *Store) SetAPICache(key, payload string, ttl time.Duration) error {
	now := time.Now()
	_, err := s.db.Exec(
		`INSERT INTO api_cache (cache_key, payload, fetched_at, expires_at) VALUES (?, ?, ?, ?)
		 ON CONFLICT(cache_key) DO UPDATE SET payload=excluded.payload, fetched_at=excluded.fetched_at, expires_at=excluded.expires_at`,
		key, payload, now.Unix(), now.Add(ttl).Unix(),
	)
	return err
}

func (s *Store) EnsureConversation(id, mode, sourceTitle, hotURL, provider string, characterIDs []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().Unix()
	idsJSON, _ := json.Marshal(characterIDs)
	var exists int
	_ = s.db.QueryRow(`SELECT 1 FROM conversations WHERE id = ?`, id).Scan(&exists)
	if exists == 1 {
		_, err := s.db.Exec(
			`UPDATE conversations SET updated_at = ?, provider = COALESCE(NULLIF(?, ''), provider) WHERE id = ?`,
			now, provider, id,
		)
		return err
	}
	_, err := s.db.Exec(
		`INSERT INTO conversations (id, mode, source_title, hot_url, character_ids, provider, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		id, mode, sourceTitle, hotURL, string(idsJSON), provider, now, now,
	)
	return err
}

func (s *Store) TouchConversation(id string) error {
	_, err := s.db.Exec(`UPDATE conversations SET updated_at = ? WHERE id = ?`, time.Now().Unix(), id)
	return err
}

func (s *Store) AddMessage(msg Message) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if msg.CreatedAt == 0 {
		msg.CreatedAt = time.Now().Unix()
	}
	res, err := s.db.Exec(
		`INSERT INTO messages (conversation_id, role, character_id, character_name, era, round_num, content, provider, model, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		msg.ConversationID, msg.Role, msg.CharacterID, msg.CharacterName, msg.Era, msg.Round, msg.Content, msg.Provider, msg.Model, msg.CreatedAt,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *Store) ListConversations(limit int) ([]Conversation, error) {
	q := `SELECT id, mode, source_title, hot_url, character_ids, provider, created_at, updated_at FROM conversations ORDER BY updated_at DESC`
	if limit > 0 {
		q += fmt.Sprintf(" LIMIT %d", limit)
	}
	rows, err := s.db.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Conversation
	for rows.Next() {
		c, err := scanConversation(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	if out == nil {
		out = []Conversation{}
	}
	return out, rows.Err()
}

func (s *Store) GetConversation(id string) (*Conversation, []Message, error) {
	row := s.db.QueryRow(
		`SELECT id, mode, source_title, hot_url, character_ids, provider, created_at, updated_at FROM conversations WHERE id = ?`,
		id,
	)
	c, err := scanConversationRow(row)
	if err == sql.ErrNoRows {
		return nil, nil, fmt.Errorf("not found")
	}
	if err != nil {
		return nil, nil, err
	}
	rows, err := s.db.Query(
		`SELECT id, conversation_id, role, character_id, character_name, era, round_num, content, provider, model, citations_json, created_at FROM messages WHERE conversation_id = ? ORDER BY id ASC`,
		id,
	)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	var msgs []Message
	for rows.Next() {
		var m Message
		if err := rows.Scan(&m.ID, &m.ConversationID, &m.Role, &m.CharacterID, &m.CharacterName, &m.Era, &m.Round, &m.Content, &m.Provider, &m.Model, &m.CreatedAt); err != nil {
			return nil, nil, err
		}
		msgs = append(msgs, m)
	}
	if msgs == nil {
		msgs = []Message{}
	}
	return &c, msgs, rows.Err()
}

type scannable interface {
	Scan(dest ...any) error
}

func scanConversation(rows *sql.Rows) (Conversation, error) {
	var c Conversation
	var idsRaw sql.NullString
	var sourceTitle, hotURL, provider sql.NullString
	if err := rows.Scan(&c.ID, &c.Mode, &sourceTitle, &hotURL, &idsRaw, &provider, &c.CreatedAt, &c.UpdatedAt); err != nil {
		return c, err
	}
	c.SourceTitle = sourceTitle.String
	c.HotURL = hotURL.String
	c.Provider = provider.String
	if idsRaw.Valid && idsRaw.String != "" {
		_ = json.Unmarshal([]byte(idsRaw.String), &c.CharacterIDs)
	}
	return c, nil
}

func scanConversationRow(row *sql.Row) (Conversation, error) {
	var c Conversation
	var idsRaw sql.NullString
	var sourceTitle, hotURL, provider sql.NullString
	if err := row.Scan(&c.ID, &c.Mode, &sourceTitle, &hotURL, &idsRaw, &provider, &c.CreatedAt, &c.UpdatedAt); err != nil {
		return c, err
	}
	c.SourceTitle = sourceTitle.String
	c.HotURL = hotURL.String
	c.Provider = provider.String
	if idsRaw.Valid && idsRaw.String != "" {
		_ = json.Unmarshal([]byte(idsRaw.String), &c.CharacterIDs)
	}
	return c, nil
}

// ParseCachePayload unmarshals cached API payload into target.
func ParseCachePayload(entry CacheEntry, target any) error {
	return json.Unmarshal([]byte(strings.TrimSpace(entry.Payload)), target)
}
