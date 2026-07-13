package storage

import (
	"database/sql"
	"fmt"
	"time"
)

type ConversationSummary struct {
	ConversationID string `json:"conversationId"`
	Content        string `json:"content"`
	CreatedAt      int64  `json:"createdAt"`
}

func (s *Store) migrateSummaries() error {
	_, err := s.db.Exec(`
CREATE TABLE IF NOT EXISTS conversation_summaries (
  conversation_id TEXT PRIMARY KEY,
  content         TEXT NOT NULL,
  created_at      INTEGER NOT NULL,
  FOREIGN KEY (conversation_id) REFERENCES conversations(id)
);
`)
	return err
}

func (s *Store) SaveConversationSummary(conversationID, content string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().Unix()
	_, err := s.db.Exec(
		`INSERT INTO conversation_summaries (conversation_id, content, created_at) VALUES (?, ?, ?)
		 ON CONFLICT(conversation_id) DO UPDATE SET content=excluded.content, created_at=excluded.created_at`,
		conversationID, content, now,
	)
	return err
}

func (s *Store) GetConversationSummary(conversationID string) (*ConversationSummary, error) {
	var sum ConversationSummary
	var createdAt int64
	err := s.db.QueryRow(
		`SELECT conversation_id, content, created_at FROM conversation_summaries WHERE conversation_id = ?`,
		conversationID,
	).Scan(&sum.ConversationID, &sum.Content, &createdAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	sum.CreatedAt = createdAt
	return &sum, nil
}

func (s *Store) ensureSummariesTable() error {
	return s.migrateSummaries()
}

// Open hook: call migrateSummaries after main migrate
func initSummaryMigration(s *Store) error {
	if err := s.migrateSummaries(); err != nil {
		return fmt.Errorf("migrate summaries: %w", err)
	}
	return nil
}
