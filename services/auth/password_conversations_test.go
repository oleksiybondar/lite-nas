package main

import (
	"context"
	"strconv"
	"testing"
	"time"

	"lite-nas/services/auth/pamauth"
	authcontract "lite-nas/shared/contracts/auth"
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

// TestPasswordConversationDeliversReply checks the one-to-one prompt channel.
func TestPasswordConversationDeliversReply(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), time.Second)
	defer cancel()
	flow := &passwordFlow{id: "pending", expiresAt: time.Now().Add(time.Minute), ctx: ctx, answers: make(chan string), events: make(chan passwordEvent)}
	flow.sequence.Store(1)
	manager := &passwordConversations{flows: map[string]*passwordFlow{"pending": flow}}
	go func() {
		<-flow.answers
		flow.events <- passwordEvent{result: pamauth.Result{Code: pamauth.OutcomeAuthenticated}}
	}()
	event, _, accepted := manager.step(ctx, "pending", 1, "new secret")
	if !accepted || event.result.Code != pamauth.OutcomeAuthenticated {
		t.Fatalf("accepted = %t, event = %+v", accepted, event)
	}
}

// TestPasswordConversationRejectsCapacityLimit bounds concurrent PAM work.
func TestPasswordConversationRejectsCapacityLimit(t *testing.T) {
	manager := newPasswordConversations(t.Context(), "missing-service", nil)
	for index := 0; index < maxPasswordConversations; index++ {
		manager.flows[strconv.Itoa(index)] = &passwordFlow{}
	}
	_, _, accepted := manager.start(t.Context(), authcontract.LoginRequest{Username: "alice"}, false)
	if accepted {
		t.Fatal("conversation exceeded capacity")
	}
}

// TestPasswordConversationStopsOnCancelledRPC prevents a stuck PAM reply.
func TestPasswordConversationStopsOnCancelledRPC(t *testing.T) {
	flowCtx, stopFlow := context.WithCancel(t.Context())
	defer stopFlow()
	flow := &passwordFlow{id: "pending", expiresAt: time.Now().Add(time.Minute), ctx: flowCtx, answers: make(chan string)}
	flow.sequence.Store(1)
	manager := &passwordConversations{flows: map[string]*passwordFlow{"pending": flow}}
	requestCtx, cancelRequest := context.WithCancel(t.Context())
	cancelRequest()
	_, _, accepted := manager.step(requestCtx, "pending", 1, "answer")
	if accepted {
		t.Fatal("cancelled RPC delivered a password answer")
	}
}
