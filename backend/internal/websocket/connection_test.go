package websocket

import (
	"sync"
	"testing"
	"time"
)

// ========================================
// Initial State Tests
// ========================================

func TestConnection_InitialState(t *testing.T) {
	hub := NewHub()
	conn := NewConnection(nil, hub)

	if conn.State() != ConnectionStatePending {
		t.Errorf("expected initial state Pending, got %v", conn.State())
	}

	if conn.PlayerID() != "" {
		t.Errorf("expected empty player ID, got %q", conn.PlayerID())
	}

	if conn.LobbyCode() != "" {
		t.Errorf("expected empty lobby code, got %q", conn.LobbyCode())
	}
}

// ========================================
// Sequence Number Tests
// ========================================

func TestConnection_NextSeq(t *testing.T) {
	hub := NewHub()
	conn := NewConnection(nil, hub)

	seq1 := conn.NextSeq()
	seq2 := conn.NextSeq()
	seq3 := conn.NextSeq()

	if seq1 != 1 {
		t.Errorf("expected first seq 1, got %d", seq1)
	}
	if seq2 != 2 {
		t.Errorf("expected second seq 2, got %d", seq2)
	}
	if seq3 != 3 {
		t.Errorf("expected third seq 3, got %d", seq3)
	}
}

func TestConnection_CurrentSeq(t *testing.T) {
	hub := NewHub()
	conn := NewConnection(nil, hub)

	initial := conn.CurrentSeq()
	if initial != 0 {
		t.Errorf("expected initial current seq 0, got %d", initial)
	}

	conn.NextSeq()
	current := conn.CurrentSeq()
	if current != 1 {
		t.Errorf("expected current seq 1, got %d", current)
	}

	// CurrentSeq should not increment
	current2 := conn.CurrentSeq()
	if current2 != 1 {
		t.Errorf("expected current seq to remain 1, got %d", current2)
	}
}

func TestConnection_UpdateLastReceivedSeq(t *testing.T) {
	hub := NewHub()
	conn := NewConnection(nil, hub)

	if conn.LastReceivedSeq() != 0 {
		t.Errorf("expected initial last received seq 0, got %d", conn.LastReceivedSeq())
	}

	conn.UpdateLastReceivedSeq(5)
	if conn.LastReceivedSeq() != 5 {
		t.Errorf("expected last received seq 5, got %d", conn.LastReceivedSeq())
	}

	// Should not go backwards
	conn.UpdateLastReceivedSeq(3)
	if conn.LastReceivedSeq() != 5 {
		t.Errorf("expected last received seq to remain 5, got %d", conn.LastReceivedSeq())
	}

	// Should go forward
	conn.UpdateLastReceivedSeq(10)
	if conn.LastReceivedSeq() != 10 {
		t.Errorf("expected last received seq 10, got %d", conn.LastReceivedSeq())
	}
}

func TestConnection_LastReceivedSeq(t *testing.T) {
	hub := NewHub()
	conn := NewConnection(nil, hub)

	conn.UpdateLastReceivedSeq(42)
	if conn.LastReceivedSeq() != 42 {
		t.Errorf("expected last received seq 42, got %d", conn.LastReceivedSeq())
	}
}

// ========================================
// Heartbeat Tests
// ========================================

func TestConnection_LastHeartbeat(t *testing.T) {
	hub := NewHub()
	conn := NewConnection(nil, hub)

	initial := conn.LastHeartbeat()
	if initial.IsZero() {
		t.Error("expected initial heartbeat to be set")
	}

	// Heartbeat should be recent (within last second)
	if time.Since(initial) > time.Second {
		t.Error("expected initial heartbeat to be recent")
	}
}

func TestConnection_UpdateHeartbeat(t *testing.T) {
	hub := NewHub()
	conn := NewConnection(nil, hub)

	before := conn.LastHeartbeat()
	time.Sleep(10 * time.Millisecond)
	conn.UpdateHeartbeat()
	after := conn.LastHeartbeat()

	if !after.After(before) {
		t.Error("expected heartbeat to be updated")
	}
}

// ========================================
// Authentication Tests
// ========================================

func TestConnection_Authenticate(t *testing.T) {
	hub := NewHub()
	conn := NewConnection(nil, hub)

	err := conn.Authenticate("player-1", "LOBBY1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if conn.State() != ConnectionStateActive {
		t.Errorf("expected state Active, got %v", conn.State())
	}

	if conn.PlayerID() != "player-1" {
		t.Errorf("expected player ID 'player-1', got %q", conn.PlayerID())
	}

	if conn.LobbyCode() != "LOBBY1" {
		t.Errorf("expected lobby code 'LOBBY1', got %q", conn.LobbyCode())
	}

	if conn.GetReconnectToken() == "" {
		t.Error("expected reconnect token to be set")
	}

	if conn.GetSessionExpiry().IsZero() {
		t.Error("expected session expiry to be set")
	}
}

// ========================================
// Reconnect Token Tests
// ========================================

func TestConnection_RefreshReconnectToken(t *testing.T) {
	hub := NewHub()
	conn := NewConnection(nil, hub)

	err := conn.Authenticate("player-1", "LOBBY1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	originalToken := conn.GetReconnectToken()

	newToken, err := conn.RefreshReconnectToken()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if newToken == "" {
		t.Error("expected new token to be non-empty")
	}

	if newToken == originalToken {
		t.Error("expected new token to be different from original")
	}

	if conn.GetReconnectToken() != newToken {
		t.Error("expected GetReconnectToken to return new token")
	}
}

func TestConnection_ValidateReconnectToken(t *testing.T) {
	hub := NewHub()
	conn := NewConnection(nil, hub)

	err := conn.Authenticate("player-1", "LOBBY1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	token := conn.GetReconnectToken()

	// Valid token should pass
	if !conn.ValidateReconnectToken(token) {
		t.Error("expected valid token to pass validation")
	}

	// Invalid token should fail
	if conn.ValidateReconnectToken("invalid-token") {
		t.Error("expected invalid token to fail validation")
	}

	// Empty token should fail
	if conn.ValidateReconnectToken("") {
		t.Error("expected empty token to fail validation")
	}
}

// ========================================
// Close Tests
// ========================================

func TestConnection_Close_Idempotent(t *testing.T) {
	hub := NewHub()
	conn := NewConnection(nil, hub)

	// First close should work
	conn.Close()
	if conn.State() != ConnectionStateClosing {
		t.Errorf("expected state Closing, got %v", conn.State())
	}

	// Second close should not panic
	conn.Close()
	if conn.State() != ConnectionStateClosing {
		t.Errorf("expected state to remain Closing, got %v", conn.State())
	}
}

// ========================================
// ErrSendBufferFull Tests
// ========================================

func TestConnection_ErrSendBufferFull(t *testing.T) {
	hub := NewHub()
	conn := NewConnection(nil, hub)

	// Fill the send buffer
	for i := 0; i < sendBufferSize; i++ {
		err := conn.SendRaw([]byte("test"))
		if err != nil {
			t.Fatalf("unexpected error filling buffer: %v", err)
		}
	}

	// Next send should fail with ErrSendBufferFull
	err := conn.SendRaw([]byte("overflow"))
	if err != ErrSendBufferFull {
		t.Errorf("expected ErrSendBufferFull, got %v", err)
	}
}

func TestConnection_ErrSendBufferFull_ErrorMessage(t *testing.T) {
	err := ErrSendBufferFull

	if err.Error() != "send buffer full" {
		t.Errorf("expected error message 'send buffer full', got %q", err.Error())
	}
}

// ========================================
// Concurrent Access Tests
// ========================================

func TestConnection_ConcurrentSequenceAccess(t *testing.T) {
	hub := NewHub()
	conn := NewConnection(nil, hub)

	var wg sync.WaitGroup
	seqs := make(chan int64, 1000)

	// Spawn many goroutines to increment seq concurrently
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				seq := conn.NextSeq()
				seqs <- seq
			}
		}()
	}

	wg.Wait()
	close(seqs)

	// Collect all sequences
	seen := make(map[int64]bool)
	for seq := range seqs {
		if seen[seq] {
			t.Errorf("duplicate sequence number: %d", seq)
		}
		seen[seq] = true
	}

	// Should have 1000 unique sequences (100 * 10)
	if len(seen) != 1000 {
		t.Errorf("expected 1000 unique sequences, got %d", len(seen))
	}

	// Final seq should be 1000
	if conn.CurrentSeq() != 1000 {
		t.Errorf("expected final seq 1000, got %d", conn.CurrentSeq())
	}
}

func TestConnection_ConcurrentHeartbeatAccess(t *testing.T) {
	hub := NewHub()
	conn := NewConnection(nil, hub)

	var wg sync.WaitGroup

	// Concurrent reads and writes
	for i := 0; i < 100; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			conn.UpdateHeartbeat()
		}()
		go func() {
			defer wg.Done()
			_ = conn.LastHeartbeat()
		}()
	}

	wg.Wait()

	// Verify we can still read heartbeat
	hb := conn.LastHeartbeat()
	if hb.IsZero() {
		t.Error("expected heartbeat to be set")
	}
}

func TestConnection_ConcurrentStateAccess(t *testing.T) {
	hub := NewHub()
	conn := NewConnection(nil, hub)

	var wg sync.WaitGroup

	// Concurrent reads
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = conn.State()
			_ = conn.PlayerID()
			_ = conn.LobbyCode()
		}()
	}

	wg.Wait()
}

// ========================================
// SetState Tests
// ========================================

func TestConnection_SetState(t *testing.T) {
	hub := NewHub()
	conn := NewConnection(nil, hub)

	conn.SetState(ConnectionStateActive)
	if conn.State() != ConnectionStateActive {
		t.Errorf("expected state Active, got %v", conn.State())
	}

	conn.SetState(ConnectionStateClosing)
	if conn.State() != ConnectionStateClosing {
		t.Errorf("expected state Closing, got %v", conn.State())
	}
}
