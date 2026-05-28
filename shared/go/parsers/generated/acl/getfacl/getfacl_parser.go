// Code generated from /home/oleksiybondar/Documents/development/lite-nas/shared/go/parsers/acl/getfacl/grammar/Getfacl.g4 by ANTLR 4.13.2. DO NOT EDIT.

package getfacl // Getfacl
import (
	"fmt"
	"strconv"
	"sync"

	"github.com/antlr4-go/antlr/v4"
)

// Suppress unused import errors
var (
	_ = fmt.Printf
	_ = strconv.Itoa
	_ = sync.Once{}
)

type GetfaclParser struct {
	*antlr.BaseParser
}

var GetfaclParserStaticData struct {
	once                   sync.Once
	serializedATN          []int32
	LiteralNames           []string
	SymbolicNames          []string
	RuleNames              []string
	PredictionContextCache *antlr.PredictionContextCache
	atn                    *antlr.ATN
	decisionToDFA          []*antlr.DFA
}

func getfaclParserInit() {
	staticData := &GetfaclParserStaticData
	staticData.LiteralNames = []string{
		"", "'file'", "'owner'", "'default:'", "'user'", "'group'", "'other'",
		"'mask'", "", "", "'#'", "':'",
	}
	staticData.SymbolicNames = []string{
		"", "", "", "DEFAULT_PREFIX", "USER_TAG", "GROUP_TAG", "OTHER_TAG",
		"MASK_TAG", "PERM", "VALUE_ATOM", "HASH", "COLON", "WS", "NL",
	}
	staticData.RuleNames = []string{
		"document", "line", "headerLine", "commentLine", "aclEntryLine", "headerKey",
		"tag", "qualifier", "valueAtom",
	}
	staticData.PredictionContextCache = antlr.NewPredictionContextCache()
	staticData.serializedATN = []int32{
		4, 1, 13, 79, 2, 0, 7, 0, 2, 1, 7, 1, 2, 2, 7, 2, 2, 3, 7, 3, 2, 4, 7,
		4, 2, 5, 7, 5, 2, 6, 7, 6, 2, 7, 7, 7, 2, 8, 7, 8, 1, 0, 5, 0, 20, 8, 0,
		10, 0, 12, 0, 23, 9, 0, 1, 0, 1, 0, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
		1, 1, 1, 1, 1, 1, 1, 1, 3, 1, 37, 8, 1, 1, 2, 1, 2, 5, 2, 41, 8, 2, 10,
		2, 12, 2, 44, 9, 2, 1, 2, 1, 2, 1, 2, 5, 2, 49, 8, 2, 10, 2, 12, 2, 52,
		9, 2, 1, 2, 3, 2, 55, 8, 2, 1, 3, 1, 3, 3, 3, 59, 8, 3, 1, 4, 3, 4, 62,
		8, 4, 1, 4, 1, 4, 1, 4, 1, 4, 1, 4, 1, 4, 1, 5, 1, 5, 1, 6, 1, 6, 1, 7,
		3, 7, 75, 8, 7, 1, 8, 1, 8, 1, 8, 0, 0, 9, 0, 2, 4, 6, 8, 10, 12, 14, 16,
		0, 2, 2, 0, 1, 2, 5, 5, 1, 0, 4, 7, 79, 0, 21, 1, 0, 0, 0, 2, 36, 1, 0,
		0, 0, 4, 38, 1, 0, 0, 0, 6, 56, 1, 0, 0, 0, 8, 61, 1, 0, 0, 0, 10, 69,
		1, 0, 0, 0, 12, 71, 1, 0, 0, 0, 14, 74, 1, 0, 0, 0, 16, 76, 1, 0, 0, 0,
		18, 20, 3, 2, 1, 0, 19, 18, 1, 0, 0, 0, 20, 23, 1, 0, 0, 0, 21, 19, 1,
		0, 0, 0, 21, 22, 1, 0, 0, 0, 22, 24, 1, 0, 0, 0, 23, 21, 1, 0, 0, 0, 24,
		25, 5, 0, 0, 1, 25, 1, 1, 0, 0, 0, 26, 27, 3, 4, 2, 0, 27, 28, 5, 13, 0,
		0, 28, 37, 1, 0, 0, 0, 29, 30, 3, 8, 4, 0, 30, 31, 5, 13, 0, 0, 31, 37,
		1, 0, 0, 0, 32, 33, 3, 6, 3, 0, 33, 34, 5, 13, 0, 0, 34, 37, 1, 0, 0, 0,
		35, 37, 5, 13, 0, 0, 36, 26, 1, 0, 0, 0, 36, 29, 1, 0, 0, 0, 36, 32, 1,
		0, 0, 0, 36, 35, 1, 0, 0, 0, 37, 3, 1, 0, 0, 0, 38, 42, 5, 10, 0, 0, 39,
		41, 5, 12, 0, 0, 40, 39, 1, 0, 0, 0, 41, 44, 1, 0, 0, 0, 42, 40, 1, 0,
		0, 0, 42, 43, 1, 0, 0, 0, 43, 45, 1, 0, 0, 0, 44, 42, 1, 0, 0, 0, 45, 46,
		3, 10, 5, 0, 46, 50, 5, 11, 0, 0, 47, 49, 5, 12, 0, 0, 48, 47, 1, 0, 0,
		0, 49, 52, 1, 0, 0, 0, 50, 48, 1, 0, 0, 0, 50, 51, 1, 0, 0, 0, 51, 54,
		1, 0, 0, 0, 52, 50, 1, 0, 0, 0, 53, 55, 3, 16, 8, 0, 54, 53, 1, 0, 0, 0,
		54, 55, 1, 0, 0, 0, 55, 5, 1, 0, 0, 0, 56, 58, 5, 10, 0, 0, 57, 59, 3,
		16, 8, 0, 58, 57, 1, 0, 0, 0, 58, 59, 1, 0, 0, 0, 59, 7, 1, 0, 0, 0, 60,
		62, 5, 3, 0, 0, 61, 60, 1, 0, 0, 0, 61, 62, 1, 0, 0, 0, 62, 63, 1, 0, 0,
		0, 63, 64, 3, 12, 6, 0, 64, 65, 5, 11, 0, 0, 65, 66, 3, 14, 7, 0, 66, 67,
		5, 11, 0, 0, 67, 68, 5, 8, 0, 0, 68, 9, 1, 0, 0, 0, 69, 70, 7, 0, 0, 0,
		70, 11, 1, 0, 0, 0, 71, 72, 7, 1, 0, 0, 72, 13, 1, 0, 0, 0, 73, 75, 3,
		16, 8, 0, 74, 73, 1, 0, 0, 0, 74, 75, 1, 0, 0, 0, 75, 15, 1, 0, 0, 0, 76,
		77, 5, 9, 0, 0, 77, 17, 1, 0, 0, 0, 8, 21, 36, 42, 50, 54, 58, 61, 74,
	}
	deserializer := antlr.NewATNDeserializer(nil)
	staticData.atn = deserializer.Deserialize(staticData.serializedATN)
	atn := staticData.atn
	staticData.decisionToDFA = make([]*antlr.DFA, len(atn.DecisionToState))
	decisionToDFA := staticData.decisionToDFA
	for index, state := range atn.DecisionToState {
		decisionToDFA[index] = antlr.NewDFA(state, index)
	}
}

// GetfaclParserInit initializes any static state used to implement GetfaclParser. By default the
// static state used to implement the parser is lazily initialized during the first call to
// NewGetfaclParser(). You can call this function if you wish to initialize the static state ahead
// of time.
func GetfaclParserInit() {
	staticData := &GetfaclParserStaticData
	staticData.once.Do(getfaclParserInit)
}

// NewGetfaclParser produces a new parser instance for the optional input antlr.TokenStream.
func NewGetfaclParser(input antlr.TokenStream) *GetfaclParser {
	GetfaclParserInit()
	this := new(GetfaclParser)
	this.BaseParser = antlr.NewBaseParser(input)
	staticData := &GetfaclParserStaticData
	this.Interpreter = antlr.NewParserATNSimulator(this, staticData.atn, staticData.decisionToDFA, staticData.PredictionContextCache)
	this.RuleNames = staticData.RuleNames
	this.LiteralNames = staticData.LiteralNames
	this.SymbolicNames = staticData.SymbolicNames
	this.GrammarFileName = "Getfacl.g4"

	return this
}

// GetfaclParser tokens.
const (
	GetfaclParserEOF            = antlr.TokenEOF
	GetfaclParserT__0           = 1
	GetfaclParserT__1           = 2
	GetfaclParserDEFAULT_PREFIX = 3
	GetfaclParserUSER_TAG       = 4
	GetfaclParserGROUP_TAG      = 5
	GetfaclParserOTHER_TAG      = 6
	GetfaclParserMASK_TAG       = 7
	GetfaclParserPERM           = 8
	GetfaclParserVALUE_ATOM     = 9
	GetfaclParserHASH           = 10
	GetfaclParserCOLON          = 11
	GetfaclParserWS             = 12
	GetfaclParserNL             = 13
)

// GetfaclParser rules.
const (
	GetfaclParserRULE_document     = 0
	GetfaclParserRULE_line         = 1
	GetfaclParserRULE_headerLine   = 2
	GetfaclParserRULE_commentLine  = 3
	GetfaclParserRULE_aclEntryLine = 4
	GetfaclParserRULE_headerKey    = 5
	GetfaclParserRULE_tag          = 6
	GetfaclParserRULE_qualifier    = 7
	GetfaclParserRULE_valueAtom    = 8
)

// IDocumentContext is an interface to support dynamic dispatch.
type IDocumentContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	EOF() antlr.TerminalNode
	AllLine() []ILineContext
	Line(i int) ILineContext

	// IsDocumentContext differentiates from other interfaces.
	IsDocumentContext()
}

type DocumentContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyDocumentContext() *DocumentContext {
	p := new(DocumentContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = GetfaclParserRULE_document
	return p
}

func InitEmptyDocumentContext(p *DocumentContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = GetfaclParserRULE_document
}

func (*DocumentContext) IsDocumentContext() {}

func NewDocumentContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *DocumentContext {
	p := new(DocumentContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = GetfaclParserRULE_document

	return p
}

func (s *DocumentContext) GetParser() antlr.Parser { return s.parser }

func (s *DocumentContext) EOF() antlr.TerminalNode {
	return s.GetToken(GetfaclParserEOF, 0)
}

func (s *DocumentContext) AllLine() []ILineContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(ILineContext); ok {
			len++
		}
	}

	tst := make([]ILineContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(ILineContext); ok {
			tst[i] = t.(ILineContext)
			i++
		}
	}

	return tst
}

func (s *DocumentContext) Line(i int) ILineContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ILineContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(ILineContext)
}

func (s *DocumentContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *DocumentContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *DocumentContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(GetfaclListener); ok {
		listenerT.EnterDocument(s)
	}
}

func (s *DocumentContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(GetfaclListener); ok {
		listenerT.ExitDocument(s)
	}
}

func (p *GetfaclParser) Document() (localctx IDocumentContext) {
	localctx = NewDocumentContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 0, GetfaclParserRULE_document)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(21)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for (int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&9464) != 0 {
		{
			p.SetState(18)
			p.Line()
		}

		p.SetState(23)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(24)
		p.Match(GetfaclParserEOF)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// ILineContext is an interface to support dynamic dispatch.
type ILineContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	HeaderLine() IHeaderLineContext
	NL() antlr.TerminalNode
	AclEntryLine() IAclEntryLineContext
	CommentLine() ICommentLineContext

	// IsLineContext differentiates from other interfaces.
	IsLineContext()
}

type LineContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyLineContext() *LineContext {
	p := new(LineContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = GetfaclParserRULE_line
	return p
}

func InitEmptyLineContext(p *LineContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = GetfaclParserRULE_line
}

func (*LineContext) IsLineContext() {}

func NewLineContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *LineContext {
	p := new(LineContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = GetfaclParserRULE_line

	return p
}

func (s *LineContext) GetParser() antlr.Parser { return s.parser }

func (s *LineContext) HeaderLine() IHeaderLineContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IHeaderLineContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IHeaderLineContext)
}

func (s *LineContext) NL() antlr.TerminalNode {
	return s.GetToken(GetfaclParserNL, 0)
}

func (s *LineContext) AclEntryLine() IAclEntryLineContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IAclEntryLineContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IAclEntryLineContext)
}

func (s *LineContext) CommentLine() ICommentLineContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ICommentLineContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ICommentLineContext)
}

func (s *LineContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *LineContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *LineContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(GetfaclListener); ok {
		listenerT.EnterLine(s)
	}
}

func (s *LineContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(GetfaclListener); ok {
		listenerT.ExitLine(s)
	}
}

func (p *GetfaclParser) Line() (localctx ILineContext) {
	localctx = NewLineContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 2, GetfaclParserRULE_line)
	p.SetState(36)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 1, p.GetParserRuleContext()) {
	case 1:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(26)
			p.HeaderLine()
		}
		{
			p.SetState(27)
			p.Match(GetfaclParserNL)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 2:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(29)
			p.AclEntryLine()
		}
		{
			p.SetState(30)
			p.Match(GetfaclParserNL)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 3:
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(32)
			p.CommentLine()
		}
		{
			p.SetState(33)
			p.Match(GetfaclParserNL)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 4:
		p.EnterOuterAlt(localctx, 4)
		{
			p.SetState(35)
			p.Match(GetfaclParserNL)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case antlr.ATNInvalidAltNumber:
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IHeaderLineContext is an interface to support dynamic dispatch.
type IHeaderLineContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	HASH() antlr.TerminalNode
	HeaderKey() IHeaderKeyContext
	COLON() antlr.TerminalNode
	AllWS() []antlr.TerminalNode
	WS(i int) antlr.TerminalNode
	ValueAtom() IValueAtomContext

	// IsHeaderLineContext differentiates from other interfaces.
	IsHeaderLineContext()
}

type HeaderLineContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyHeaderLineContext() *HeaderLineContext {
	p := new(HeaderLineContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = GetfaclParserRULE_headerLine
	return p
}

func InitEmptyHeaderLineContext(p *HeaderLineContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = GetfaclParserRULE_headerLine
}

func (*HeaderLineContext) IsHeaderLineContext() {}

func NewHeaderLineContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *HeaderLineContext {
	p := new(HeaderLineContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = GetfaclParserRULE_headerLine

	return p
}

func (s *HeaderLineContext) GetParser() antlr.Parser { return s.parser }

func (s *HeaderLineContext) HASH() antlr.TerminalNode {
	return s.GetToken(GetfaclParserHASH, 0)
}

func (s *HeaderLineContext) HeaderKey() IHeaderKeyContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IHeaderKeyContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IHeaderKeyContext)
}

func (s *HeaderLineContext) COLON() antlr.TerminalNode {
	return s.GetToken(GetfaclParserCOLON, 0)
}

func (s *HeaderLineContext) AllWS() []antlr.TerminalNode {
	return s.GetTokens(GetfaclParserWS)
}

func (s *HeaderLineContext) WS(i int) antlr.TerminalNode {
	return s.GetToken(GetfaclParserWS, i)
}

func (s *HeaderLineContext) ValueAtom() IValueAtomContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IValueAtomContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IValueAtomContext)
}

func (s *HeaderLineContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *HeaderLineContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *HeaderLineContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(GetfaclListener); ok {
		listenerT.EnterHeaderLine(s)
	}
}

func (s *HeaderLineContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(GetfaclListener); ok {
		listenerT.ExitHeaderLine(s)
	}
}

func (p *GetfaclParser) HeaderLine() (localctx IHeaderLineContext) {
	localctx = NewHeaderLineContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 4, GetfaclParserRULE_headerLine)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(38)
		p.Match(GetfaclParserHASH)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(42)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == GetfaclParserWS {
		{
			p.SetState(39)
			p.Match(GetfaclParserWS)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

		p.SetState(44)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(45)
		p.HeaderKey()
	}
	{
		p.SetState(46)
		p.Match(GetfaclParserCOLON)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(50)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == GetfaclParserWS {
		{
			p.SetState(47)
			p.Match(GetfaclParserWS)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

		p.SetState(52)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	p.SetState(54)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == GetfaclParserVALUE_ATOM {
		{
			p.SetState(53)
			p.ValueAtom()
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// ICommentLineContext is an interface to support dynamic dispatch.
type ICommentLineContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	HASH() antlr.TerminalNode
	ValueAtom() IValueAtomContext

	// IsCommentLineContext differentiates from other interfaces.
	IsCommentLineContext()
}

type CommentLineContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyCommentLineContext() *CommentLineContext {
	p := new(CommentLineContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = GetfaclParserRULE_commentLine
	return p
}

func InitEmptyCommentLineContext(p *CommentLineContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = GetfaclParserRULE_commentLine
}

func (*CommentLineContext) IsCommentLineContext() {}

func NewCommentLineContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *CommentLineContext {
	p := new(CommentLineContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = GetfaclParserRULE_commentLine

	return p
}

func (s *CommentLineContext) GetParser() antlr.Parser { return s.parser }

func (s *CommentLineContext) HASH() antlr.TerminalNode {
	return s.GetToken(GetfaclParserHASH, 0)
}

func (s *CommentLineContext) ValueAtom() IValueAtomContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IValueAtomContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IValueAtomContext)
}

func (s *CommentLineContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *CommentLineContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *CommentLineContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(GetfaclListener); ok {
		listenerT.EnterCommentLine(s)
	}
}

func (s *CommentLineContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(GetfaclListener); ok {
		listenerT.ExitCommentLine(s)
	}
}

func (p *GetfaclParser) CommentLine() (localctx ICommentLineContext) {
	localctx = NewCommentLineContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 6, GetfaclParserRULE_commentLine)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(56)
		p.Match(GetfaclParserHASH)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(58)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == GetfaclParserVALUE_ATOM {
		{
			p.SetState(57)
			p.ValueAtom()
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IAclEntryLineContext is an interface to support dynamic dispatch.
type IAclEntryLineContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	Tag() ITagContext
	AllCOLON() []antlr.TerminalNode
	COLON(i int) antlr.TerminalNode
	Qualifier() IQualifierContext
	PERM() antlr.TerminalNode
	DEFAULT_PREFIX() antlr.TerminalNode

	// IsAclEntryLineContext differentiates from other interfaces.
	IsAclEntryLineContext()
}

type AclEntryLineContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyAclEntryLineContext() *AclEntryLineContext {
	p := new(AclEntryLineContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = GetfaclParserRULE_aclEntryLine
	return p
}

func InitEmptyAclEntryLineContext(p *AclEntryLineContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = GetfaclParserRULE_aclEntryLine
}

func (*AclEntryLineContext) IsAclEntryLineContext() {}

func NewAclEntryLineContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *AclEntryLineContext {
	p := new(AclEntryLineContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = GetfaclParserRULE_aclEntryLine

	return p
}

func (s *AclEntryLineContext) GetParser() antlr.Parser { return s.parser }

func (s *AclEntryLineContext) Tag() ITagContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ITagContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ITagContext)
}

func (s *AclEntryLineContext) AllCOLON() []antlr.TerminalNode {
	return s.GetTokens(GetfaclParserCOLON)
}

func (s *AclEntryLineContext) COLON(i int) antlr.TerminalNode {
	return s.GetToken(GetfaclParserCOLON, i)
}

func (s *AclEntryLineContext) Qualifier() IQualifierContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IQualifierContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IQualifierContext)
}

func (s *AclEntryLineContext) PERM() antlr.TerminalNode {
	return s.GetToken(GetfaclParserPERM, 0)
}

func (s *AclEntryLineContext) DEFAULT_PREFIX() antlr.TerminalNode {
	return s.GetToken(GetfaclParserDEFAULT_PREFIX, 0)
}

func (s *AclEntryLineContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *AclEntryLineContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *AclEntryLineContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(GetfaclListener); ok {
		listenerT.EnterAclEntryLine(s)
	}
}

func (s *AclEntryLineContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(GetfaclListener); ok {
		listenerT.ExitAclEntryLine(s)
	}
}

func (p *GetfaclParser) AclEntryLine() (localctx IAclEntryLineContext) {
	localctx = NewAclEntryLineContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 8, GetfaclParserRULE_aclEntryLine)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(61)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == GetfaclParserDEFAULT_PREFIX {
		{
			p.SetState(60)
			p.Match(GetfaclParserDEFAULT_PREFIX)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
	}
	{
		p.SetState(63)
		p.Tag()
	}
	{
		p.SetState(64)
		p.Match(GetfaclParserCOLON)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(65)
		p.Qualifier()
	}
	{
		p.SetState(66)
		p.Match(GetfaclParserCOLON)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(67)
		p.Match(GetfaclParserPERM)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IHeaderKeyContext is an interface to support dynamic dispatch.
type IHeaderKeyContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	GROUP_TAG() antlr.TerminalNode

	// IsHeaderKeyContext differentiates from other interfaces.
	IsHeaderKeyContext()
}

type HeaderKeyContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyHeaderKeyContext() *HeaderKeyContext {
	p := new(HeaderKeyContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = GetfaclParserRULE_headerKey
	return p
}

func InitEmptyHeaderKeyContext(p *HeaderKeyContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = GetfaclParserRULE_headerKey
}

func (*HeaderKeyContext) IsHeaderKeyContext() {}

func NewHeaderKeyContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *HeaderKeyContext {
	p := new(HeaderKeyContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = GetfaclParserRULE_headerKey

	return p
}

func (s *HeaderKeyContext) GetParser() antlr.Parser { return s.parser }

func (s *HeaderKeyContext) GROUP_TAG() antlr.TerminalNode {
	return s.GetToken(GetfaclParserGROUP_TAG, 0)
}

func (s *HeaderKeyContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *HeaderKeyContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *HeaderKeyContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(GetfaclListener); ok {
		listenerT.EnterHeaderKey(s)
	}
}

func (s *HeaderKeyContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(GetfaclListener); ok {
		listenerT.ExitHeaderKey(s)
	}
}

func (p *GetfaclParser) HeaderKey() (localctx IHeaderKeyContext) {
	localctx = NewHeaderKeyContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 10, GetfaclParserRULE_headerKey)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(69)
		_la = p.GetTokenStream().LA(1)

		if !((int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&38) != 0) {
			p.GetErrorHandler().RecoverInline(p)
		} else {
			p.GetErrorHandler().ReportMatch(p)
			p.Consume()
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// ITagContext is an interface to support dynamic dispatch.
type ITagContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	USER_TAG() antlr.TerminalNode
	GROUP_TAG() antlr.TerminalNode
	OTHER_TAG() antlr.TerminalNode
	MASK_TAG() antlr.TerminalNode

	// IsTagContext differentiates from other interfaces.
	IsTagContext()
}

type TagContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyTagContext() *TagContext {
	p := new(TagContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = GetfaclParserRULE_tag
	return p
}

func InitEmptyTagContext(p *TagContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = GetfaclParserRULE_tag
}

func (*TagContext) IsTagContext() {}

func NewTagContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *TagContext {
	p := new(TagContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = GetfaclParserRULE_tag

	return p
}

func (s *TagContext) GetParser() antlr.Parser { return s.parser }

func (s *TagContext) USER_TAG() antlr.TerminalNode {
	return s.GetToken(GetfaclParserUSER_TAG, 0)
}

func (s *TagContext) GROUP_TAG() antlr.TerminalNode {
	return s.GetToken(GetfaclParserGROUP_TAG, 0)
}

func (s *TagContext) OTHER_TAG() antlr.TerminalNode {
	return s.GetToken(GetfaclParserOTHER_TAG, 0)
}

func (s *TagContext) MASK_TAG() antlr.TerminalNode {
	return s.GetToken(GetfaclParserMASK_TAG, 0)
}

func (s *TagContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *TagContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *TagContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(GetfaclListener); ok {
		listenerT.EnterTag(s)
	}
}

func (s *TagContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(GetfaclListener); ok {
		listenerT.ExitTag(s)
	}
}

func (p *GetfaclParser) Tag() (localctx ITagContext) {
	localctx = NewTagContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 12, GetfaclParserRULE_tag)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(71)
		_la = p.GetTokenStream().LA(1)

		if !((int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&240) != 0) {
			p.GetErrorHandler().RecoverInline(p)
		} else {
			p.GetErrorHandler().ReportMatch(p)
			p.Consume()
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IQualifierContext is an interface to support dynamic dispatch.
type IQualifierContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	ValueAtom() IValueAtomContext

	// IsQualifierContext differentiates from other interfaces.
	IsQualifierContext()
}

type QualifierContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyQualifierContext() *QualifierContext {
	p := new(QualifierContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = GetfaclParserRULE_qualifier
	return p
}

func InitEmptyQualifierContext(p *QualifierContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = GetfaclParserRULE_qualifier
}

func (*QualifierContext) IsQualifierContext() {}

func NewQualifierContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *QualifierContext {
	p := new(QualifierContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = GetfaclParserRULE_qualifier

	return p
}

func (s *QualifierContext) GetParser() antlr.Parser { return s.parser }

func (s *QualifierContext) ValueAtom() IValueAtomContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IValueAtomContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IValueAtomContext)
}

func (s *QualifierContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *QualifierContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *QualifierContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(GetfaclListener); ok {
		listenerT.EnterQualifier(s)
	}
}

func (s *QualifierContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(GetfaclListener); ok {
		listenerT.ExitQualifier(s)
	}
}

func (p *GetfaclParser) Qualifier() (localctx IQualifierContext) {
	localctx = NewQualifierContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 14, GetfaclParserRULE_qualifier)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(74)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == GetfaclParserVALUE_ATOM {
		{
			p.SetState(73)
			p.ValueAtom()
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IValueAtomContext is an interface to support dynamic dispatch.
type IValueAtomContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	VALUE_ATOM() antlr.TerminalNode

	// IsValueAtomContext differentiates from other interfaces.
	IsValueAtomContext()
}

type ValueAtomContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyValueAtomContext() *ValueAtomContext {
	p := new(ValueAtomContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = GetfaclParserRULE_valueAtom
	return p
}

func InitEmptyValueAtomContext(p *ValueAtomContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = GetfaclParserRULE_valueAtom
}

func (*ValueAtomContext) IsValueAtomContext() {}

func NewValueAtomContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ValueAtomContext {
	p := new(ValueAtomContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = GetfaclParserRULE_valueAtom

	return p
}

func (s *ValueAtomContext) GetParser() antlr.Parser { return s.parser }

func (s *ValueAtomContext) VALUE_ATOM() antlr.TerminalNode {
	return s.GetToken(GetfaclParserVALUE_ATOM, 0)
}

func (s *ValueAtomContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ValueAtomContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *ValueAtomContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(GetfaclListener); ok {
		listenerT.EnterValueAtom(s)
	}
}

func (s *ValueAtomContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(GetfaclListener); ok {
		listenerT.ExitValueAtom(s)
	}
}

func (p *GetfaclParser) ValueAtom() (localctx IValueAtomContext) {
	localctx = NewValueAtomContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 16, GetfaclParserRULE_valueAtom)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(76)
		p.Match(GetfaclParserVALUE_ATOM)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}
