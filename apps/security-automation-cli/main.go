package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"

	"lite-nas/shared/cliauth"
	"lite-nas/shared/config"
	contract "lite-nas/shared/contracts/securityautomation"
	"lite-nas/shared/fileio"
	"lite-nas/shared/messaging"
	"lite-nas/shared/modules"
)

const configPath = "/etc/lite-nas/security-automation-cli.conf"

var (
	ruleIDPattern      = regexp.MustCompile(`^[A-Za-z0-9_]{1,64}$`)
	incidentIDPattern  = regexp.MustCompile(`^[A-Za-z0-9]{1,64}$`)
	operandNamePattern = regexp.MustCompile(`^[a-z][a-z0-9_]{0,63}$`)
)

// main runs one authenticated operator command.
func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	if err := run(ctx, os.Args[1:]); err != nil && !errors.Is(err, context.Canceled) {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// run authenticates the current user and sends one validated automation RPC.
func run(ctx context.Context, args []string) error {
	if err := validateCommand(args); err != nil {
		return err
	}
	reader, err := fileio.NewFileReader(configPath)
	if err != nil {
		return err
	}
	iniFile, err := config.LoadINI(reader)
	if err != nil {
		return err
	}
	messagingConfig, err := config.LoadMessagingConfig(iniFile)
	if err != nil {
		return err
	}
	loggingConfig, err := config.LoadLoggingConfig(iniFile)
	if err != nil {
		return err
	}
	infra, err := modules.NewCoreClientInfra("security-automation-cli", loggingConfig, messagingConfig)
	if err != nil {
		return err
	}
	defer infra.Close()
	token, err := cliauth.LoginInteractive(ctx, infra.Client)
	if err != nil {
		return err
	}
	return executeCommand(ctx, infra.Client, token, args)
}

// validateCommand rejects malformed or option-shaped identifiers before login.
func validateCommand(args []string) error {
	usage := errors.New("usage: security-automation-cli rules|active|rollback ID|force-normalize ID|enable RULE|disable RULE|cooldown-reset RULE KEY|force-mitigate RULE name=value... --yes")
	if len(args) == 0 {
		return usage
	}
	switch args[0] {
	case "rules", "active":
		if len(args) != 1 {
			return usage
		}
	case "rollback", "force-normalize":
		if len(args) != 2 || !incidentIDPattern.MatchString(args[1]) {
			return usage
		}
	case "enable", "disable":
		if len(args) != 2 || !ruleIDPattern.MatchString(args[1]) {
			return usage
		}
	case "cooldown-reset":
		if len(args) != 3 || !ruleIDPattern.MatchString(args[1]) || len(args[2]) == 0 || len(args[2]) > 256 || args[2][0] == '-' {
			return usage
		}
		for _, char := range args[2] {
			if char < 33 || char > 126 {
				return usage
			}
		}
	case "force-mitigate":
		if len(args) < 3 || len(args) > 19 || !ruleIDPattern.MatchString(args[1]) || args[len(args)-1] != "--yes" {
			return usage
		}
		seen := make(map[string]bool)
		for _, argument := range args[2 : len(args)-1] {
			name, value, found := strings.Cut(argument, "=")
			if !found || !operandNamePattern.MatchString(name) || seen[name] || !validMitigationOperand(value) {
				return usage
			}
			seen[name] = true
		}
	default:
		return usage
	}
	return nil
}

// executeCommand calls the same authenticated NATS RPCs as the gateway.
func executeCommand(ctx context.Context, client messaging.Client, token string, args []string) error {
	if err := validateCommand(args); err != nil {
		return err
	}
	var result any
	var err error
	switch args[0] {
	case "rules":
		var response contract.RulesListResponse
		err = client.Request(ctx, contract.RulesListSubject, contract.ListRequest{AccessToken: token}, &response)
		result = response
	case "active":
		var response contract.ActiveListResponse
		err = client.Request(ctx, contract.ActiveListSubject, contract.ListRequest{AccessToken: token}, &response)
		result = response
	case "rollback", "force-normalize":
		subject := contract.RollbackSubject
		if args[0] == "force-normalize" {
			subject = contract.ForceNormalizeSubject
		}
		var response contract.OperationResponse
		err = client.Request(ctx, subject, contract.RollbackRequest{AccessToken: token, IncidentID: args[1]}, &response)
		result = response
	case "enable", "disable":
		var response contract.OperationResponse
		err = client.Request(ctx, contract.RuleSetEnabledSubject, contract.RuleSetEnabledRequest{AccessToken: token, RuleID: args[1], Enabled: args[0] == "enable"}, &response)
		result = response
	case "cooldown-reset":
		var response contract.OperationResponse
		err = client.Request(ctx, contract.CooldownResetSubject, contract.CooldownResetRequest{AccessToken: token, RuleID: args[1], CorrelationKey: args[2]}, &response)
		result = response
	case "force-mitigate":
		operands := make(map[string]string, len(args)-3)
		for _, argument := range args[2 : len(args)-1] {
			name, value, _ := strings.Cut(argument, "=")
			operands[name] = value
		}
		var rules contract.RulesListResponse
		if err := client.Request(ctx, contract.RulesListSubject, contract.ListRequest{AccessToken: token}, &rules); err != nil {
			return err
		}
		var selected *contract.Rule
		for index := range rules.Rules {
			if rules.Rules[index].ID == args[1] {
				selected = &rules.Rules[index]
				break
			}
		}
		if selected == nil {
			return fmt.Errorf("unknown automation rule %q", args[1])
		}
		operands, err = completeMitigationOperands(*selected, operands, promptMitigationOperand)
		if err != nil {
			return err
		}
		var response contract.OperationResponse
		err = client.Request(ctx, contract.ForceMitigateSubject, contract.ForceMitigateRequest{AccessToken: token, RuleID: args[1], Operands: operands}, &response)
		result = response
	}
	if err != nil {
		return err
	}
	return json.NewEncoder(os.Stdout).Encode(result)
}
