package messaging

import "testing"

func TestEnvelopeFields(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		got  func(Envelope) any
		want any
	}{
		{name: "subject", got: func(envelope Envelope) any { return envelope.Subject }, want: "system.metrics.sample"},
		{name: "reply to", got: func(envelope Envelope) any { return envelope.ReplyTo }, want: "_INBOX.reply"},
		{name: "trace id", got: func(envelope Envelope) any { return envelope.Headers["trace-id"] }, want: "abc123"},
		{name: "payload", got: func(envelope Envelope) any { return string(envelope.Payload) }, want: `{"ok":true}`},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			envelope := envelopeFixture()
			if got := testCase.got(envelope); got != testCase.want {
				t.Fatalf("%s = %#v, want %#v", testCase.name, got, testCase.want)
			}
		})
	}
}

func envelopeFixture() Envelope {
	return Envelope{
		Subject: "system.metrics.sample",
		ReplyTo: "_INBOX.reply",
		Headers: map[string]string{
			"content-type": ContentTypeJSON,
			"trace-id":     "abc123",
		},
		Payload: []byte(`{"ok":true}`),
	}
}
