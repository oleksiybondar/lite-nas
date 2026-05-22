package configtest

// InvalidSharedMessagingLoggingConfigFixture returns a minimal invalid shared
// messaging/logging config payload used by loaders that require full section
// validation.
func InvalidSharedMessagingLoggingConfigFixture() string {
	return "[messaging]\n" +
		"url=nats://127.0.0.1:4222\n" +
		"timeout=5s\n" +
		"[logging]\n" +
		"output=file\n"
}
