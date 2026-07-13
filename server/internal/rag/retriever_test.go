package rag

import "testing"

func TestAugmentRanksTagsAndDeduplicatesSources(t *testing.T) {
	r := &Retriever{byCharacter: map[string][]Passage{"c": {
		{Title: "正文命中", Text: "这是关于知行合一的材料", Source: "same"},
		{Title: "标签命中", Text: "材料", Tags: []string{"知行合一"}, Source: "same"},
		{Title: "另一来源", Text: "知行合一", Source: "other"},
	}}}
	_, citations := r.Augment("c", "怎样做到知行合一", "", "base")
	if len(citations) != 2 {
		t.Fatalf("got %d citations, want 2", len(citations))
	}
	if citations[0].Title != "标签命中" {
		t.Fatalf("got first %q, want tag match", citations[0].Title)
	}
}

func TestAugmentReturnsNoCitationsWithoutMatch(t *testing.T) {
	r := &Retriever{byCharacter: map[string][]Passage{"c": {{Title: "历史", Text: "旧事", Source: "s"}}}}
	prompt, citations := r.Augment("c", "量子计算", "", "base")
	if prompt != "base" || len(citations) != 0 {
		t.Fatalf("unexpected result: %q %+v", prompt, citations)
	}
}

func TestAugmentSystemPrompt_matchesTags(t *testing.T) {
	r := &Retriever{
		byCharacter: map[string][]Passage{
			"luxun": {
				{Title: "国民性", Text: "看客的材料", Tags: []string{"国民性"}},
				{Title: "无关", Text: "其他话题", Tags: []string{"无关"}},
			},
		},
	}
	out := r.AugmentSystemPrompt("luxun", "国民性如何", "热榜议题", "base")
	if !contains(out, "国民性") || !contains(out, "看客") {
		t.Fatalf("expected matched passage in prompt: %s", out)
	}
}

func contains(s, sub string) bool {
	return len(sub) == 0 || (len(s) >= len(sub) && indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
