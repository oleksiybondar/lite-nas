package pamauth

import (
	"errors"

	pam "github.com/msteinert/pam/v2"
)

// Prompt is one request for input from the host PAM conversation.
type Prompt struct {
	Text     string
	Secret   bool
	Messages []Message
}

// Ask supplies one PAM prompt and waits for the user's response. Callers must
// cancel the context when the browser conversation expires or is abandoned.
type Ask func(Prompt) (string, error)

// RunConversation authenticates with the current password, evaluates account
// policy, and carries a required or requested password change through PAM.
// The caller owns the worker goroutine for the entire lifetime of this call.
// trace receives stage names only; it must not log PAM answers or token values.
func RunConversation(serviceName string, request AuthenticateRequest, onDemand bool, ask Ask, trace func(string)) (Result, error) {
	messages := make([]Message, 0)
	changing := false
	attempts := 0
	countedFailure := false
	respond := func(style pam.Style, text string) (string, error) {
		switch style {
		case pam.TextInfo, pam.ErrorMsg:
			level := MessageLevelInfo
			if style == pam.ErrorMsg {
				level = MessageLevelError
				if changing && !countedFailure {
					attempts++
					countedFailure = true
				}
			}
			messages = append(messages, Message{Level: level, Text: text})
			if attempts >= 3 {
				return "", pam.ErrMaxtries
			}
			return "", nil
		case pam.PromptEchoOff:
			if !changing {
				if onDemand {
					answer, err := ask(Prompt{Text: text, Secret: true, Messages: messages})
					messages = nil
					return answer, err
				}
				return request.Password, nil
			}
			if attempts >= 3 {
				return "", pam.ErrMaxtries
			}
			answer, err := ask(Prompt{Text: text, Secret: true, Messages: messages})
			messages = nil
			countedFailure = false
			return answer, err
		case pam.PromptEchoOn:
			if !changing && !onDemand {
				return request.Username, nil
			}
			answer, err := ask(Prompt{Text: text, Messages: messages})
			messages = nil
			return answer, err
		default:
			return "", pam.ErrConv
		}
	}
	transaction, err := pam.StartFunc(serviceName, request.Username, respond)
	if err != nil {
		return newServiceUnavailableResult(request.Username, err.Error()), err
	}
	defer endTransaction(transaction)
	if err := transaction.Authenticate(pam.DisallowNullAuthtok); err != nil {
		return Result{Code: OutcomeDenied, Username: request.Username, Messages: buildMessages(messages, Message{Level: MessageLevelError, Text: err.Error()})}, nil
	}
	trace("authenticated")
	accountErr := transaction.AcctMgmt(0)
	if accountErr != nil && !errors.Is(accountErr, pam.ErrNewAuthtokReqd) {
		return Result{Code: OutcomeDenied, Username: request.Username, Messages: buildMessages(messages, Message{Level: MessageLevelError, Text: accountErr.Error()})}, nil
	}
	if accountErr == nil && !onDemand {
		return Result{Code: OutcomeAuthenticated, Username: request.Username, Messages: messages}, nil
	}
	trace("account_checked")
	changing = true
	flags := pam.Flags(0)
	if errors.Is(accountErr, pam.ErrNewAuthtokReqd) {
		flags = pam.ChangeExpiredAuthtok
	}
	trace("password_change_started")
	if err := transaction.ChangeAuthTok(flags); err != nil {
		return Result{Code: OutcomeDenied, Username: request.Username, Messages: buildMessages(messages, Message{Level: MessageLevelError, Text: err.Error()})}, nil
	}
	trace("password_change_completed")
	if err := transaction.AcctMgmt(0); err != nil {
		return Result{Code: OutcomeDenied, Username: request.Username, Messages: buildMessages(messages, Message{Level: MessageLevelError, Text: err.Error()})}, nil
	}
	return Result{Code: OutcomeAuthenticated, Username: request.Username, Messages: messages}, nil
}
