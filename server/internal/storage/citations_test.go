package storage

import "testing"

func TestMessageCitationsRoundTrip(t *testing.T) {
	store, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })
	if err := store.EnsureConversation("c1", "single", "topic", "", "deepseek", []string{"luxun"}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.AddMessage(Message{ConversationID: "c1", Role: "assistant", Content: "reply", Citations: []Citation{{Title: "title", Source: "source", Excerpt: "excerpt"}}}); err != nil {
		t.Fatal(err)
	}
	_, messages, err := store.GetConversation("c1")
	if err != nil {
		t.Fatal(err)
	}
	if len(messages) != 1 || len(messages[0].Citations) != 1 {
		t.Fatalf("citations not restored: %+v", messages)
	}
}
