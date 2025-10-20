package automerge_test

// TestCase represents a test scenario with expected outcomes
type TestCase struct {
	Name string
	Text string
	Want string // Expected result after operation
}

// CRDT merge test scenarios
var MergeTestCases = []struct {
	Name      string
	Doc1Text  string
	Doc2Text  string
	WantBoth  bool   // Should contain content from both docs
	WantEither bool  // Can be either doc (non-deterministic but valid)
}{
	{
		Name:     "empty documents",
		Doc1Text: "",
		Doc2Text: "",
		WantBoth: true, // Both empty = still empty
	},
	{
		Name:     "one empty, one with text",
		Doc1Text: "Hello",
		Doc2Text: "",
		WantBoth: true, // Should preserve "Hello"
	},
	{
		Name:     "both with same text",
		Doc1Text: "Hello",
		Doc2Text: "Hello",
		WantBoth: true, // Should preserve "Hello"
	},
	{
		Name:     "different text (Alice and Bob)",
		Doc1Text: "Hello from Alice!",
		Doc2Text: "Hello from Bob!",
		WantBoth: true, // CRDT should preserve both
	},
	{
		Name:     "overlapping text",
		Doc1Text: "Hello World",
		Doc2Text: "Hello Everyone",
		WantBoth: false, // Merge behavior depends on CRDT implementation
		WantEither: true,
	},
}

// Text splice test scenarios
var SpliceTestCases = []struct {
	Name   string
	Start  string // Initial text
	Pos    uint
	Del    int64
	Insert string
	Want   string
}{
	{
		Name:   "insert at beginning of empty",
		Start:  "",
		Pos:    0,
		Del:    0,
		Insert: "Hello",
		Want:   "Hello",
	},
	{
		Name:   "insert at end",
		Start:  "Hello",
		Pos:    5,
		Del:    0,
		Insert: ", World!",
		Want:   "Hello, World!",
	},
	{
		Name:   "insert in middle",
		Start:  "Helo",
		Pos:    2,
		Del:    0,
		Insert: "l",
		Want:   "Hello",
	},
	{
		Name:   "delete from middle",
		Start:  "Hello, World!",
		Pos:    7,
		Del:    5,
		Insert: "",
		Want:   "Hello, !",
	},
	{
		Name:   "replace text",
		Start:  "Hello, World!",
		Pos:    7,
		Del:    5,
		Insert: "Earth",
		Want:   "Hello, Earth!",
	},
	{
		Name:   "delete all",
		Start:  "Hello",
		Pos:    0,
		Del:    5,
		Insert: "",
		Want:   "",
	},
}

// Unicode test scenarios
var UnicodeTestCases = []TestCase{
	{
		Name: "japanese kanji",
		Text: "こんにちは世界",
		Want: "こんにちは世界",
	},
	{
		Name: "chinese characters",
		Text: "你好世界",
		Want: "你好世界",
	},
	{
		Name: "emoji only",
		Text: "🌍🚀✅❌🎉",
		Want: "🌍🚀✅❌🎉",
	},
	{
		Name: "mixed ascii and unicode",
		Text: "Hello 世界! 🌍🚀",
		Want: "Hello 世界! 🌍🚀",
	},
	{
		Name: "emoji with skin tones",
		Text: "👋🏻👋🏿👨‍👩‍👧‍👦",
		Want: "👋🏻👋🏿👨‍👩‍👧‍👦",
	},
}

// Binary format test scenarios
var BinaryFormatTestCases = []struct {
	Name            string
	Text            string
	MinSize         int // Minimum expected snapshot size
	MaxSize         int // Maximum expected snapshot size (0 = no limit)
	MustHaveMagic   bool
}{
	{
		Name:          "empty document",
		Text:          "",
		MinSize:       50,  // Empty doc still has header
		MaxSize:       200,
		MustHaveMagic: true,
	},
	{
		Name:          "short text",
		Text:          "Hi",
		MinSize:       60,
		MaxSize:       300,
		MustHaveMagic: true,
	},
	{
		Name:          "normal text",
		Text:          "Hello, World!",
		MinSize:       80,
		MaxSize:       500,
		MustHaveMagic: true,
	},
	{
		Name:          "long text",
		Text:          "The quick brown fox jumps over the lazy dog. " +
			"Pack my box with five dozen liquor jugs. " +
			"How vexingly quick daft zebras jump!",
		MinSize:       200,
		MaxSize:       2000,
		MustHaveMagic: true,
	},
	{
		Name:          "unicode text",
		Text:          "Hello 世界! 🌍🚀 Emoji: ✅❌🎉",
		MinSize:       100,
		MaxSize:       600,
		MustHaveMagic: true,
	},
}

// Automerge binary format magic bytes (should appear at start of all snapshots)
var AutomergeMagicBytes = []byte{0x85, 0x6f, 0x4a, 0x83}
