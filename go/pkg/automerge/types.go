package automerge

import "fmt"

// ObjType represents the type of an Automerge object
type ObjType string

const (
	ObjTypeMap  ObjType = "map"
	ObjTypeList ObjType = "list"
	ObjTypeText ObjType = "text"
)

// ScalarValue represents primitive values in Automerge
type ScalarValue interface {
	isScalarValue()
}

// Scalar value types
type (
	String    string
	Int       int64
	Uint      uint64
	Float     float64
	Boolean   bool
	Null      struct{}
	Counter   int64
	Timestamp int64
	Bytes     []byte
)

func (String) isScalarValue()    {}
func (Int) isScalarValue()       {}
func (Uint) isScalarValue()      {}
func (Float) isScalarValue()     {}
func (Boolean) isScalarValue()   {}
func (Null) isScalarValue()      {}
func (Counter) isScalarValue()   {}
func (Timestamp) isScalarValue() {}
func (Bytes) isScalarValue()     {}

// Value represents any value in an Automerge document
type Value struct {
	scalar ScalarValue
	objID  *ObjID
}

// NewString creates a string value
func NewString(s string) Value {
	return Value{scalar: String(s)}
}

// NewInt creates an integer value
func NewInt(i int64) Value {
	return Value{scalar: Int(i)}
}

// NewBool creates a boolean value
func NewBool(b bool) Value {
	return Value{scalar: Boolean(b)}
}

// NewFloat creates a float value
func NewFloat(f float64) Value {
	return Value{scalar: Float(f)}
}

// NewNull creates a null value
func NewNull() Value {
	return Value{scalar: Null{}}
}

// IsScalar returns true if this is a scalar value
func (v Value) IsScalar() bool {
	return v.scalar != nil
}

// IsObject returns true if this is an object reference
func (v Value) IsObject() bool {
	return v.objID != nil
}

// AsString returns the value as a string if possible
func (v Value) AsString() (string, bool) {
	if s, ok := v.scalar.(String); ok {
		return string(s), true
	}
	return "", false
}

// AsInt returns the value as an int if possible
func (v Value) AsInt() (int64, bool) {
	if i, ok := v.scalar.(Int); ok {
		return int64(i), true
	}
	return 0, false
}

// AsBool returns the value as a boolean if possible
func (v Value) AsBool() (bool, bool) {
	if b, ok := v.scalar.(Boolean); ok {
		return bool(b), true
	}
	return false, false
}

// AsFloat returns the value as a float if possible
func (v Value) AsFloat() (float64, bool) {
	if f, ok := v.scalar.(Float); ok {
		return float64(f), true
	}
	return 0, false
}

// ObjID represents an object identifier
// Currently opaque - will be expanded when we support multi-object operations
type ObjID struct {
	id string // Internal representation
}

func (o ObjID) String() string {
	return o.id
}

// Path represents a path to an object or value in the document tree
type Path struct {
	segments []segment
}

type segment struct {
	key   string // For maps
	index *uint  // For lists (nil for maps)
}

// Root returns a path to the root object
func Root() Path {
	return Path{segments: nil}
}

// Get appends a map key to the path
func (p Path) Get(key string) Path {
	return Path{
		segments: append(p.segments, segment{key: key}),
	}
}

// Index appends a list index to the path
func (p Path) Index(idx uint) Path {
	return Path{
		segments: append(p.segments, segment{index: &idx}),
	}
}

// IsRoot returns true if this is the root path
func (p Path) IsRoot() bool {
	return len(p.segments) == 0
}

// Len returns the number of segments in the path
func (p Path) Len() int {
	return len(p.segments)
}

// Key returns the last segment as a map key
// Panics if the last segment is not a map key
func (p Path) Key() string {
	if len(p.segments) == 0 {
		panic("path has no segments")
	}
	seg := p.segments[len(p.segments)-1]
	if seg.index != nil {
		panic("last segment is a list index, not a map key")
	}
	return seg.key
}

// String returns a human-readable path representation
func (p Path) String() string {
	if p.IsRoot() {
		return "/"
	}
	result := ""
	for _, seg := range p.segments {
		if seg.index != nil {
			result += fmt.Sprintf("[%d]", *seg.index)
		} else {
			result += "/" + seg.key
		}
	}
	return result
}

// ChangeHash represents a hash of a change
type ChangeHash [32]byte

func (h ChangeHash) String() string {
	return fmt.Sprintf("%x", h[:])
}

// Change represents a single change in the document history
// Currently opaque - will be expanded in M1 when we implement sync protocol
type Change struct {
	hash    ChangeHash
	actorID string
	seq     uint64
	// ... other fields TBD
}

// ActorID identifies an actor making changes
type ActorID string

// Mark represents rich text formatting
// For M4 milestone
type Mark struct {
	Name  string
	Value Value
	Start uint
	End   uint
}

// MarkSet is a collection of marks at a position
type MarkSet []Mark

// SyncState tracks synchronization state between peers
// For M1 milestone
type SyncState struct {
	// Internal state - opaque for now
	data []byte
}

// NewSyncState creates a new sync state
func NewSyncState() *SyncState {
	return &SyncState{}
}
