// Code generated from /home/oleksiybondar/Documents/development/lite-nas/shared/go/parsers/acl/getfacl/grammar/Getfacl.g4 by ANTLR 4.13.2. DO NOT EDIT.

package getfacl // Getfacl
import "github.com/antlr4-go/antlr/v4"

// BaseGetfaclListener is a complete listener for a parse tree produced by GetfaclParser.
type BaseGetfaclListener struct{}

var _ GetfaclListener = &BaseGetfaclListener{}

// VisitTerminal is called when a terminal node is visited.
func (s *BaseGetfaclListener) VisitTerminal(node antlr.TerminalNode) {}

// VisitErrorNode is called when an error node is visited.
func (s *BaseGetfaclListener) VisitErrorNode(node antlr.ErrorNode) {}

// EnterEveryRule is called when any rule is entered.
func (s *BaseGetfaclListener) EnterEveryRule(ctx antlr.ParserRuleContext) {}

// ExitEveryRule is called when any rule is exited.
func (s *BaseGetfaclListener) ExitEveryRule(ctx antlr.ParserRuleContext) {}

// EnterDocument is called when production document is entered.
func (s *BaseGetfaclListener) EnterDocument(ctx *DocumentContext) {}

// ExitDocument is called when production document is exited.
func (s *BaseGetfaclListener) ExitDocument(ctx *DocumentContext) {}

// EnterLine is called when production line is entered.
func (s *BaseGetfaclListener) EnterLine(ctx *LineContext) {}

// ExitLine is called when production line is exited.
func (s *BaseGetfaclListener) ExitLine(ctx *LineContext) {}

// EnterHeaderLine is called when production headerLine is entered.
func (s *BaseGetfaclListener) EnterHeaderLine(ctx *HeaderLineContext) {}

// ExitHeaderLine is called when production headerLine is exited.
func (s *BaseGetfaclListener) ExitHeaderLine(ctx *HeaderLineContext) {}

// EnterCommentLine is called when production commentLine is entered.
func (s *BaseGetfaclListener) EnterCommentLine(ctx *CommentLineContext) {}

// ExitCommentLine is called when production commentLine is exited.
func (s *BaseGetfaclListener) ExitCommentLine(ctx *CommentLineContext) {}

// EnterAclEntryLine is called when production aclEntryLine is entered.
func (s *BaseGetfaclListener) EnterAclEntryLine(ctx *AclEntryLineContext) {}

// ExitAclEntryLine is called when production aclEntryLine is exited.
func (s *BaseGetfaclListener) ExitAclEntryLine(ctx *AclEntryLineContext) {}

// EnterHeaderKey is called when production headerKey is entered.
func (s *BaseGetfaclListener) EnterHeaderKey(ctx *HeaderKeyContext) {}

// ExitHeaderKey is called when production headerKey is exited.
func (s *BaseGetfaclListener) ExitHeaderKey(ctx *HeaderKeyContext) {}

// EnterTag is called when production tag is entered.
func (s *BaseGetfaclListener) EnterTag(ctx *TagContext) {}

// ExitTag is called when production tag is exited.
func (s *BaseGetfaclListener) ExitTag(ctx *TagContext) {}

// EnterQualifier is called when production qualifier is entered.
func (s *BaseGetfaclListener) EnterQualifier(ctx *QualifierContext) {}

// ExitQualifier is called when production qualifier is exited.
func (s *BaseGetfaclListener) ExitQualifier(ctx *QualifierContext) {}

// EnterValueAtom is called when production valueAtom is entered.
func (s *BaseGetfaclListener) EnterValueAtom(ctx *ValueAtomContext) {}

// ExitValueAtom is called when production valueAtom is exited.
func (s *BaseGetfaclListener) ExitValueAtom(ctx *ValueAtomContext) {}
