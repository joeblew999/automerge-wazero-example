module github.com/joeblew999/automerge-wazero-example/guigui-demo

go 1.25.3

// Local dependency on parent module (for pkg/automerge and pkg/wazero)
require github.com/joeblew999/automerge-wazero-example v0.0.0

// Guigui UI framework (will be added when implementing)
// require github.com/guigui-gui/guigui v0.0.0-latest

// Use local parent module (go/ directory)
replace github.com/joeblew999/automerge-wazero-example => ../..
