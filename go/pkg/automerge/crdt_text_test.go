package automerge_test

import (
	"context"
	"testing"

	"github.com/joeblew999/automerge-wazero-example/pkg/automerge"
)

// TestText_SpliceOperations tests all text splice operations
func TestText_SpliceOperations(t *testing.T) {
	tests := []struct {
		name   string
		start  string // Initial text state
		pos    uint
		del    int
		insert string
		want   string
	}{
		{
			name:   "insert at empty",
			start:  "",
			pos:    0,
			del:    0,
			insert: "Hello",
			want:   "Hello",
		},
		{
			name:   "append to end",
			start:  "Hello",
			pos:    5,
			del:    0,
			insert: ", World!",
			want:   "Hello, World!",
		},
		{
			name:   "insert in middle",
			start:  "Helo",
			pos:    2,
			del:    0,
			insert: "l",
			want:   "Hello",
		},
		{
			name:   "delete from middle",
			start:  "Hello, World!",
			pos:    7,
			del:    5,
			insert: "",
			want:   "Hello, !",
		},
		{
			name:   "replace text",
			start:  "Hello, World!",
			pos:    7,
			del:    5,
			insert: "Earth",
			want:   "Hello, Earth!",
		},
		{
			name:   "delete all",
			start:  "Hello",
			pos:    0,
			del:    5,
			insert: "",
			want:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			doc, err := automerge.NewWithWASM(ctx, automerge.TestWASMPath)
			if err != nil {
				t.Fatalf("New() failed: %v", err)
			}
			defer doc.Close(ctx)

			path := automerge.Root().Get("content")

			// Set initial state
			if tt.start != "" {
				if err := doc.SpliceText(ctx, path, 0, 0, tt.start); err != nil {
					t.Fatalf("Initial SpliceText() failed: %v", err)
				}
			}

			// Perform operation
			if err := doc.SpliceText(ctx, path, tt.pos, tt.del, tt.insert); err != nil {
				t.Fatalf("SpliceText() failed: %v", err)
			}

			// Verify result
			got, err := doc.GetText(ctx, path)
			if err != nil {
				t.Fatalf("GetText() failed: %v", err)
			}

			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

// TestText_Unicode tests Unicode and emoji support
func TestText_Unicode(t *testing.T) {
	tests := []struct {
		name string
		text string
	}{
		{
			name: "japanese kanji",
			text: "こんにちは世界",
		},
		{
			name: "chinese characters",
			text: "你好世界",
		},
		{
			name: "emoji only",
			text: "🌍🚀✅❌🎉",
		},
		{
			name: "mixed ascii and unicode",
			text: "Hello 世界! 🌍🚀",
		},
		{
			name: "emoji with skin tones",
			text: "👋🏻👋🏿👨‍👩‍👧‍👦",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			doc, err := automerge.NewWithWASM(ctx, automerge.TestWASMPath)
			if err != nil {
				t.Fatalf("New() failed: %v", err)
			}
			defer doc.Close(ctx)

			path := automerge.Root().Get("content")

			// Insert Unicode text
			if err := doc.SpliceText(ctx, path, 0, 0, tt.text); err != nil {
				t.Fatalf("SpliceText() failed: %v", err)
			}

			// Verify it's preserved
			got, err := doc.GetText(ctx, path)
			if err != nil {
				t.Fatalf("GetText() failed: %v", err)
			}

			if got != tt.text {
				t.Errorf("got %q, want %q", got, tt.text)
			}
		})
	}
}

// TestText_Length tests text length calculations
func TestText_Length(t *testing.T) {
	tests := []struct {
		name       string
		text       string
		wantLength uint32
	}{
		{
			name:       "empty",
			text:       "",
			wantLength: 0,
		},
		{
			name:       "ascii",
			text:       "Hello",
			wantLength: 5,
		},
		{
			name:       "unicode",
			text:       "Hello 世界!",
			wantLength: 13, // UTF-8 byte length (世=3 bytes, 界=3 bytes)
		},
		{
			name:       "emoji",
			text:       "🌍",
			wantLength: 4, // UTF-8 bytes for emoji
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			doc, err := automerge.NewWithWASM(ctx, automerge.TestWASMPath)
			if err != nil {
				t.Fatalf("New() failed: %v", err)
			}
			defer doc.Close(ctx)

			path := automerge.Root().Get("content")

			if tt.text != "" {
				if err := doc.SpliceText(ctx, path, 0, 0, tt.text); err != nil {
					t.Fatalf("SpliceText() failed: %v", err)
				}
			}

			got, err := doc.TextLength(ctx, path)
			if err != nil {
				t.Fatalf("TextLength() failed: %v", err)
			}

			if got != tt.wantLength {
				t.Errorf("got %d, want %d", got, tt.wantLength)
			}
		})
	}
}
