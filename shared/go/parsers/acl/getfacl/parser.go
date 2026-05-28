package getfacl

import (
	"fmt"
	"strings"

	"github.com/antlr4-go/antlr/v4"

	gen "lite-nas/shared/parsers/generated/acl/getfacl"
)

// Parse parses one getfacl output payload into a typed document.
func Parse(input string) (Document, error) {
	listener := &syntaxErrorListener{}

	lexer := gen.NewGetfaclLexer(antlr.NewInputStream(input))
	lexer.RemoveErrorListeners()
	lexer.AddErrorListener(listener)

	stream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
	parser := gen.NewGetfaclParser(stream)
	parser.RemoveErrorListeners()
	parser.AddErrorListener(listener)

	ctx := parser.Document()
	if listener.firstError != nil {
		return Document{}, listener.firstError
	}

	documentCtx, _ := ctx.(*gen.DocumentContext)
	document, err := mapDocument(documentCtx)
	if err != nil {
		return Document{}, err
	}

	if document.Owner == "" || document.Group == "" {
		return Document{}, fmt.Errorf("getfacl output missing owner/group headers")
	}

	return document, nil
}

type syntaxErrorListener struct {
	*antlr.DefaultErrorListener
	firstError error
}

func (l *syntaxErrorListener) SyntaxError(
	_ antlr.Recognizer,
	_ interface{},
	line int,
	column int,
	msg string,
	_ antlr.RecognitionException,
) {
	if l.firstError != nil {
		return
	}
	l.firstError = fmt.Errorf("getfacl parse error at %d:%d: %s", line, column, msg)
}

func mapDocument(ctx *gen.DocumentContext) (Document, error) {
	document := Document{
		NamedUsers:  make(map[string]Permission),
		NamedGroups: make(map[string]Permission),
	}
	if ctx == nil {
		return document, nil
	}

	for _, line := range ctx.AllLine() {
		if err := mapLine(&document, line); err != nil {
			return Document{}, err
		}
	}

	return document, nil
}

func mapLine(document *Document, line gen.ILineContext) error {
	typedLine, ok := line.(*gen.LineContext)
	if !ok {
		return nil
	}

	if headerCtx, ok := typedLine.HeaderLine().(*gen.HeaderLineContext); ok {
		applyHeaderLine(document, headerCtx)
		return nil
	}

	entryCtx, ok := typedLine.AclEntryLine().(*gen.AclEntryLineContext)
	if !ok {
		return nil
	}

	return applyACLEntry(document, entryCtx)
}

func applyHeaderLine(document *Document, line *gen.HeaderLineContext) {
	key, value, ok := parseHeaderKeyValue(line)
	if !ok {
		return
	}
	applyHeaderField(document, key, value)
}

func parseHeaderKeyValue(line *gen.HeaderLineContext) (string, string, bool) {
	if line == nil {
		return "", "", false
	}

	keyCtx, ok := line.HeaderKey().(*gen.HeaderKeyContext)
	if !ok || keyCtx == nil {
		return "", "", false
	}

	return keyCtx.GetText(), parseValueAtom(line.ValueAtom()), true
}

func parseValueAtom(valueAtom gen.IValueAtomContext) string {
	valueCtx, _ := valueAtom.(*gen.ValueAtomContext)
	if valueCtx == nil {
		return ""
	}
	return strings.TrimSpace(valueCtx.GetText())
}

func applyHeaderField(document *Document, key string, value string) {
	switch key {
	case "file":
		document.FilePath = value
	case "owner":
		document.Owner = value
	case "group":
		document.Group = value
	}
}

func applyACLEntry(document *Document, line *gen.AclEntryLineContext) error {
	if line == nil {
		return nil
	}

	// Directory default ACL entries are intentionally ignored for direct path
	// access checks against the current inode permissions.
	if line.DEFAULT_PREFIX() != nil {
		return nil
	}

	tag, qualifier, permission, ok, err := extractACLEntryFields(line)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}

	applyACLEntryToDocument(document, tag, qualifier, permission)
	return nil
}

func extractACLEntryFields(line *gen.AclEntryLineContext) (string, string, Permission, bool, error) {
	tagCtx, ok := line.Tag().(*gen.TagContext)
	if !ok || tagCtx == nil {
		return "", "", Permission{}, false, nil
	}

	permission, err := parsePermission(strings.TrimSpace(line.PERM().GetText()))
	if err != nil {
		return "", "", Permission{}, false, err
	}

	return strings.TrimSpace(tagCtx.GetText()), parseQualifier(line.Qualifier()), permission, true, nil
}

func parseQualifier(qualifierCtx gen.IQualifierContext) string {
	typedQualifier, _ := qualifierCtx.(*gen.QualifierContext)
	if typedQualifier == nil {
		return ""
	}

	valueCtx, _ := typedQualifier.ValueAtom().(*gen.ValueAtomContext)
	if valueCtx == nil {
		return ""
	}

	return strings.TrimSpace(valueCtx.GetText())
}

func applyACLEntryToDocument(document *Document, tag string, qualifier string, permission Permission) {
	if qualifier != "" {
		applyQualifiedACLEntry(document, tag, qualifier, permission)
		return
	}

	applyUnqualifiedACLEntry(document, tag, permission)
}

func applyQualifiedACLEntry(document *Document, tag string, qualifier string, permission Permission) {
	switch tag {
	case "user":
		document.NamedUsers[qualifier] = permission
	case "group":
		document.NamedGroups[qualifier] = permission
	}
}

func applyUnqualifiedACLEntry(document *Document, tag string, permission Permission) {
	switch tag {
	case "user":
		document.User = permission
	case "group":
		document.GroupObject = permission
	case "other":
		document.Other = permission
	case "mask":
		mask := permission
		document.Mask = &mask
	}
}

func parsePermission(raw string) (Permission, error) {
	if len(raw) != 3 {
		return Permission{}, fmt.Errorf("invalid permission triplet: %q", raw)
	}

	return Permission{
		Read:    raw[0] == 'r',
		Write:   raw[1] == 'w',
		Execute: raw[2] == 'x',
	}, nil
}
