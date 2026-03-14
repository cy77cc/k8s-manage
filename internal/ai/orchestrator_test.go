package ai

import "testing"

func TestComputeTextDelta(t *testing.T) {
	tests := []struct {
		name      string
		last      string
		current   string
		wantChunk string
		wantNext  string
		wantEmit  bool
	}{
		{
			name:      "first chunk emits full content",
			current:   "你",
			wantChunk: "你",
			wantNext:  "你",
			wantEmit:  true,
		},
		{
			name:      "cumulative content emits suffix only",
			last:      "你",
			current:   "你好",
			wantChunk: "好",
			wantNext:  "你好",
			wantEmit:  true,
		},
		{
			name:     "unchanged content emits nothing",
			last:     "你好",
			current:  "你好",
			wantNext: "你好",
		},
		{
			name:     "non prefix reset emits nothing until stable user text resumes",
			last:     "你好",
			current:  "{\"steps\":[\"check\"]}",
			wantNext: "你好",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotChunk, gotNext, gotEmit := computeTextDelta(tt.last, tt.current)
			if gotChunk != tt.wantChunk {
				t.Fatalf("chunk = %q, want %q", gotChunk, tt.wantChunk)
			}
			if gotNext != tt.wantNext {
				t.Fatalf("next = %q, want %q", gotNext, tt.wantNext)
			}
			if gotEmit != tt.wantEmit {
				t.Fatalf("emit = %v, want %v", gotEmit, tt.wantEmit)
			}
		})
	}
}

func TestLooksLikeInternalJSONPayload(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{input: "{\"steps\":[\"check\"]}", want: true},
		{input: "{\"deployment\":\"nginx\",\"replicas\":3}", want: true},
		{input: "正常回答", want: false},
		{input: " {not-json", want: false},
	}

	for _, tt := range tests {
		if got := looksLikeInternalJSONPayload(tt.input); got != tt.want {
			t.Fatalf("looksLikeInternalJSONPayload(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}
