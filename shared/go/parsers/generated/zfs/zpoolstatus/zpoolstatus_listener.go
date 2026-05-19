// Code generated from /home/oleksiybondar/Documents/development/lite-nas/shared/go/parsers/zfs/status/grammar/ZpoolStatus.g4 by ANTLR 4.13.2. DO NOT EDIT.

package zpoolstatus // ZpoolStatus
import "github.com/antlr4-go/antlr/v4"

// ZpoolStatusListener is a complete listener for a parse tree produced by ZpoolStatusParser.
type ZpoolStatusListener interface {
	antlr.ParseTreeListener

	// EnterDocument is called when entering the document production.
	EnterDocument(c *DocumentContext)

	// EnterLeadingBlankLines is called when entering the leadingBlankLines production.
	EnterLeadingBlankLines(c *LeadingBlankLinesContext)

	// EnterPoolBlock is called when entering the poolBlock production.
	EnterPoolBlock(c *PoolBlockContext)

	// EnterTrailingBlankLines is called when entering the trailingBlankLines production.
	EnterTrailingBlankLines(c *TrailingBlankLinesContext)

	// EnterMetadataSection is called when entering the metadataSection production.
	EnterMetadataSection(c *MetadataSectionContext)

	// EnterMetadataLine is called when entering the metadataLine production.
	EnterMetadataLine(c *MetadataLineContext)

	// EnterPoolLine is called when entering the poolLine production.
	EnterPoolLine(c *PoolLineContext)

	// EnterStateLine is called when entering the stateLine production.
	EnterStateLine(c *StateLineContext)

	// EnterScanLine is called when entering the scanLine production.
	EnterScanLine(c *ScanLineContext)

	// EnterStatusLine is called when entering the statusLine production.
	EnterStatusLine(c *StatusLineContext)

	// EnterActionLine is called when entering the actionLine production.
	EnterActionLine(c *ActionLineContext)

	// EnterSeeLine is called when entering the seeLine production.
	EnterSeeLine(c *SeeLineContext)

	// EnterConfigSection is called when entering the configSection production.
	EnterConfigSection(c *ConfigSectionContext)

	// EnterConfigHeaderLine is called when entering the configHeaderLine production.
	EnterConfigHeaderLine(c *ConfigHeaderLineContext)

	// EnterConfigRowLine is called when entering the configRowLine production.
	EnterConfigRowLine(c *ConfigRowLineContext)

	// EnterErrorsLine is called when entering the errorsLine production.
	EnterErrorsLine(c *ErrorsLineContext)

	// EnterTextLine is called when entering the textLine production.
	EnterTextLine(c *TextLineContext)

	// EnterTextAtom is called when entering the textAtom production.
	EnterTextAtom(c *TextAtomContext)

	// EnterHeaderAtom is called when entering the headerAtom production.
	EnterHeaderAtom(c *HeaderAtomContext)

	// EnterRowAtom is called when entering the rowAtom production.
	EnterRowAtom(c *RowAtomContext)

	// ExitDocument is called when exiting the document production.
	ExitDocument(c *DocumentContext)

	// ExitLeadingBlankLines is called when exiting the leadingBlankLines production.
	ExitLeadingBlankLines(c *LeadingBlankLinesContext)

	// ExitPoolBlock is called when exiting the poolBlock production.
	ExitPoolBlock(c *PoolBlockContext)

	// ExitTrailingBlankLines is called when exiting the trailingBlankLines production.
	ExitTrailingBlankLines(c *TrailingBlankLinesContext)

	// ExitMetadataSection is called when exiting the metadataSection production.
	ExitMetadataSection(c *MetadataSectionContext)

	// ExitMetadataLine is called when exiting the metadataLine production.
	ExitMetadataLine(c *MetadataLineContext)

	// ExitPoolLine is called when exiting the poolLine production.
	ExitPoolLine(c *PoolLineContext)

	// ExitStateLine is called when exiting the stateLine production.
	ExitStateLine(c *StateLineContext)

	// ExitScanLine is called when exiting the scanLine production.
	ExitScanLine(c *ScanLineContext)

	// ExitStatusLine is called when exiting the statusLine production.
	ExitStatusLine(c *StatusLineContext)

	// ExitActionLine is called when exiting the actionLine production.
	ExitActionLine(c *ActionLineContext)

	// ExitSeeLine is called when exiting the seeLine production.
	ExitSeeLine(c *SeeLineContext)

	// ExitConfigSection is called when exiting the configSection production.
	ExitConfigSection(c *ConfigSectionContext)

	// ExitConfigHeaderLine is called when exiting the configHeaderLine production.
	ExitConfigHeaderLine(c *ConfigHeaderLineContext)

	// ExitConfigRowLine is called when exiting the configRowLine production.
	ExitConfigRowLine(c *ConfigRowLineContext)

	// ExitErrorsLine is called when exiting the errorsLine production.
	ExitErrorsLine(c *ErrorsLineContext)

	// ExitTextLine is called when exiting the textLine production.
	ExitTextLine(c *TextLineContext)

	// ExitTextAtom is called when exiting the textAtom production.
	ExitTextAtom(c *TextAtomContext)

	// ExitHeaderAtom is called when exiting the headerAtom production.
	ExitHeaderAtom(c *HeaderAtomContext)

	// ExitRowAtom is called when exiting the rowAtom production.
	ExitRowAtom(c *RowAtomContext)
}
