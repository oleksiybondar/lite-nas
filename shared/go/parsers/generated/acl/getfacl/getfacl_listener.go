// Code generated from /home/oleksiybondar/Documents/development/lite-nas/shared/go/parsers/acl/getfacl/grammar/Getfacl.g4 by ANTLR 4.13.2. DO NOT EDIT.

package getfacl // Getfacl
import "github.com/antlr4-go/antlr/v4"

// GetfaclListener is a complete listener for a parse tree produced by GetfaclParser.
type GetfaclListener interface {
	antlr.ParseTreeListener

	// EnterDocument is called when entering the document production.
	EnterDocument(c *DocumentContext)

	// EnterLine is called when entering the line production.
	EnterLine(c *LineContext)

	// EnterHeaderLine is called when entering the headerLine production.
	EnterHeaderLine(c *HeaderLineContext)

	// EnterCommentLine is called when entering the commentLine production.
	EnterCommentLine(c *CommentLineContext)

	// EnterAclEntryLine is called when entering the aclEntryLine production.
	EnterAclEntryLine(c *AclEntryLineContext)

	// EnterHeaderKey is called when entering the headerKey production.
	EnterHeaderKey(c *HeaderKeyContext)

	// EnterTag is called when entering the tag production.
	EnterTag(c *TagContext)

	// EnterQualifier is called when entering the qualifier production.
	EnterQualifier(c *QualifierContext)

	// EnterValueAtom is called when entering the valueAtom production.
	EnterValueAtom(c *ValueAtomContext)

	// ExitDocument is called when exiting the document production.
	ExitDocument(c *DocumentContext)

	// ExitLine is called when exiting the line production.
	ExitLine(c *LineContext)

	// ExitHeaderLine is called when exiting the headerLine production.
	ExitHeaderLine(c *HeaderLineContext)

	// ExitCommentLine is called when exiting the commentLine production.
	ExitCommentLine(c *CommentLineContext)

	// ExitAclEntryLine is called when exiting the aclEntryLine production.
	ExitAclEntryLine(c *AclEntryLineContext)

	// ExitHeaderKey is called when exiting the headerKey production.
	ExitHeaderKey(c *HeaderKeyContext)

	// ExitTag is called when exiting the tag production.
	ExitTag(c *TagContext)

	// ExitQualifier is called when exiting the qualifier production.
	ExitQualifier(c *QualifierContext)

	// ExitValueAtom is called when exiting the valueAtom production.
	ExitValueAtom(c *ValueAtomContext)
}
