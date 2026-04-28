package pamauth

import (
	"strings"

	pam "github.com/msteinert/pam/v2"
)

type authenticator struct {
	serviceName string
}

// NewAuthenticator constructs a PAM-backed authenticator bound to a PAM
// service stack name such as `litenas-auth`.
func NewAuthenticator(serviceName string) (Authenticator, error) {
	if err := validateServiceName(serviceName); err != nil {
		return nil, err
	}

	return authenticator{
		serviceName: serviceName,
	}, nil
}

// Authenticate performs credential verification followed by account-management
// checks against the configured PAM service.
func (a authenticator) Authenticate(request AuthenticateRequest) (Result, error) {
	conversation := newAuthenticateConversation(request)
	transaction, err := pam.StartFunc(a.serviceName, request.Username, conversation.respond)
	if err != nil {
		return newServiceUnavailableResult(request.Username, err.Error()), err
	}
	defer func() {
		_ = transaction.End()
	}()

	if err := transaction.Authenticate(pam.DisallowNullAuthtok); err != nil {
		return conversation.resultForAuthError(request.Username, err), nil
	}

	if err := transaction.AcctMgmt(0); err != nil {
		return conversation.resultForAccountError(request.Username, err), nil
	}

	return Result{
		Code:     OutcomeAuthenticated,
		Username: request.Username,
		Messages: conversation.messages,
	}, nil
}

// ChangePassword executes a PAM-backed password update flow and surfaces the
// resulting structured outcome.
func (a authenticator) ChangePassword(request PasswordChangeRequest) (Result, error) {
	conversation := newPasswordChangeConversation(request)
	transaction, err := pam.StartFunc(a.serviceName, request.Username, conversation.respond)
	if err != nil {
		return newServiceUnavailableResult(request.Username, err.Error()), err
	}
	defer func() {
		_ = transaction.End()
	}()

	if err := transaction.ChangeAuthTok(0); err != nil {
		return Result{
			Code:     OutcomeDenied,
			Username: request.Username,
			Messages: buildMessages(conversation.messages, Message{Level: MessageLevelError, Text: err.Error()}),
		}, nil
	}

	return Result{
		Code:     OutcomeAuthenticated,
		Username: request.Username,
		Messages: conversation.messages,
	}, nil
}

type authenticateConversation struct {
	request  AuthenticateRequest
	messages []Message
}

func newAuthenticateConversation(request AuthenticateRequest) *authenticateConversation {
	return &authenticateConversation{request: request}
}

func (c *authenticateConversation) respond(style pam.Style, msg string) (string, error) {
	switch style {
	case pam.PromptEchoOff:
		return c.request.Password, nil
	case pam.PromptEchoOn:
		return c.request.Username, nil
	case pam.ErrorMsg:
		c.messages = append(c.messages, Message{Level: MessageLevelError, Text: msg})
		return "", nil
	case pam.TextInfo:
		c.messages = append(c.messages, Message{Level: MessageLevelInfo, Text: msg})
		return "", nil
	default:
		return "", nil
	}
}

func (c *authenticateConversation) resultForAuthError(username string, err error) Result {
	return Result{
		Code:     OutcomeDenied,
		Username: username,
		Messages: buildMessages(c.messages, Message{Level: MessageLevelError, Text: err.Error()}),
	}
}

func (c *authenticateConversation) resultForAccountError(username string, err error) Result {
	return Result{
		Code:              OutcomePasswordChangeNeeded,
		Username:          username,
		CanChangePassword: true,
		Messages:          buildMessages(c.messages, Message{Level: MessageLevelWarn, Text: err.Error()}),
	}
}

type passwordChangeConversation struct {
	request  PasswordChangeRequest
	messages []Message
}

func newPasswordChangeConversation(request PasswordChangeRequest) *passwordChangeConversation {
	return &passwordChangeConversation{request: request}
}

func (c *passwordChangeConversation) respond(style pam.Style, msg string) (string, error) {
	switch style {
	case pam.PromptEchoOff:
		return c.resolveSecret(msg), nil
	case pam.PromptEchoOn:
		return c.request.Username, nil
	case pam.ErrorMsg:
		c.messages = append(c.messages, Message{Level: MessageLevelError, Text: msg})
		return "", nil
	case pam.TextInfo:
		c.messages = append(c.messages, Message{Level: MessageLevelInfo, Text: msg})
		return "", nil
	default:
		return "", nil
	}
}

func (c *passwordChangeConversation) resolveSecret(msg string) string {
	lower := strings.ToLower(msg)
	if strings.Contains(lower, "new") {
		return c.request.NewPassword
	}

	return c.request.OldPassword
}
