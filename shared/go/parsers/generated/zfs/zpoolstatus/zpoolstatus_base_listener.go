// Code generated from /home/oleksiybondar/Documents/development/lite-nas/shared/go/parsers/zfs/status/grammar/ZpoolStatus.g4 by ANTLR 4.13.2. DO NOT EDIT.

package zpoolstatus // ZpoolStatus
import "github.com/antlr4-go/antlr/v4"

// BaseZpoolStatusListener is a complete listener for a parse tree produced by ZpoolStatusParser.
type BaseZpoolStatusListener struct{}

var _ ZpoolStatusListener = &BaseZpoolStatusListener{}

// VisitTerminal is called when a terminal node is visited.
func (s *BaseZpoolStatusListener) VisitTerminal(node antlr.TerminalNode) {}

// VisitErrorNode is called when an error node is visited.
func (s *BaseZpoolStatusListener) VisitErrorNode(node antlr.ErrorNode) {}

// EnterEveryRule is called when any rule is entered.
func (s *BaseZpoolStatusListener) EnterEveryRule(ctx antlr.ParserRuleContext) {}

// ExitEveryRule is called when any rule is exited.
func (s *BaseZpoolStatusListener) ExitEveryRule(ctx antlr.ParserRuleContext) {}

// EnterDocument is called when production document is entered.
func (s *BaseZpoolStatusListener) EnterDocument(ctx *DocumentContext) {}

// ExitDocument is called when production document is exited.
func (s *BaseZpoolStatusListener) ExitDocument(ctx *DocumentContext) {}

// EnterLeadingBlankLines is called when production leadingBlankLines is entered.
func (s *BaseZpoolStatusListener) EnterLeadingBlankLines(ctx *LeadingBlankLinesContext) {}

// ExitLeadingBlankLines is called when production leadingBlankLines is exited.
func (s *BaseZpoolStatusListener) ExitLeadingBlankLines(ctx *LeadingBlankLinesContext) {}

// EnterPoolBlock is called when production poolBlock is entered.
func (s *BaseZpoolStatusListener) EnterPoolBlock(ctx *PoolBlockContext) {}

// ExitPoolBlock is called when production poolBlock is exited.
func (s *BaseZpoolStatusListener) ExitPoolBlock(ctx *PoolBlockContext) {}

// EnterTrailingBlankLines is called when production trailingBlankLines is entered.
func (s *BaseZpoolStatusListener) EnterTrailingBlankLines(ctx *TrailingBlankLinesContext) {}

// ExitTrailingBlankLines is called when production trailingBlankLines is exited.
func (s *BaseZpoolStatusListener) ExitTrailingBlankLines(ctx *TrailingBlankLinesContext) {}

// EnterMetadataSection is called when production metadataSection is entered.
func (s *BaseZpoolStatusListener) EnterMetadataSection(ctx *MetadataSectionContext) {}

// ExitMetadataSection is called when production metadataSection is exited.
func (s *BaseZpoolStatusListener) ExitMetadataSection(ctx *MetadataSectionContext) {}

// EnterMetadataLine is called when production metadataLine is entered.
func (s *BaseZpoolStatusListener) EnterMetadataLine(ctx *MetadataLineContext) {}

// ExitMetadataLine is called when production metadataLine is exited.
func (s *BaseZpoolStatusListener) ExitMetadataLine(ctx *MetadataLineContext) {}

// EnterPoolLine is called when production poolLine is entered.
func (s *BaseZpoolStatusListener) EnterPoolLine(ctx *PoolLineContext) {}

// ExitPoolLine is called when production poolLine is exited.
func (s *BaseZpoolStatusListener) ExitPoolLine(ctx *PoolLineContext) {}

// EnterStateLine is called when production stateLine is entered.
func (s *BaseZpoolStatusListener) EnterStateLine(ctx *StateLineContext) {}

// ExitStateLine is called when production stateLine is exited.
func (s *BaseZpoolStatusListener) ExitStateLine(ctx *StateLineContext) {}

// EnterScanLine is called when production scanLine is entered.
func (s *BaseZpoolStatusListener) EnterScanLine(ctx *ScanLineContext) {}

// ExitScanLine is called when production scanLine is exited.
func (s *BaseZpoolStatusListener) ExitScanLine(ctx *ScanLineContext) {}

// EnterStatusLine is called when production statusLine is entered.
func (s *BaseZpoolStatusListener) EnterStatusLine(ctx *StatusLineContext) {}

// ExitStatusLine is called when production statusLine is exited.
func (s *BaseZpoolStatusListener) ExitStatusLine(ctx *StatusLineContext) {}

// EnterActionLine is called when production actionLine is entered.
func (s *BaseZpoolStatusListener) EnterActionLine(ctx *ActionLineContext) {}

// ExitActionLine is called when production actionLine is exited.
func (s *BaseZpoolStatusListener) ExitActionLine(ctx *ActionLineContext) {}

// EnterSeeLine is called when production seeLine is entered.
func (s *BaseZpoolStatusListener) EnterSeeLine(ctx *SeeLineContext) {}

// ExitSeeLine is called when production seeLine is exited.
func (s *BaseZpoolStatusListener) ExitSeeLine(ctx *SeeLineContext) {}

// EnterConfigSection is called when production configSection is entered.
func (s *BaseZpoolStatusListener) EnterConfigSection(ctx *ConfigSectionContext) {}

// ExitConfigSection is called when production configSection is exited.
func (s *BaseZpoolStatusListener) ExitConfigSection(ctx *ConfigSectionContext) {}

// EnterConfigHeaderLine is called when production configHeaderLine is entered.
func (s *BaseZpoolStatusListener) EnterConfigHeaderLine(ctx *ConfigHeaderLineContext) {}

// ExitConfigHeaderLine is called when production configHeaderLine is exited.
func (s *BaseZpoolStatusListener) ExitConfigHeaderLine(ctx *ConfigHeaderLineContext) {}

// EnterConfigRowLine is called when production configRowLine is entered.
func (s *BaseZpoolStatusListener) EnterConfigRowLine(ctx *ConfigRowLineContext) {}

// ExitConfigRowLine is called when production configRowLine is exited.
func (s *BaseZpoolStatusListener) ExitConfigRowLine(ctx *ConfigRowLineContext) {}

// EnterErrorsLine is called when production errorsLine is entered.
func (s *BaseZpoolStatusListener) EnterErrorsLine(ctx *ErrorsLineContext) {}

// ExitErrorsLine is called when production errorsLine is exited.
func (s *BaseZpoolStatusListener) ExitErrorsLine(ctx *ErrorsLineContext) {}

// EnterTextLine is called when production textLine is entered.
func (s *BaseZpoolStatusListener) EnterTextLine(ctx *TextLineContext) {}

// ExitTextLine is called when production textLine is exited.
func (s *BaseZpoolStatusListener) ExitTextLine(ctx *TextLineContext) {}

// EnterTextAtom is called when production textAtom is entered.
func (s *BaseZpoolStatusListener) EnterTextAtom(ctx *TextAtomContext) {}

// ExitTextAtom is called when production textAtom is exited.
func (s *BaseZpoolStatusListener) ExitTextAtom(ctx *TextAtomContext) {}

// EnterHeaderAtom is called when production headerAtom is entered.
func (s *BaseZpoolStatusListener) EnterHeaderAtom(ctx *HeaderAtomContext) {}

// ExitHeaderAtom is called when production headerAtom is exited.
func (s *BaseZpoolStatusListener) ExitHeaderAtom(ctx *HeaderAtomContext) {}

// EnterRowAtom is called when production rowAtom is entered.
func (s *BaseZpoolStatusListener) EnterRowAtom(ctx *RowAtomContext) {}

// ExitRowAtom is called when production rowAtom is exited.
func (s *BaseZpoolStatusListener) ExitRowAtom(ctx *RowAtomContext) {}
