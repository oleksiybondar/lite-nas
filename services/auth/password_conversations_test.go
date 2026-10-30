package main

import (
	"context"
	"testing"
	"time"
)

// TestPasswordConversationRejectsOldSequence prevents replaying a prior answer.
func TestPasswordConversationRejectsOldSequence(t *testing.T) {
	flow := &passwordFlow{id: "pending", expiresAt: time.Now().Add(time.Minute)}
	flow.sequence.Store(2)
	manager := &passwordConversations{flows: map[string]*passwordFlow{"pending": flow}}
	_, _, accepted := manager.step(t.Context(), "pending", 1, "answer")
	if accepted {
		t.Fatal("old answer sequence was accepted")
	}
}

// TestPasswordConversationRejectsExpiredFlow bounds the PAM waiting period.
func TestPasswordConversationRejectsExpiredFlow(t *testing.T) {
	flow := &passwordFlow{id: "pending", expiresAt: time.Now().Add(-time.Second), cancel: func() {}}
	flow.sequence.Store(1)
	manager := &passwordConversations{flows: map[string]*passwordFlow{"pending": flow}, ctx: context.Background()}
	_, _, accepted := manager.step(t.Context(), "pending", 1, "answer")
	if accepted {
		t.Fatal("expired conversation was accepted")
	}
}
