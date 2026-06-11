// Code generated from /home/oleksiybondar/Documents/development/lite-nas/shared/go/parsers/zfs/status/grammar/ZpoolStatus.g4 by ANTLR 4.13.2. DO NOT EDIT.

package zpoolstatus // ZpoolStatus
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

type ZpoolStatusParser struct {
	*antlr.BaseParser
}

var ZpoolStatusParserStaticData struct {
	once                   sync.Once
	serializedATN          []int32
	LiteralNames           []string
	SymbolicNames          []string
	RuleNames              []string
	PredictionContextCache *antlr.PredictionContextCache
	atn                    *antlr.ATN
	decisionToDFA          []*antlr.DFA
}

func zpoolstatusParserInit() {
	staticData := &ZpoolStatusParserStaticData
	staticData.LiteralNames = []string{
		"", "'pool:'", "'state:'", "'scan:'", "'config:'", "'errors:'", "'status:'",
		"'action:'", "'see:'",
	}
	staticData.SymbolicNames = []string{
		"", "POOL_KV", "STATE_KV", "SCAN_KV", "CONFIG_KV", "ERRORS_KV", "STATUS_KV",
		"ACTION_KV", "SEE_KV", "ATOM", "WS", "NL",
	}
	staticData.RuleNames = []string{
		"document", "leadingBlankLines", "poolBlock", "trailingBlankLines",
		"metadataSection", "metadataLine", "poolLine", "stateLine", "scanLine",
		"statusLine", "actionLine", "seeLine", "configSection", "configHeaderLine",
		"configRowLine", "errorsLine", "textLine", "textAtom", "headerAtom",
		"rowAtom",
	}
	staticData.PredictionContextCache = antlr.NewPredictionContextCache()
	staticData.serializedATN = []int32{
		4, 1, 11, 151, 2, 0, 7, 0, 2, 1, 7, 1, 2, 2, 7, 2, 2, 3, 7, 3, 2, 4, 7,
		4, 2, 5, 7, 5, 2, 6, 7, 6, 2, 7, 7, 7, 2, 8, 7, 8, 2, 9, 7, 9, 2, 10, 7,
		10, 2, 11, 7, 11, 2, 12, 7, 12, 2, 13, 7, 13, 2, 14, 7, 14, 2, 15, 7, 15,
		2, 16, 7, 16, 2, 17, 7, 17, 2, 18, 7, 18, 2, 19, 7, 19, 1, 0, 3, 0, 42,
		8, 0, 1, 0, 4, 0, 45, 8, 0, 11, 0, 12, 0, 46, 1, 0, 1, 0, 1, 1, 4, 1, 52,
		8, 1, 11, 1, 12, 1, 53, 1, 2, 1, 2, 3, 2, 58, 8, 2, 1, 2, 1, 2, 5, 2, 62,
		8, 2, 10, 2, 12, 2, 65, 9, 2, 1, 2, 1, 2, 3, 2, 69, 8, 2, 1, 3, 4, 3, 72,
		8, 3, 11, 3, 12, 3, 73, 1, 4, 4, 4, 77, 8, 4, 11, 4, 12, 4, 78, 1, 5, 1,
		5, 1, 5, 1, 5, 1, 5, 3, 5, 86, 8, 5, 1, 6, 1, 6, 1, 6, 1, 7, 1, 7, 1, 7,
		1, 8, 1, 8, 1, 8, 1, 9, 1, 9, 1, 9, 1, 10, 1, 10, 1, 10, 1, 11, 1, 11,
		1, 11, 1, 12, 1, 12, 1, 12, 5, 12, 109, 8, 12, 10, 12, 12, 12, 112, 9,
		12, 1, 12, 1, 12, 4, 12, 116, 8, 12, 11, 12, 12, 12, 117, 1, 13, 4, 13,
		121, 8, 13, 11, 13, 12, 13, 122, 1, 13, 1, 13, 1, 14, 4, 14, 128, 8, 14,
		11, 14, 12, 14, 129, 1, 14, 1, 14, 1, 15, 1, 15, 1, 15, 1, 16, 5, 16, 138,
		8, 16, 10, 16, 12, 16, 141, 9, 16, 1, 16, 1, 16, 1, 17, 1, 17, 1, 18, 1,
		18, 1, 19, 1, 19, 1, 19, 0, 0, 20, 0, 2, 4, 6, 8, 10, 12, 14, 16, 18, 20,
		22, 24, 26, 28, 30, 32, 34, 36, 38, 0, 0, 147, 0, 41, 1, 0, 0, 0, 2, 51,
		1, 0, 0, 0, 4, 55, 1, 0, 0, 0, 6, 71, 1, 0, 0, 0, 8, 76, 1, 0, 0, 0, 10,
		85, 1, 0, 0, 0, 12, 87, 1, 0, 0, 0, 14, 90, 1, 0, 0, 0, 16, 93, 1, 0, 0,
		0, 18, 96, 1, 0, 0, 0, 20, 99, 1, 0, 0, 0, 22, 102, 1, 0, 0, 0, 24, 105,
		1, 0, 0, 0, 26, 120, 1, 0, 0, 0, 28, 127, 1, 0, 0, 0, 30, 133, 1, 0, 0,
		0, 32, 139, 1, 0, 0, 0, 34, 144, 1, 0, 0, 0, 36, 146, 1, 0, 0, 0, 38, 148,
		1, 0, 0, 0, 40, 42, 3, 2, 1, 0, 41, 40, 1, 0, 0, 0, 41, 42, 1, 0, 0, 0,
		42, 44, 1, 0, 0, 0, 43, 45, 3, 4, 2, 0, 44, 43, 1, 0, 0, 0, 45, 46, 1,
		0, 0, 0, 46, 44, 1, 0, 0, 0, 46, 47, 1, 0, 0, 0, 47, 48, 1, 0, 0, 0, 48,
		49, 5, 0, 0, 1, 49, 1, 1, 0, 0, 0, 50, 52, 5, 11, 0, 0, 51, 50, 1, 0, 0,
		0, 52, 53, 1, 0, 0, 0, 53, 51, 1, 0, 0, 0, 53, 54, 1, 0, 0, 0, 54, 3, 1,
		0, 0, 0, 55, 57, 3, 12, 6, 0, 56, 58, 3, 8, 4, 0, 57, 56, 1, 0, 0, 0, 57,
		58, 1, 0, 0, 0, 58, 59, 1, 0, 0, 0, 59, 63, 3, 24, 12, 0, 60, 62, 5, 11,
		0, 0, 61, 60, 1, 0, 0, 0, 62, 65, 1, 0, 0, 0, 63, 61, 1, 0, 0, 0, 63, 64,
		1, 0, 0, 0, 64, 66, 1, 0, 0, 0, 65, 63, 1, 0, 0, 0, 66, 68, 3, 30, 15,
		0, 67, 69, 3, 6, 3, 0, 68, 67, 1, 0, 0, 0, 68, 69, 1, 0, 0, 0, 69, 5, 1,
		0, 0, 0, 70, 72, 5, 11, 0, 0, 71, 70, 1, 0, 0, 0, 72, 73, 1, 0, 0, 0, 73,
		71, 1, 0, 0, 0, 73, 74, 1, 0, 0, 0, 74, 7, 1, 0, 0, 0, 75, 77, 3, 10, 5,
		0, 76, 75, 1, 0, 0, 0, 77, 78, 1, 0, 0, 0, 78, 76, 1, 0, 0, 0, 78, 79,
		1, 0, 0, 0, 79, 9, 1, 0, 0, 0, 80, 86, 3, 14, 7, 0, 81, 86, 3, 16, 8, 0,
		82, 86, 3, 18, 9, 0, 83, 86, 3, 20, 10, 0, 84, 86, 3, 22, 11, 0, 85, 80,
		1, 0, 0, 0, 85, 81, 1, 0, 0, 0, 85, 82, 1, 0, 0, 0, 85, 83, 1, 0, 0, 0,
		85, 84, 1, 0, 0, 0, 86, 11, 1, 0, 0, 0, 87, 88, 5, 1, 0, 0, 88, 89, 3,
		32, 16, 0, 89, 13, 1, 0, 0, 0, 90, 91, 5, 2, 0, 0, 91, 92, 3, 32, 16, 0,
		92, 15, 1, 0, 0, 0, 93, 94, 5, 3, 0, 0, 94, 95, 3, 32, 16, 0, 95, 17, 1,
		0, 0, 0, 96, 97, 5, 6, 0, 0, 97, 98, 3, 32, 16, 0, 98, 19, 1, 0, 0, 0,
		99, 100, 5, 7, 0, 0, 100, 101, 3, 32, 16, 0, 101, 21, 1, 0, 0, 0, 102,
		103, 5, 8, 0, 0, 103, 104, 3, 32, 16, 0, 104, 23, 1, 0, 0, 0, 105, 106,
		5, 4, 0, 0, 106, 110, 5, 11, 0, 0, 107, 109, 5, 11, 0, 0, 108, 107, 1,
		0, 0, 0, 109, 112, 1, 0, 0, 0, 110, 108, 1, 0, 0, 0, 110, 111, 1, 0, 0,
		0, 111, 113, 1, 0, 0, 0, 112, 110, 1, 0, 0, 0, 113, 115, 3, 26, 13, 0,
		114, 116, 3, 28, 14, 0, 115, 114, 1, 0, 0, 0, 116, 117, 1, 0, 0, 0, 117,
		115, 1, 0, 0, 0, 117, 118, 1, 0, 0, 0, 118, 25, 1, 0, 0, 0, 119, 121, 3,
		36, 18, 0, 120, 119, 1, 0, 0, 0, 121, 122, 1, 0, 0, 0, 122, 120, 1, 0,
		0, 0, 122, 123, 1, 0, 0, 0, 123, 124, 1, 0, 0, 0, 124, 125, 5, 11, 0, 0,
		125, 27, 1, 0, 0, 0, 126, 128, 3, 38, 19, 0, 127, 126, 1, 0, 0, 0, 128,
		129, 1, 0, 0, 0, 129, 127, 1, 0, 0, 0, 129, 130, 1, 0, 0, 0, 130, 131,
		1, 0, 0, 0, 131, 132, 5, 11, 0, 0, 132, 29, 1, 0, 0, 0, 133, 134, 5, 5,
		0, 0, 134, 135, 3, 32, 16, 0, 135, 31, 1, 0, 0, 0, 136, 138, 3, 34, 17,
		0, 137, 136, 1, 0, 0, 0, 138, 141, 1, 0, 0, 0, 139, 137, 1, 0, 0, 0, 139,
		140, 1, 0, 0, 0, 140, 142, 1, 0, 0, 0, 141, 139, 1, 0, 0, 0, 142, 143,
		5, 11, 0, 0, 143, 33, 1, 0, 0, 0, 144, 145, 5, 9, 0, 0, 145, 35, 1, 0,
		0, 0, 146, 147, 5, 9, 0, 0, 147, 37, 1, 0, 0, 0, 148, 149, 5, 9, 0, 0,
		149, 39, 1, 0, 0, 0, 14, 41, 46, 53, 57, 63, 68, 73, 78, 85, 110, 117,
		122, 129, 139,
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

// ZpoolStatusParserInit initializes any static state used to implement ZpoolStatusParser. By default the
// static state used to implement the parser is lazily initialized during the first call to
// NewZpoolStatusParser(). You can call this function if you wish to initialize the static state ahead
// of time.
func ZpoolStatusParserInit() {
	staticData := &ZpoolStatusParserStaticData
	staticData.once.Do(zpoolstatusParserInit)
}

// NewZpoolStatusParser produces a new parser instance for the optional input antlr.TokenStream.
func NewZpoolStatusParser(input antlr.TokenStream) *ZpoolStatusParser {
	ZpoolStatusParserInit()
	this := new(ZpoolStatusParser)
	this.BaseParser = antlr.NewBaseParser(input)
	staticData := &ZpoolStatusParserStaticData
	this.Interpreter = antlr.NewParserATNSimulator(this, staticData.atn, staticData.decisionToDFA, staticData.PredictionContextCache)
	this.RuleNames = staticData.RuleNames
	this.LiteralNames = staticData.LiteralNames
	this.SymbolicNames = staticData.SymbolicNames
	this.GrammarFileName = "ZpoolStatus.g4"

	return this
}

// ZpoolStatusParser tokens.
const (
	ZpoolStatusParserEOF       = antlr.TokenEOF
	ZpoolStatusParserPOOL_KV   = 1
	ZpoolStatusParserSTATE_KV  = 2
	ZpoolStatusParserSCAN_KV   = 3
	ZpoolStatusParserCONFIG_KV = 4
	ZpoolStatusParserERRORS_KV = 5
	ZpoolStatusParserSTATUS_KV = 6
	ZpoolStatusParserACTION_KV = 7
	ZpoolStatusParserSEE_KV    = 8
	ZpoolStatusParserATOM      = 9
	ZpoolStatusParserWS        = 10
	ZpoolStatusParserNL        = 11
)

// ZpoolStatusParser rules.
const (
	ZpoolStatusParserRULE_document           = 0
	ZpoolStatusParserRULE_leadingBlankLines  = 1
	ZpoolStatusParserRULE_poolBlock          = 2
	ZpoolStatusParserRULE_trailingBlankLines = 3
	ZpoolStatusParserRULE_metadataSection    = 4
	ZpoolStatusParserRULE_metadataLine       = 5
	ZpoolStatusParserRULE_poolLine           = 6
	ZpoolStatusParserRULE_stateLine          = 7
	ZpoolStatusParserRULE_scanLine           = 8
	ZpoolStatusParserRULE_statusLine         = 9
	ZpoolStatusParserRULE_actionLine         = 10
	ZpoolStatusParserRULE_seeLine            = 11
	ZpoolStatusParserRULE_configSection      = 12
	ZpoolStatusParserRULE_configHeaderLine   = 13
	ZpoolStatusParserRULE_configRowLine      = 14
	ZpoolStatusParserRULE_errorsLine         = 15
	ZpoolStatusParserRULE_textLine           = 16
	ZpoolStatusParserRULE_textAtom           = 17
	ZpoolStatusParserRULE_headerAtom         = 18
	ZpoolStatusParserRULE_rowAtom            = 19
)

// IDocumentContext is an interface to support dynamic dispatch.
type IDocumentContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	EOF() antlr.TerminalNode
	LeadingBlankLines() ILeadingBlankLinesContext
	AllPoolBlock() []IPoolBlockContext
	PoolBlock(i int) IPoolBlockContext

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
	p.RuleIndex = ZpoolStatusParserRULE_document
	return p
}

func InitEmptyDocumentContext(p *DocumentContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ZpoolStatusParserRULE_document
}

func (*DocumentContext) IsDocumentContext() {}

func NewDocumentContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *DocumentContext {
	p := new(DocumentContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = ZpoolStatusParserRULE_document

	return p
}

func (s *DocumentContext) GetParser() antlr.Parser { return s.parser }

func (s *DocumentContext) EOF() antlr.TerminalNode {
	return s.GetToken(ZpoolStatusParserEOF, 0)
}

func (s *DocumentContext) LeadingBlankLines() ILeadingBlankLinesContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ILeadingBlankLinesContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ILeadingBlankLinesContext)
}

func (s *DocumentContext) AllPoolBlock() []IPoolBlockContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IPoolBlockContext); ok {
			len++
		}
	}

	tst := make([]IPoolBlockContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IPoolBlockContext); ok {
			tst[i] = t.(IPoolBlockContext)
			i++
		}
	}

	return tst
}

func (s *DocumentContext) PoolBlock(i int) IPoolBlockContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IPoolBlockContext); ok {
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

	return t.(IPoolBlockContext)
}

func (s *DocumentContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *DocumentContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *DocumentContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ZpoolStatusListener); ok {
		listenerT.EnterDocument(s)
	}
}

func (s *DocumentContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ZpoolStatusListener); ok {
		listenerT.ExitDocument(s)
	}
}

func (p *ZpoolStatusParser) Document() (localctx IDocumentContext) {
	localctx = NewDocumentContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 0, ZpoolStatusParserRULE_document)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(41)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == ZpoolStatusParserNL {
		{
			p.SetState(40)
			p.LeadingBlankLines()
		}
	}
	p.SetState(44)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for ok := true; ok; ok = _la == ZpoolStatusParserPOOL_KV {
		{
			p.SetState(43)
			p.PoolBlock()
		}

		p.SetState(46)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(48)
		p.Match(ZpoolStatusParserEOF)
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

// ILeadingBlankLinesContext is an interface to support dynamic dispatch.
type ILeadingBlankLinesContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AllNL() []antlr.TerminalNode
	NL(i int) antlr.TerminalNode

	// IsLeadingBlankLinesContext differentiates from other interfaces.
	IsLeadingBlankLinesContext()
}

type LeadingBlankLinesContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyLeadingBlankLinesContext() *LeadingBlankLinesContext {
	p := new(LeadingBlankLinesContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ZpoolStatusParserRULE_leadingBlankLines
	return p
}

func InitEmptyLeadingBlankLinesContext(p *LeadingBlankLinesContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ZpoolStatusParserRULE_leadingBlankLines
}

func (*LeadingBlankLinesContext) IsLeadingBlankLinesContext() {}

func NewLeadingBlankLinesContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *LeadingBlankLinesContext {
	p := new(LeadingBlankLinesContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = ZpoolStatusParserRULE_leadingBlankLines

	return p
}

func (s *LeadingBlankLinesContext) GetParser() antlr.Parser { return s.parser }

func (s *LeadingBlankLinesContext) AllNL() []antlr.TerminalNode {
	return s.GetTokens(ZpoolStatusParserNL)
}

func (s *LeadingBlankLinesContext) NL(i int) antlr.TerminalNode {
	return s.GetToken(ZpoolStatusParserNL, i)
}

func (s *LeadingBlankLinesContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *LeadingBlankLinesContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *LeadingBlankLinesContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ZpoolStatusListener); ok {
		listenerT.EnterLeadingBlankLines(s)
	}
}

func (s *LeadingBlankLinesContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ZpoolStatusListener); ok {
		listenerT.ExitLeadingBlankLines(s)
	}
}

func (p *ZpoolStatusParser) LeadingBlankLines() (localctx ILeadingBlankLinesContext) {
	localctx = NewLeadingBlankLinesContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 2, ZpoolStatusParserRULE_leadingBlankLines)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(51)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for ok := true; ok; ok = _la == ZpoolStatusParserNL {
		{
			p.SetState(50)
			p.Match(ZpoolStatusParserNL)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

		p.SetState(53)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
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

// IPoolBlockContext is an interface to support dynamic dispatch.
type IPoolBlockContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	PoolLine() IPoolLineContext
	ConfigSection() IConfigSectionContext
	ErrorsLine() IErrorsLineContext
	MetadataSection() IMetadataSectionContext
	AllNL() []antlr.TerminalNode
	NL(i int) antlr.TerminalNode
	TrailingBlankLines() ITrailingBlankLinesContext

	// IsPoolBlockContext differentiates from other interfaces.
	IsPoolBlockContext()
}

type PoolBlockContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyPoolBlockContext() *PoolBlockContext {
	p := new(PoolBlockContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ZpoolStatusParserRULE_poolBlock
	return p
}

func InitEmptyPoolBlockContext(p *PoolBlockContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ZpoolStatusParserRULE_poolBlock
}

func (*PoolBlockContext) IsPoolBlockContext() {}

func NewPoolBlockContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *PoolBlockContext {
	p := new(PoolBlockContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = ZpoolStatusParserRULE_poolBlock

	return p
}

func (s *PoolBlockContext) GetParser() antlr.Parser { return s.parser }

func (s *PoolBlockContext) PoolLine() IPoolLineContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IPoolLineContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IPoolLineContext)
}

func (s *PoolBlockContext) ConfigSection() IConfigSectionContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IConfigSectionContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IConfigSectionContext)
}

func (s *PoolBlockContext) ErrorsLine() IErrorsLineContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IErrorsLineContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IErrorsLineContext)
}

func (s *PoolBlockContext) MetadataSection() IMetadataSectionContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IMetadataSectionContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IMetadataSectionContext)
}

func (s *PoolBlockContext) AllNL() []antlr.TerminalNode {
	return s.GetTokens(ZpoolStatusParserNL)
}

func (s *PoolBlockContext) NL(i int) antlr.TerminalNode {
	return s.GetToken(ZpoolStatusParserNL, i)
}

func (s *PoolBlockContext) TrailingBlankLines() ITrailingBlankLinesContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ITrailingBlankLinesContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ITrailingBlankLinesContext)
}

func (s *PoolBlockContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *PoolBlockContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *PoolBlockContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ZpoolStatusListener); ok {
		listenerT.EnterPoolBlock(s)
	}
}

func (s *PoolBlockContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ZpoolStatusListener); ok {
		listenerT.ExitPoolBlock(s)
	}
}

func (p *ZpoolStatusParser) PoolBlock() (localctx IPoolBlockContext) {
	localctx = NewPoolBlockContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 4, ZpoolStatusParserRULE_poolBlock)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(55)
		p.PoolLine()
	}
	p.SetState(57)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if (int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&460) != 0 {
		{
			p.SetState(56)
			p.MetadataSection()
		}
	}
	{
		p.SetState(59)
		p.ConfigSection()
	}
	p.SetState(63)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == ZpoolStatusParserNL {
		{
			p.SetState(60)
			p.Match(ZpoolStatusParserNL)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

		p.SetState(65)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(66)
		p.ErrorsLine()
	}
	p.SetState(68)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == ZpoolStatusParserNL {
		{
			p.SetState(67)
			p.TrailingBlankLines()
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

// ITrailingBlankLinesContext is an interface to support dynamic dispatch.
type ITrailingBlankLinesContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AllNL() []antlr.TerminalNode
	NL(i int) antlr.TerminalNode

	// IsTrailingBlankLinesContext differentiates from other interfaces.
	IsTrailingBlankLinesContext()
}

type TrailingBlankLinesContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyTrailingBlankLinesContext() *TrailingBlankLinesContext {
	p := new(TrailingBlankLinesContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ZpoolStatusParserRULE_trailingBlankLines
	return p
}

func InitEmptyTrailingBlankLinesContext(p *TrailingBlankLinesContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ZpoolStatusParserRULE_trailingBlankLines
}

func (*TrailingBlankLinesContext) IsTrailingBlankLinesContext() {}

func NewTrailingBlankLinesContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *TrailingBlankLinesContext {
	p := new(TrailingBlankLinesContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = ZpoolStatusParserRULE_trailingBlankLines

	return p
}

func (s *TrailingBlankLinesContext) GetParser() antlr.Parser { return s.parser }

func (s *TrailingBlankLinesContext) AllNL() []antlr.TerminalNode {
	return s.GetTokens(ZpoolStatusParserNL)
}

func (s *TrailingBlankLinesContext) NL(i int) antlr.TerminalNode {
	return s.GetToken(ZpoolStatusParserNL, i)
}

func (s *TrailingBlankLinesContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *TrailingBlankLinesContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *TrailingBlankLinesContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ZpoolStatusListener); ok {
		listenerT.EnterTrailingBlankLines(s)
	}
}

func (s *TrailingBlankLinesContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ZpoolStatusListener); ok {
		listenerT.ExitTrailingBlankLines(s)
	}
}

func (p *ZpoolStatusParser) TrailingBlankLines() (localctx ITrailingBlankLinesContext) {
	localctx = NewTrailingBlankLinesContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 6, ZpoolStatusParserRULE_trailingBlankLines)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(71)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for ok := true; ok; ok = _la == ZpoolStatusParserNL {
		{
			p.SetState(70)
			p.Match(ZpoolStatusParserNL)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

		p.SetState(73)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
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

// IMetadataSectionContext is an interface to support dynamic dispatch.
type IMetadataSectionContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AllMetadataLine() []IMetadataLineContext
	MetadataLine(i int) IMetadataLineContext

	// IsMetadataSectionContext differentiates from other interfaces.
	IsMetadataSectionContext()
}

type MetadataSectionContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyMetadataSectionContext() *MetadataSectionContext {
	p := new(MetadataSectionContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ZpoolStatusParserRULE_metadataSection
	return p
}

func InitEmptyMetadataSectionContext(p *MetadataSectionContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ZpoolStatusParserRULE_metadataSection
}

func (*MetadataSectionContext) IsMetadataSectionContext() {}

func NewMetadataSectionContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *MetadataSectionContext {
	p := new(MetadataSectionContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = ZpoolStatusParserRULE_metadataSection

	return p
}

func (s *MetadataSectionContext) GetParser() antlr.Parser { return s.parser }

func (s *MetadataSectionContext) AllMetadataLine() []IMetadataLineContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IMetadataLineContext); ok {
			len++
		}
	}

	tst := make([]IMetadataLineContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IMetadataLineContext); ok {
			tst[i] = t.(IMetadataLineContext)
			i++
		}
	}

	return tst
}

func (s *MetadataSectionContext) MetadataLine(i int) IMetadataLineContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IMetadataLineContext); ok {
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

	return t.(IMetadataLineContext)
}

func (s *MetadataSectionContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *MetadataSectionContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *MetadataSectionContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ZpoolStatusListener); ok {
		listenerT.EnterMetadataSection(s)
	}
}

func (s *MetadataSectionContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ZpoolStatusListener); ok {
		listenerT.ExitMetadataSection(s)
	}
}

func (p *ZpoolStatusParser) MetadataSection() (localctx IMetadataSectionContext) {
	localctx = NewMetadataSectionContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 8, ZpoolStatusParserRULE_metadataSection)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(76)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for ok := true; ok; ok = ((int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&460) != 0) {
		{
			p.SetState(75)
			p.MetadataLine()
		}

		p.SetState(78)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
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

// IMetadataLineContext is an interface to support dynamic dispatch.
type IMetadataLineContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	StateLine() IStateLineContext
	ScanLine() IScanLineContext
	StatusLine() IStatusLineContext
	ActionLine() IActionLineContext
	SeeLine() ISeeLineContext

	// IsMetadataLineContext differentiates from other interfaces.
	IsMetadataLineContext()
}

type MetadataLineContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyMetadataLineContext() *MetadataLineContext {
	p := new(MetadataLineContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ZpoolStatusParserRULE_metadataLine
	return p
}

func InitEmptyMetadataLineContext(p *MetadataLineContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ZpoolStatusParserRULE_metadataLine
}

func (*MetadataLineContext) IsMetadataLineContext() {}

func NewMetadataLineContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *MetadataLineContext {
	p := new(MetadataLineContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = ZpoolStatusParserRULE_metadataLine

	return p
}

func (s *MetadataLineContext) GetParser() antlr.Parser { return s.parser }

func (s *MetadataLineContext) StateLine() IStateLineContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IStateLineContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IStateLineContext)
}

func (s *MetadataLineContext) ScanLine() IScanLineContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IScanLineContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IScanLineContext)
}

func (s *MetadataLineContext) StatusLine() IStatusLineContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IStatusLineContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IStatusLineContext)
}

func (s *MetadataLineContext) ActionLine() IActionLineContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IActionLineContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IActionLineContext)
}

func (s *MetadataLineContext) SeeLine() ISeeLineContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ISeeLineContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ISeeLineContext)
}

func (s *MetadataLineContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *MetadataLineContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *MetadataLineContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ZpoolStatusListener); ok {
		listenerT.EnterMetadataLine(s)
	}
}

func (s *MetadataLineContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ZpoolStatusListener); ok {
		listenerT.ExitMetadataLine(s)
	}
}

func (p *ZpoolStatusParser) MetadataLine() (localctx IMetadataLineContext) {
	localctx = NewMetadataLineContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 10, ZpoolStatusParserRULE_metadataLine)
	p.SetState(85)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case ZpoolStatusParserSTATE_KV:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(80)
			p.StateLine()
		}

	case ZpoolStatusParserSCAN_KV:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(81)
			p.ScanLine()
		}

	case ZpoolStatusParserSTATUS_KV:
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(82)
			p.StatusLine()
		}

	case ZpoolStatusParserACTION_KV:
		p.EnterOuterAlt(localctx, 4)
		{
			p.SetState(83)
			p.ActionLine()
		}

	case ZpoolStatusParserSEE_KV:
		p.EnterOuterAlt(localctx, 5)
		{
			p.SetState(84)
			p.SeeLine()
		}

	default:
		p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
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

// IPoolLineContext is an interface to support dynamic dispatch.
type IPoolLineContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	POOL_KV() antlr.TerminalNode
	TextLine() ITextLineContext

	// IsPoolLineContext differentiates from other interfaces.
	IsPoolLineContext()
}

type PoolLineContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyPoolLineContext() *PoolLineContext {
	p := new(PoolLineContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ZpoolStatusParserRULE_poolLine
	return p
}

func InitEmptyPoolLineContext(p *PoolLineContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ZpoolStatusParserRULE_poolLine
}

func (*PoolLineContext) IsPoolLineContext() {}

func NewPoolLineContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *PoolLineContext {
	p := new(PoolLineContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = ZpoolStatusParserRULE_poolLine

	return p
}

func (s *PoolLineContext) GetParser() antlr.Parser { return s.parser }

func (s *PoolLineContext) POOL_KV() antlr.TerminalNode {
	return s.GetToken(ZpoolStatusParserPOOL_KV, 0)
}

func (s *PoolLineContext) TextLine() ITextLineContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ITextLineContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ITextLineContext)
}

func (s *PoolLineContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *PoolLineContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *PoolLineContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ZpoolStatusListener); ok {
		listenerT.EnterPoolLine(s)
	}
}

func (s *PoolLineContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ZpoolStatusListener); ok {
		listenerT.ExitPoolLine(s)
	}
}

func (p *ZpoolStatusParser) PoolLine() (localctx IPoolLineContext) {
	localctx = NewPoolLineContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 12, ZpoolStatusParserRULE_poolLine)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(87)
		p.Match(ZpoolStatusParserPOOL_KV)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(88)
		p.TextLine()
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

// IStateLineContext is an interface to support dynamic dispatch.
type IStateLineContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	STATE_KV() antlr.TerminalNode
	TextLine() ITextLineContext

	// IsStateLineContext differentiates from other interfaces.
	IsStateLineContext()
}

type StateLineContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyStateLineContext() *StateLineContext {
	p := new(StateLineContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ZpoolStatusParserRULE_stateLine
	return p
}

func InitEmptyStateLineContext(p *StateLineContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ZpoolStatusParserRULE_stateLine
}

func (*StateLineContext) IsStateLineContext() {}

func NewStateLineContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *StateLineContext {
	p := new(StateLineContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = ZpoolStatusParserRULE_stateLine

	return p
}

func (s *StateLineContext) GetParser() antlr.Parser { return s.parser }

func (s *StateLineContext) STATE_KV() antlr.TerminalNode {
	return s.GetToken(ZpoolStatusParserSTATE_KV, 0)
}

func (s *StateLineContext) TextLine() ITextLineContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ITextLineContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ITextLineContext)
}

func (s *StateLineContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *StateLineContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *StateLineContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ZpoolStatusListener); ok {
		listenerT.EnterStateLine(s)
	}
}

func (s *StateLineContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ZpoolStatusListener); ok {
		listenerT.ExitStateLine(s)
	}
}

func (p *ZpoolStatusParser) StateLine() (localctx IStateLineContext) {
	localctx = NewStateLineContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 14, ZpoolStatusParserRULE_stateLine)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(90)
		p.Match(ZpoolStatusParserSTATE_KV)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(91)
		p.TextLine()
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

// IScanLineContext is an interface to support dynamic dispatch.
type IScanLineContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	SCAN_KV() antlr.TerminalNode
	TextLine() ITextLineContext

	// IsScanLineContext differentiates from other interfaces.
	IsScanLineContext()
}

type ScanLineContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyScanLineContext() *ScanLineContext {
	p := new(ScanLineContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ZpoolStatusParserRULE_scanLine
	return p
}

func InitEmptyScanLineContext(p *ScanLineContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ZpoolStatusParserRULE_scanLine
}

func (*ScanLineContext) IsScanLineContext() {}

func NewScanLineContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ScanLineContext {
	p := new(ScanLineContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = ZpoolStatusParserRULE_scanLine

	return p
}

func (s *ScanLineContext) GetParser() antlr.Parser { return s.parser }

func (s *ScanLineContext) SCAN_KV() antlr.TerminalNode {
	return s.GetToken(ZpoolStatusParserSCAN_KV, 0)
}

func (s *ScanLineContext) TextLine() ITextLineContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ITextLineContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ITextLineContext)
}

func (s *ScanLineContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ScanLineContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *ScanLineContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ZpoolStatusListener); ok {
		listenerT.EnterScanLine(s)
	}
}

func (s *ScanLineContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ZpoolStatusListener); ok {
		listenerT.ExitScanLine(s)
	}
}

func (p *ZpoolStatusParser) ScanLine() (localctx IScanLineContext) {
	localctx = NewScanLineContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 16, ZpoolStatusParserRULE_scanLine)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(93)
		p.Match(ZpoolStatusParserSCAN_KV)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(94)
		p.TextLine()
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

// IStatusLineContext is an interface to support dynamic dispatch.
type IStatusLineContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	STATUS_KV() antlr.TerminalNode
	TextLine() ITextLineContext

	// IsStatusLineContext differentiates from other interfaces.
	IsStatusLineContext()
}

type StatusLineContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyStatusLineContext() *StatusLineContext {
	p := new(StatusLineContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ZpoolStatusParserRULE_statusLine
	return p
}

func InitEmptyStatusLineContext(p *StatusLineContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ZpoolStatusParserRULE_statusLine
}

func (*StatusLineContext) IsStatusLineContext() {}

func NewStatusLineContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *StatusLineContext {
	p := new(StatusLineContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = ZpoolStatusParserRULE_statusLine

	return p
}

func (s *StatusLineContext) GetParser() antlr.Parser { return s.parser }

func (s *StatusLineContext) STATUS_KV() antlr.TerminalNode {
	return s.GetToken(ZpoolStatusParserSTATUS_KV, 0)
}

func (s *StatusLineContext) TextLine() ITextLineContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ITextLineContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ITextLineContext)
}

func (s *StatusLineContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *StatusLineContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *StatusLineContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ZpoolStatusListener); ok {
		listenerT.EnterStatusLine(s)
	}
}

func (s *StatusLineContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ZpoolStatusListener); ok {
		listenerT.ExitStatusLine(s)
	}
}

func (p *ZpoolStatusParser) StatusLine() (localctx IStatusLineContext) {
	localctx = NewStatusLineContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 18, ZpoolStatusParserRULE_statusLine)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(96)
		p.Match(ZpoolStatusParserSTATUS_KV)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(97)
		p.TextLine()
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

// IActionLineContext is an interface to support dynamic dispatch.
type IActionLineContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	ACTION_KV() antlr.TerminalNode
	TextLine() ITextLineContext

	// IsActionLineContext differentiates from other interfaces.
	IsActionLineContext()
}

type ActionLineContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyActionLineContext() *ActionLineContext {
	p := new(ActionLineContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ZpoolStatusParserRULE_actionLine
	return p
}

func InitEmptyActionLineContext(p *ActionLineContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ZpoolStatusParserRULE_actionLine
}

func (*ActionLineContext) IsActionLineContext() {}

func NewActionLineContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ActionLineContext {
	p := new(ActionLineContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = ZpoolStatusParserRULE_actionLine

	return p
}

func (s *ActionLineContext) GetParser() antlr.Parser { return s.parser }

func (s *ActionLineContext) ACTION_KV() antlr.TerminalNode {
	return s.GetToken(ZpoolStatusParserACTION_KV, 0)
}

func (s *ActionLineContext) TextLine() ITextLineContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ITextLineContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ITextLineContext)
}

func (s *ActionLineContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ActionLineContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *ActionLineContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ZpoolStatusListener); ok {
		listenerT.EnterActionLine(s)
	}
}

func (s *ActionLineContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ZpoolStatusListener); ok {
		listenerT.ExitActionLine(s)
	}
}

func (p *ZpoolStatusParser) ActionLine() (localctx IActionLineContext) {
	localctx = NewActionLineContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 20, ZpoolStatusParserRULE_actionLine)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(99)
		p.Match(ZpoolStatusParserACTION_KV)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(100)
		p.TextLine()
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

// ISeeLineContext is an interface to support dynamic dispatch.
type ISeeLineContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	SEE_KV() antlr.TerminalNode
	TextLine() ITextLineContext

	// IsSeeLineContext differentiates from other interfaces.
	IsSeeLineContext()
}

type SeeLineContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptySeeLineContext() *SeeLineContext {
	p := new(SeeLineContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ZpoolStatusParserRULE_seeLine
	return p
}

func InitEmptySeeLineContext(p *SeeLineContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ZpoolStatusParserRULE_seeLine
}

func (*SeeLineContext) IsSeeLineContext() {}

func NewSeeLineContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *SeeLineContext {
	p := new(SeeLineContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = ZpoolStatusParserRULE_seeLine

	return p
}

func (s *SeeLineContext) GetParser() antlr.Parser { return s.parser }

func (s *SeeLineContext) SEE_KV() antlr.TerminalNode {
	return s.GetToken(ZpoolStatusParserSEE_KV, 0)
}

func (s *SeeLineContext) TextLine() ITextLineContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ITextLineContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ITextLineContext)
}

func (s *SeeLineContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *SeeLineContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *SeeLineContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ZpoolStatusListener); ok {
		listenerT.EnterSeeLine(s)
	}
}

func (s *SeeLineContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ZpoolStatusListener); ok {
		listenerT.ExitSeeLine(s)
	}
}

func (p *ZpoolStatusParser) SeeLine() (localctx ISeeLineContext) {
	localctx = NewSeeLineContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 22, ZpoolStatusParserRULE_seeLine)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(102)
		p.Match(ZpoolStatusParserSEE_KV)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(103)
		p.TextLine()
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

// IConfigSectionContext is an interface to support dynamic dispatch.
type IConfigSectionContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	CONFIG_KV() antlr.TerminalNode
	AllNL() []antlr.TerminalNode
	NL(i int) antlr.TerminalNode
	ConfigHeaderLine() IConfigHeaderLineContext
	AllConfigRowLine() []IConfigRowLineContext
	ConfigRowLine(i int) IConfigRowLineContext

	// IsConfigSectionContext differentiates from other interfaces.
	IsConfigSectionContext()
}

type ConfigSectionContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyConfigSectionContext() *ConfigSectionContext {
	p := new(ConfigSectionContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ZpoolStatusParserRULE_configSection
	return p
}

func InitEmptyConfigSectionContext(p *ConfigSectionContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ZpoolStatusParserRULE_configSection
}

func (*ConfigSectionContext) IsConfigSectionContext() {}

func NewConfigSectionContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ConfigSectionContext {
	p := new(ConfigSectionContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = ZpoolStatusParserRULE_configSection

	return p
}

func (s *ConfigSectionContext) GetParser() antlr.Parser { return s.parser }

func (s *ConfigSectionContext) CONFIG_KV() antlr.TerminalNode {
	return s.GetToken(ZpoolStatusParserCONFIG_KV, 0)
}

func (s *ConfigSectionContext) AllNL() []antlr.TerminalNode {
	return s.GetTokens(ZpoolStatusParserNL)
}

func (s *ConfigSectionContext) NL(i int) antlr.TerminalNode {
	return s.GetToken(ZpoolStatusParserNL, i)
}

func (s *ConfigSectionContext) ConfigHeaderLine() IConfigHeaderLineContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IConfigHeaderLineContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IConfigHeaderLineContext)
}

func (s *ConfigSectionContext) AllConfigRowLine() []IConfigRowLineContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IConfigRowLineContext); ok {
			len++
		}
	}

	tst := make([]IConfigRowLineContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IConfigRowLineContext); ok {
			tst[i] = t.(IConfigRowLineContext)
			i++
		}
	}

	return tst
}

func (s *ConfigSectionContext) ConfigRowLine(i int) IConfigRowLineContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IConfigRowLineContext); ok {
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

	return t.(IConfigRowLineContext)
}

func (s *ConfigSectionContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ConfigSectionContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *ConfigSectionContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ZpoolStatusListener); ok {
		listenerT.EnterConfigSection(s)
	}
}

func (s *ConfigSectionContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ZpoolStatusListener); ok {
		listenerT.ExitConfigSection(s)
	}
}

func (p *ZpoolStatusParser) ConfigSection() (localctx IConfigSectionContext) {
	localctx = NewConfigSectionContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 24, ZpoolStatusParserRULE_configSection)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(105)
		p.Match(ZpoolStatusParserCONFIG_KV)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(106)
		p.Match(ZpoolStatusParserNL)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(110)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == ZpoolStatusParserNL {
		{
			p.SetState(107)
			p.Match(ZpoolStatusParserNL)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

		p.SetState(112)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(113)
		p.ConfigHeaderLine()
	}
	p.SetState(115)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for ok := true; ok; ok = _la == ZpoolStatusParserATOM {
		{
			p.SetState(114)
			p.ConfigRowLine()
		}

		p.SetState(117)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
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

// IConfigHeaderLineContext is an interface to support dynamic dispatch.
type IConfigHeaderLineContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	NL() antlr.TerminalNode
	AllHeaderAtom() []IHeaderAtomContext
	HeaderAtom(i int) IHeaderAtomContext

	// IsConfigHeaderLineContext differentiates from other interfaces.
	IsConfigHeaderLineContext()
}

type ConfigHeaderLineContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyConfigHeaderLineContext() *ConfigHeaderLineContext {
	p := new(ConfigHeaderLineContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ZpoolStatusParserRULE_configHeaderLine
	return p
}

func InitEmptyConfigHeaderLineContext(p *ConfigHeaderLineContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ZpoolStatusParserRULE_configHeaderLine
}

func (*ConfigHeaderLineContext) IsConfigHeaderLineContext() {}

func NewConfigHeaderLineContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ConfigHeaderLineContext {
	p := new(ConfigHeaderLineContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = ZpoolStatusParserRULE_configHeaderLine

	return p
}

func (s *ConfigHeaderLineContext) GetParser() antlr.Parser { return s.parser }

func (s *ConfigHeaderLineContext) NL() antlr.TerminalNode {
	return s.GetToken(ZpoolStatusParserNL, 0)
}

func (s *ConfigHeaderLineContext) AllHeaderAtom() []IHeaderAtomContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IHeaderAtomContext); ok {
			len++
		}
	}

	tst := make([]IHeaderAtomContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IHeaderAtomContext); ok {
			tst[i] = t.(IHeaderAtomContext)
			i++
		}
	}

	return tst
}

func (s *ConfigHeaderLineContext) HeaderAtom(i int) IHeaderAtomContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IHeaderAtomContext); ok {
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

	return t.(IHeaderAtomContext)
}

func (s *ConfigHeaderLineContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ConfigHeaderLineContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *ConfigHeaderLineContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ZpoolStatusListener); ok {
		listenerT.EnterConfigHeaderLine(s)
	}
}

func (s *ConfigHeaderLineContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ZpoolStatusListener); ok {
		listenerT.ExitConfigHeaderLine(s)
	}
}

func (p *ZpoolStatusParser) ConfigHeaderLine() (localctx IConfigHeaderLineContext) {
	localctx = NewConfigHeaderLineContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 26, ZpoolStatusParserRULE_configHeaderLine)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(120)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for ok := true; ok; ok = _la == ZpoolStatusParserATOM {
		{
			p.SetState(119)
			p.HeaderAtom()
		}

		p.SetState(122)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(124)
		p.Match(ZpoolStatusParserNL)
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

// IConfigRowLineContext is an interface to support dynamic dispatch.
type IConfigRowLineContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	NL() antlr.TerminalNode
	AllRowAtom() []IRowAtomContext
	RowAtom(i int) IRowAtomContext

	// IsConfigRowLineContext differentiates from other interfaces.
	IsConfigRowLineContext()
}

type ConfigRowLineContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyConfigRowLineContext() *ConfigRowLineContext {
	p := new(ConfigRowLineContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ZpoolStatusParserRULE_configRowLine
	return p
}

func InitEmptyConfigRowLineContext(p *ConfigRowLineContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ZpoolStatusParserRULE_configRowLine
}

func (*ConfigRowLineContext) IsConfigRowLineContext() {}

func NewConfigRowLineContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ConfigRowLineContext {
	p := new(ConfigRowLineContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = ZpoolStatusParserRULE_configRowLine

	return p
}

func (s *ConfigRowLineContext) GetParser() antlr.Parser { return s.parser }

func (s *ConfigRowLineContext) NL() antlr.TerminalNode {
	return s.GetToken(ZpoolStatusParserNL, 0)
}

func (s *ConfigRowLineContext) AllRowAtom() []IRowAtomContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IRowAtomContext); ok {
			len++
		}
	}

	tst := make([]IRowAtomContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IRowAtomContext); ok {
			tst[i] = t.(IRowAtomContext)
			i++
		}
	}

	return tst
}

func (s *ConfigRowLineContext) RowAtom(i int) IRowAtomContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IRowAtomContext); ok {
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

	return t.(IRowAtomContext)
}

func (s *ConfigRowLineContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ConfigRowLineContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *ConfigRowLineContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ZpoolStatusListener); ok {
		listenerT.EnterConfigRowLine(s)
	}
}

func (s *ConfigRowLineContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ZpoolStatusListener); ok {
		listenerT.ExitConfigRowLine(s)
	}
}

func (p *ZpoolStatusParser) ConfigRowLine() (localctx IConfigRowLineContext) {
	localctx = NewConfigRowLineContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 28, ZpoolStatusParserRULE_configRowLine)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(127)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for ok := true; ok; ok = _la == ZpoolStatusParserATOM {
		{
			p.SetState(126)
			p.RowAtom()
		}

		p.SetState(129)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(131)
		p.Match(ZpoolStatusParserNL)
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

// IErrorsLineContext is an interface to support dynamic dispatch.
type IErrorsLineContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	ERRORS_KV() antlr.TerminalNode
	TextLine() ITextLineContext

	// IsErrorsLineContext differentiates from other interfaces.
	IsErrorsLineContext()
}

type ErrorsLineContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyErrorsLineContext() *ErrorsLineContext {
	p := new(ErrorsLineContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ZpoolStatusParserRULE_errorsLine
	return p
}

func InitEmptyErrorsLineContext(p *ErrorsLineContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ZpoolStatusParserRULE_errorsLine
}

func (*ErrorsLineContext) IsErrorsLineContext() {}

func NewErrorsLineContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ErrorsLineContext {
	p := new(ErrorsLineContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = ZpoolStatusParserRULE_errorsLine

	return p
}

func (s *ErrorsLineContext) GetParser() antlr.Parser { return s.parser }

func (s *ErrorsLineContext) ERRORS_KV() antlr.TerminalNode {
	return s.GetToken(ZpoolStatusParserERRORS_KV, 0)
}

func (s *ErrorsLineContext) TextLine() ITextLineContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ITextLineContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ITextLineContext)
}

func (s *ErrorsLineContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ErrorsLineContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *ErrorsLineContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ZpoolStatusListener); ok {
		listenerT.EnterErrorsLine(s)
	}
}

func (s *ErrorsLineContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ZpoolStatusListener); ok {
		listenerT.ExitErrorsLine(s)
	}
}

func (p *ZpoolStatusParser) ErrorsLine() (localctx IErrorsLineContext) {
	localctx = NewErrorsLineContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 30, ZpoolStatusParserRULE_errorsLine)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(133)
		p.Match(ZpoolStatusParserERRORS_KV)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(134)
		p.TextLine()
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

// ITextLineContext is an interface to support dynamic dispatch.
type ITextLineContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	NL() antlr.TerminalNode
	AllTextAtom() []ITextAtomContext
	TextAtom(i int) ITextAtomContext

	// IsTextLineContext differentiates from other interfaces.
	IsTextLineContext()
}

type TextLineContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyTextLineContext() *TextLineContext {
	p := new(TextLineContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ZpoolStatusParserRULE_textLine
	return p
}

func InitEmptyTextLineContext(p *TextLineContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ZpoolStatusParserRULE_textLine
}

func (*TextLineContext) IsTextLineContext() {}

func NewTextLineContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *TextLineContext {
	p := new(TextLineContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = ZpoolStatusParserRULE_textLine

	return p
}

func (s *TextLineContext) GetParser() antlr.Parser { return s.parser }

func (s *TextLineContext) NL() antlr.TerminalNode {
	return s.GetToken(ZpoolStatusParserNL, 0)
}

func (s *TextLineContext) AllTextAtom() []ITextAtomContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(ITextAtomContext); ok {
			len++
		}
	}

	tst := make([]ITextAtomContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(ITextAtomContext); ok {
			tst[i] = t.(ITextAtomContext)
			i++
		}
	}

	return tst
}

func (s *TextLineContext) TextAtom(i int) ITextAtomContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ITextAtomContext); ok {
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

	return t.(ITextAtomContext)
}

func (s *TextLineContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *TextLineContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *TextLineContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ZpoolStatusListener); ok {
		listenerT.EnterTextLine(s)
	}
}

func (s *TextLineContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ZpoolStatusListener); ok {
		listenerT.ExitTextLine(s)
	}
}

func (p *ZpoolStatusParser) TextLine() (localctx ITextLineContext) {
	localctx = NewTextLineContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 32, ZpoolStatusParserRULE_textLine)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(139)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == ZpoolStatusParserATOM {
		{
			p.SetState(136)
			p.TextAtom()
		}

		p.SetState(141)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(142)
		p.Match(ZpoolStatusParserNL)
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

// ITextAtomContext is an interface to support dynamic dispatch.
type ITextAtomContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	ATOM() antlr.TerminalNode

	// IsTextAtomContext differentiates from other interfaces.
	IsTextAtomContext()
}

type TextAtomContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyTextAtomContext() *TextAtomContext {
	p := new(TextAtomContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ZpoolStatusParserRULE_textAtom
	return p
}

func InitEmptyTextAtomContext(p *TextAtomContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ZpoolStatusParserRULE_textAtom
}

func (*TextAtomContext) IsTextAtomContext() {}

func NewTextAtomContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *TextAtomContext {
	p := new(TextAtomContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = ZpoolStatusParserRULE_textAtom

	return p
}

func (s *TextAtomContext) GetParser() antlr.Parser { return s.parser }

func (s *TextAtomContext) ATOM() antlr.TerminalNode {
	return s.GetToken(ZpoolStatusParserATOM, 0)
}

func (s *TextAtomContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *TextAtomContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *TextAtomContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ZpoolStatusListener); ok {
		listenerT.EnterTextAtom(s)
	}
}

func (s *TextAtomContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ZpoolStatusListener); ok {
		listenerT.ExitTextAtom(s)
	}
}

func (p *ZpoolStatusParser) TextAtom() (localctx ITextAtomContext) {
	localctx = NewTextAtomContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 34, ZpoolStatusParserRULE_textAtom)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(144)
		p.Match(ZpoolStatusParserATOM)
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

// IHeaderAtomContext is an interface to support dynamic dispatch.
type IHeaderAtomContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	ATOM() antlr.TerminalNode

	// IsHeaderAtomContext differentiates from other interfaces.
	IsHeaderAtomContext()
}

type HeaderAtomContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyHeaderAtomContext() *HeaderAtomContext {
	p := new(HeaderAtomContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ZpoolStatusParserRULE_headerAtom
	return p
}

func InitEmptyHeaderAtomContext(p *HeaderAtomContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ZpoolStatusParserRULE_headerAtom
}

func (*HeaderAtomContext) IsHeaderAtomContext() {}

func NewHeaderAtomContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *HeaderAtomContext {
	p := new(HeaderAtomContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = ZpoolStatusParserRULE_headerAtom

	return p
}

func (s *HeaderAtomContext) GetParser() antlr.Parser { return s.parser }

func (s *HeaderAtomContext) ATOM() antlr.TerminalNode {
	return s.GetToken(ZpoolStatusParserATOM, 0)
}

func (s *HeaderAtomContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *HeaderAtomContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *HeaderAtomContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ZpoolStatusListener); ok {
		listenerT.EnterHeaderAtom(s)
	}
}

func (s *HeaderAtomContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ZpoolStatusListener); ok {
		listenerT.ExitHeaderAtom(s)
	}
}

func (p *ZpoolStatusParser) HeaderAtom() (localctx IHeaderAtomContext) {
	localctx = NewHeaderAtomContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 36, ZpoolStatusParserRULE_headerAtom)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(146)
		p.Match(ZpoolStatusParserATOM)
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

// IRowAtomContext is an interface to support dynamic dispatch.
type IRowAtomContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	ATOM() antlr.TerminalNode

	// IsRowAtomContext differentiates from other interfaces.
	IsRowAtomContext()
}

type RowAtomContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyRowAtomContext() *RowAtomContext {
	p := new(RowAtomContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ZpoolStatusParserRULE_rowAtom
	return p
}

func InitEmptyRowAtomContext(p *RowAtomContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ZpoolStatusParserRULE_rowAtom
}

func (*RowAtomContext) IsRowAtomContext() {}

func NewRowAtomContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *RowAtomContext {
	p := new(RowAtomContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = ZpoolStatusParserRULE_rowAtom

	return p
}

func (s *RowAtomContext) GetParser() antlr.Parser { return s.parser }

func (s *RowAtomContext) ATOM() antlr.TerminalNode {
	return s.GetToken(ZpoolStatusParserATOM, 0)
}

func (s *RowAtomContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *RowAtomContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *RowAtomContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ZpoolStatusListener); ok {
		listenerT.EnterRowAtom(s)
	}
}

func (s *RowAtomContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ZpoolStatusListener); ok {
		listenerT.ExitRowAtom(s)
	}
}

func (p *ZpoolStatusParser) RowAtom() (localctx IRowAtomContext) {
	localctx = NewRowAtomContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 38, ZpoolStatusParserRULE_rowAtom)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(148)
		p.Match(ZpoolStatusParserATOM)
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
