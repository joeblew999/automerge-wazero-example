package server

// SSE client management functions

// Broadcast sends a text update to all connected SSE clients
func (s *Server) Broadcast(text string) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, ch := range s.clients {
		select {
		case ch <- text:
		default:
			// Channel full, skip
		}
	}
}

// AddClient registers a new SSE client channel
func (s *Server) AddClient(ch chan string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.clients = append(s.clients, ch)
}

// RemoveClient unregisters an SSE client channel
func (s *Server) RemoveClient(ch chan string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, client := range s.clients {
		if client == ch {
			s.clients = append(s.clients[:i], s.clients[i+1:]...)
			break
		}
	}
}
