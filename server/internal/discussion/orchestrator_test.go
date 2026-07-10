package discussion

import (
	"path/filepath"
	"testing"

	"renwen/server/internal/character"
)

func TestValidateCharacterIDs(t *testing.T) {
	store := mustCharacterStore(t)

	_, err := validateCharacterIDs([]string{"luxun"}, store)
	if err == nil {
		t.Fatal("expected min group members error")
	}

	ids, err := validateCharacterIDs([]string{"luxun", "sushi", "luxun"}, store)
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 2 {
		t.Fatalf("expected 2 unique ids, got %d", len(ids))
	}

	_, err = validateCharacterIDs([]string{"luxun", "sushi", "lihongzhang", "libai", "zhugeliang", "wangyangming"}, store)
	if err == nil {
		t.Fatal("expected max group members error")
	}
}

func mustCharacterStore(t *testing.T) *character.Store {
	t.Helper()
	for _, p := range []string{
		filepath.Join("..", "..", "config", "characters.json"),
		filepath.Join("config", "characters.json"),
		filepath.Join("server", "config", "characters.json"),
	} {
		s, err := character.LoadFromFile(p)
		if err == nil {
			return s
		}
	}
	t.Fatal("characters.json not found")
	return nil
}
