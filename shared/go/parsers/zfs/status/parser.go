package status

import (
	"fmt"
	"strings"

	"github.com/antlr4-go/antlr/v4"

	gen "lite-nas/shared/parsers/generated/zfs/zpoolstatus"
)

// ParseZpoolStatus parses zpool status text into a typed document model.
func ParseZpoolStatus(input string, mode ParseMode) (StatusDocument, []Diagnostic, error) {
	if mode == "" {
		mode = ParseModeStrict
	}

	listener := &diagnosticErrorListener{}
	lexer := gen.NewZpoolStatusLexer(antlr.NewInputStream(input))
	lexer.RemoveErrorListeners()
	lexer.AddErrorListener(listener)

	stream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
	parser := gen.NewZpoolStatusParser(stream)
	parser.RemoveErrorListeners()
	parser.AddErrorListener(listener)

	ctx := parser.Document()
	documentCtx, _ := ctx.(*gen.DocumentContext)
	sourceLines := strings.Split(input, "\n")
	document := mapDocument(documentCtx, sourceLines)
	diagnostics := listener.Diagnostics()

	if mode == ParseModeStrict && hasErrorDiagnostic(diagnostics) {
		return StatusDocument{}, diagnostics, fmt.Errorf("zpool status parsing failed in strict mode")
	}

	return document, diagnostics, nil
}

// diagnosticErrorListener captures ANTLR syntax issues as parser diagnostics.
type diagnosticErrorListener struct {
	*antlr.DefaultErrorListener
	diagnostics []Diagnostic
}

// SyntaxError stores one syntax error from lexer or parser processing.
func (l *diagnosticErrorListener) SyntaxError(
	_ antlr.Recognizer,
	_ interface{},
	line int,
	column int,
	msg string,
	_ antlr.RecognitionException,
) {
	l.diagnostics = append(l.diagnostics, Diagnostic{
		Severity: DiagnosticSeverityError,
		Message:  msg,
		Line:     line,
		Column:   column,
	})
}

// Diagnostics returns captured parser diagnostics.
func (l *diagnosticErrorListener) Diagnostics() []Diagnostic {
	if len(l.diagnostics) == 0 {
		return nil
	}

	out := make([]Diagnostic, len(l.diagnostics))
	copy(out, l.diagnostics)
	return out
}

// mapDocument maps parse tree context to typed status document.
func mapDocument(ctx *gen.DocumentContext, sourceLines []string) StatusDocument {
	if ctx == nil {
		return StatusDocument{}
	}

	poolBlocks := ctx.AllPoolBlock()
	pools := make([]PoolBlock, 0, len(poolBlocks))

	for _, block := range poolBlocks {
		typedBlock, ok := block.(*gen.PoolBlockContext)
		if !ok {
			continue
		}
		pools = append(pools, mapPoolBlock(typedBlock, sourceLines))
	}

	return StatusDocument{Pools: pools}
}

// mapPoolBlock maps one pool parse block into PoolBlock DTO.
func mapPoolBlock(block *gen.PoolBlockContext, sourceLines []string) PoolBlock {
	poolLine, _ := block.PoolLine().(*gen.PoolLineContext)
	errorsLine, _ := block.ErrorsLine().(*gen.ErrorsLineContext)
	configSection := block.ConfigSection()

	metadata := PoolMetadata{}
	if meta := block.MetadataSection(); meta != nil {
		typedMeta, ok := meta.(*gen.MetadataSectionContext)
		if ok {
			metadata = mapMetadata(typedMeta)
		}
	}

	return PoolBlock{
		PoolName:      extractPoolLineText(poolLine),
		Metadata:      metadata,
		Config:        mapConfig(configSection, sourceLines),
		ErrorsSummary: extractErrorsLineText(errorsLine),
	}
}

// mapMetadata maps optional metadata lines to PoolMetadata.
func mapMetadata(meta *gen.MetadataSectionContext) PoolMetadata {
	result := PoolMetadata{}
	for _, line := range meta.AllMetadataLine() {
		typedLine, ok := line.(*gen.MetadataLineContext)
		if !ok {
			continue
		}
		applyMetadataLine(&result, typedLine)
	}

	return result
}

// applyMetadataLine maps one metadata variant into the target metadata object.
func applyMetadataLine(result *PoolMetadata, line *gen.MetadataLineContext) {
	if line.StateLine() != nil {
		result.State = extractStateLineText(line.StateLine())
		return
	}
	if line.ScanLine() != nil {
		result.Scan = extractScanLineText(line.ScanLine())
		return
	}
	if line.StatusLine() != nil {
		result.Status = extractStatusLineText(line.StatusLine())
		return
	}
	if line.ActionLine() != nil {
		result.Action = extractActionLineText(line.ActionLine())
		return
	}
	if line.SeeLine() != nil {
		result.See = extractSeeLineText(line.SeeLine())
	}
}

// configFlatRow stores one parsed config row before hierarchy assembly.
type configFlatRow struct {
	indent  int
	name    string
	columns map[string]string
}

// mapConfig maps config header and rows into a hierarchical config tree.
func mapConfig(config gen.IConfigSectionContext, sourceLines []string) ConfigTree {
	if config == nil {
		return ConfigTree{}
	}

	typedConfig, ok := config.(*gen.ConfigSectionContext)
	if !ok {
		return ConfigTree{}
	}

	header := mapConfigHeader(typedConfig.ConfigHeaderLine())
	rows := typedConfig.AllConfigRowLine()

	flatRows := make([]configFlatRow, 0, len(rows))
	for _, row := range rows {
		typedRow, rowOK := row.(*gen.ConfigRowLineContext)
		if !rowOK {
			continue
		}
		flatRows = append(flatRows, mapConfigRow(header, typedRow, sourceLines))
	}

	roots := buildConfigTree(flatRows)
	return ConfigTree{
		Header: header,
		Roots:  roots,
	}
}

// mapConfigHeader extracts ordered header columns from header line.
func mapConfigHeader(headerCtx gen.IConfigHeaderLineContext) []string {
	if headerCtx == nil {
		return nil
	}

	typedHeader, ok := headerCtx.(*gen.ConfigHeaderLineContext)
	if !ok {
		return nil
	}
	atoms := typedHeader.AllHeaderAtom()
	header := make([]string, 0, len(atoms))
	for _, atom := range atoms {
		header = append(header, strings.TrimSpace(atom.GetText()))
	}
	return header
}

// mapConfigRow maps one config row into a flat row representation.
func mapConfigRow(header []string, rowCtx *gen.ConfigRowLineContext, sourceLines []string) configFlatRow {
	atoms := rowCtx.AllRowAtom()
	values := make([]string, 0, len(atoms))
	for _, atom := range atoms {
		values = append(values, strings.TrimSpace(atom.GetText()))
	}

	name := ""
	if len(values) > 0 {
		name = values[0]
	}

	columns := make(map[string]string, len(header))
	limit := minInt(len(values), len(header))
	for index := 0; index < limit; index++ {
		columns[header[index]] = values[index]
	}

	if len(values) > len(header) {
		for index := len(header); index < len(values); index++ {
			extraKey := fmt.Sprintf("_extra_%d", index-len(header)+1)
			columns[extraKey] = values[index]
		}
	}

	return configFlatRow{
		indent:  lineIndentWidth(rowCtx.GetStart().GetLine(), sourceLines),
		name:    name,
		columns: columns,
	}
}

// lineIndentWidth returns indentation width for a 1-based source line index.
func lineIndentWidth(line int, sourceLines []string) int {
	index := line - 1
	if isOutOfBounds(index, len(sourceLines)) {
		return 0
	}

	return leadingIndentWidth(sourceLines[index])
}

// isOutOfBounds reports whether index is outside [0,length).
func isOutOfBounds(index int, length int) bool {
	return index < 0 || index >= length
}

// leadingIndentWidth measures indentation in spaces with tab width equal to 4.
func leadingIndentWidth(line string) int {
	width := 0
	for _, char := range line {
		charWidth, isIndentChar := indentCharWidth(char)
		if !isIndentChar {
			break
		}
		width += charWidth
	}
	return width
}

// indentCharWidth returns indentation width for supported indent characters.
func indentCharWidth(char rune) (int, bool) {
	if char == '\t' {
		return 4, true
	}
	if char == ' ' {
		return 1, true
	}
	return 0, false
}

// buildConfigTree assembles hierarchical nodes from flat config rows.
func buildConfigTree(rows []configFlatRow) []ConfigNode {
	roots := make([]ConfigNode, 0)
	stack := make([]*ConfigNode, 0)

	for _, row := range rows {
		node := ConfigNode{
			Name:    row.name,
			Columns: row.columns,
			Indent:  row.indent,
		}

		for len(stack) > 0 && stack[len(stack)-1].Indent >= row.indent {
			stack = stack[:len(stack)-1]
		}

		if len(stack) == 0 {
			roots = append(roots, node)
			stack = append(stack, &roots[len(roots)-1])
			continue
		}

		parent := stack[len(stack)-1]
		parent.Children = append(parent.Children, node)
		stack = append(stack, &parent.Children[len(parent.Children)-1])
	}

	return roots
}

// extractTextLine returns trimmed text captured by a text line node.
func extractTextLine(textLine gen.ITextLineContext) string {
	if textLine == nil {
		return ""
	}

	typedTextLine, ok := textLine.(*gen.TextLineContext)
	if !ok {
		return ""
	}

	textAtoms := typedTextLine.AllTextAtom()
	if len(textAtoms) == 0 {
		return ""
	}

	parts := make([]string, 0, len(textAtoms))
	for _, textAtom := range textAtoms {
		parts = append(parts, strings.TrimSpace(textAtom.GetText()))
	}
	return strings.TrimSpace(strings.Join(parts, " "))
}

// extractPoolLineText returns the text payload from a pool line.
func extractPoolLineText(poolLine *gen.PoolLineContext) string {
	if poolLine == nil {
		return ""
	}
	return extractTextLine(poolLine.TextLine())
}

// extractErrorsLineText returns the text payload from an errors line.
func extractErrorsLineText(errorsLine *gen.ErrorsLineContext) string {
	if errorsLine == nil {
		return ""
	}
	return extractTextLine(errorsLine.TextLine())
}

// extractStateLineText returns the text payload from a state line.
func extractStateLineText(stateLine gen.IStateLineContext) string {
	if stateLine == nil {
		return ""
	}
	typedStateLine, ok := stateLine.(*gen.StateLineContext)
	if !ok {
		return ""
	}
	return extractTextLine(typedStateLine.TextLine())
}

// extractScanLineText returns the text payload from a scan line.
func extractScanLineText(scanLine gen.IScanLineContext) string {
	if scanLine == nil {
		return ""
	}
	typedScanLine, ok := scanLine.(*gen.ScanLineContext)
	if !ok {
		return ""
	}
	return extractTextLine(typedScanLine.TextLine())
}

// extractStatusLineText returns the text payload from a status line.
func extractStatusLineText(statusLine gen.IStatusLineContext) string {
	if statusLine == nil {
		return ""
	}
	typedStatusLine, ok := statusLine.(*gen.StatusLineContext)
	if !ok {
		return ""
	}
	return extractTextLine(typedStatusLine.TextLine())
}

// extractActionLineText returns the text payload from an action line.
func extractActionLineText(actionLine gen.IActionLineContext) string {
	if actionLine == nil {
		return ""
	}
	typedActionLine, ok := actionLine.(*gen.ActionLineContext)
	if !ok {
		return ""
	}
	return extractTextLine(typedActionLine.TextLine())
}

// extractSeeLineText returns the text payload from a see line.
func extractSeeLineText(seeLine gen.ISeeLineContext) string {
	if seeLine == nil {
		return ""
	}
	typedSeeLine, ok := seeLine.(*gen.SeeLineContext)
	if !ok {
		return ""
	}
	return extractTextLine(typedSeeLine.TextLine())
}

// hasErrorDiagnostic returns true when diagnostics contain an error severity.
func hasErrorDiagnostic(diagnostics []Diagnostic) bool {
	for _, diagnostic := range diagnostics {
		if diagnostic.Severity == DiagnosticSeverityError {
			return true
		}
	}
	return false
}

// minInt returns the smaller integer value.
func minInt(left int, right int) int {
	if left < right {
		return left
	}
	return right
}
