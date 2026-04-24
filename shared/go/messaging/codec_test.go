package messaging

import "testing"

func TestJSONCodecDecodedFields(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		got  func(map[string]any) any
		want any
	}{
		{name: "status", got: func(decoded map[string]any) any { return decoded["status"] }, want: "ok"},
		{name: "count", got: func(decoded map[string]any) any { return decoded["count"] }, want: float64(2)},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			decoded := loadDecodedJSONFixture(t)
			if got := testCase.got(decoded); got != testCase.want {
				t.Fatalf("%s = %#v, want %#v", testCase.name, got, testCase.want)
			}
		})
	}
}

func TestJSONCodecContentType(t *testing.T) {
	t.Parallel()

	codec := NewJSONCodec()
	if codec.ContentType() != ContentTypeJSON {
		t.Fatalf("ContentType() = %q, want %q", codec.ContentType(), ContentTypeJSON)
	}
}

func TestJSONCodecRejectsInvalidPayload(t *testing.T) {
	t.Parallel()

	codec := NewJSONCodec()

	var decoded map[string]any
	if err := codec.Unmarshal([]byte("{"), &decoded); err == nil {
		t.Fatal("expected unmarshal error")
	}
}

func loadDecodedJSONFixture(t *testing.T) map[string]any {
	t.Helper()

	codec := NewJSONCodec()

	data, err := codec.Marshal(map[string]any{"status": "ok", "count": 2})
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	var decoded map[string]any
	if err := codec.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	return decoded
}
