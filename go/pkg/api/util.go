package api

import "github.com/joeblew999/automerge-wazero-example/pkg/automerge"

// parsePathString converts path string like "ROOT" to automerge.Path
// For simplicity, just support "ROOT" for now
// TODO: Parse dotted paths like "ROOT.users.alice"
func parsePathString(pathStr string) automerge.Path {
	return automerge.Root()
}
