// Code generated from /home/oleksiybondar/Documents/development/lite-nas/shared/go/parsers/zfs/status/grammar/ZpoolStatus.g4 by ANTLR 4.13.2. DO NOT EDIT.

package zpoolstatus

import (
	"fmt"
	"sync"
	"unicode"

	"github.com/antlr4-go/antlr/v4"
)

// Suppress unused import error
var (
	_ = fmt.Printf
	_ = sync.Once{}
	_ = unicode.IsLetter
)

type ZpoolStatusLexer struct {
	*antlr.BaseLexer
	channelNames []string
	modeNames    []string
	// TODO: EOF string
}

var ZpoolStatusLexerLexerStaticData struct {
	once                   sync.Once
	serializedATN          []int32
	ChannelNames           []string
	ModeNames              []string
	LiteralNames           []string
	SymbolicNames          []string
	RuleNames              []string
	PredictionContextCache *antlr.PredictionContextCache
	atn                    *antlr.ATN
	decisionToDFA          []*antlr.DFA
}

func zpoolstatuslexerLexerInit() {
	staticData := &ZpoolStatusLexerLexerStaticData
	staticData.ChannelNames = []string{
		"DEFAULT_TOKEN_CHANNEL", "HIDDEN",
	}
	staticData.ModeNames = []string{
		"DEFAULT_MODE",
	}
	staticData.LiteralNames = []string{
		"", "'pool:'", "'state:'", "'scan:'", "'config:'", "'errors:'", "'status:'",
		"'action:'", "'see:'",
	}
	staticData.SymbolicNames = []string{
		"", "POOL_KV", "STATE_KV", "SCAN_KV", "CONFIG_KV", "ERRORS_KV", "STATUS_KV",
		"ACTION_KV", "SEE_KV", "ATOM", "WS", "NL",
	}
	staticData.RuleNames = []string{
		"POOL_KV", "STATE_KV", "SCAN_KV", "CONFIG_KV", "ERRORS_KV", "STATUS_KV",
		"ACTION_KV", "SEE_KV", "ATOM", "WS", "NL",
	}
	staticData.PredictionContextCache = antlr.NewPredictionContextCache()
	staticData.serializedATN = []int32{
		4, 0, 11, 96, 6, -1, 2, 0, 7, 0, 2, 1, 7, 1, 2, 2, 7, 2, 2, 3, 7, 3, 2,
		4, 7, 4, 2, 5, 7, 5, 2, 6, 7, 6, 2, 7, 7, 7, 2, 8, 7, 8, 2, 9, 7, 9, 2,
		10, 7, 10, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 1, 1, 1, 1, 1, 1, 1,
		1, 1, 1, 1, 1, 1, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 3, 1, 3, 1, 3,
		1, 3, 1, 3, 1, 3, 1, 3, 1, 3, 1, 4, 1, 4, 1, 4, 1, 4, 1, 4, 1, 4, 1, 4,
		1, 4, 1, 5, 1, 5, 1, 5, 1, 5, 1, 5, 1, 5, 1, 5, 1, 5, 1, 6, 1, 6, 1, 6,
		1, 6, 1, 6, 1, 6, 1, 6, 1, 6, 1, 7, 1, 7, 1, 7, 1, 7, 1, 7, 1, 8, 4, 8,
		81, 8, 8, 11, 8, 12, 8, 82, 1, 9, 4, 9, 86, 8, 9, 11, 9, 12, 9, 87, 1,
		9, 1, 9, 1, 10, 3, 10, 93, 8, 10, 1, 10, 1, 10, 0, 0, 11, 1, 1, 3, 2, 5,
		3, 7, 4, 9, 5, 11, 6, 13, 7, 15, 8, 17, 9, 19, 10, 21, 11, 1, 0, 2, 3,
		0, 9, 10, 13, 13, 32, 32, 2, 0, 9, 9, 32, 32, 98, 0, 1, 1, 0, 0, 0, 0,
		3, 1, 0, 0, 0, 0, 5, 1, 0, 0, 0, 0, 7, 1, 0, 0, 0, 0, 9, 1, 0, 0, 0, 0,
		11, 1, 0, 0, 0, 0, 13, 1, 0, 0, 0, 0, 15, 1, 0, 0, 0, 0, 17, 1, 0, 0, 0,
		0, 19, 1, 0, 0, 0, 0, 21, 1, 0, 0, 0, 1, 23, 1, 0, 0, 0, 3, 29, 1, 0, 0,
		0, 5, 36, 1, 0, 0, 0, 7, 42, 1, 0, 0, 0, 9, 50, 1, 0, 0, 0, 11, 58, 1,
		0, 0, 0, 13, 66, 1, 0, 0, 0, 15, 74, 1, 0, 0, 0, 17, 80, 1, 0, 0, 0, 19,
		85, 1, 0, 0, 0, 21, 92, 1, 0, 0, 0, 23, 24, 5, 112, 0, 0, 24, 25, 5, 111,
		0, 0, 25, 26, 5, 111, 0, 0, 26, 27, 5, 108, 0, 0, 27, 28, 5, 58, 0, 0,
		28, 2, 1, 0, 0, 0, 29, 30, 5, 115, 0, 0, 30, 31, 5, 116, 0, 0, 31, 32,
		5, 97, 0, 0, 32, 33, 5, 116, 0, 0, 33, 34, 5, 101, 0, 0, 34, 35, 5, 58,
		0, 0, 35, 4, 1, 0, 0, 0, 36, 37, 5, 115, 0, 0, 37, 38, 5, 99, 0, 0, 38,
		39, 5, 97, 0, 0, 39, 40, 5, 110, 0, 0, 40, 41, 5, 58, 0, 0, 41, 6, 1, 0,
		0, 0, 42, 43, 5, 99, 0, 0, 43, 44, 5, 111, 0, 0, 44, 45, 5, 110, 0, 0,
		45, 46, 5, 102, 0, 0, 46, 47, 5, 105, 0, 0, 47, 48, 5, 103, 0, 0, 48, 49,
		5, 58, 0, 0, 49, 8, 1, 0, 0, 0, 50, 51, 5, 101, 0, 0, 51, 52, 5, 114, 0,
		0, 52, 53, 5, 114, 0, 0, 53, 54, 5, 111, 0, 0, 54, 55, 5, 114, 0, 0, 55,
		56, 5, 115, 0, 0, 56, 57, 5, 58, 0, 0, 57, 10, 1, 0, 0, 0, 58, 59, 5, 115,
		0, 0, 59, 60, 5, 116, 0, 0, 60, 61, 5, 97, 0, 0, 61, 62, 5, 116, 0, 0,
		62, 63, 5, 117, 0, 0, 63, 64, 5, 115, 0, 0, 64, 65, 5, 58, 0, 0, 65, 12,
		1, 0, 0, 0, 66, 67, 5, 97, 0, 0, 67, 68, 5, 99, 0, 0, 68, 69, 5, 116, 0,
		0, 69, 70, 5, 105, 0, 0, 70, 71, 5, 111, 0, 0, 71, 72, 5, 110, 0, 0, 72,
		73, 5, 58, 0, 0, 73, 14, 1, 0, 0, 0, 74, 75, 5, 115, 0, 0, 75, 76, 5, 101,
		0, 0, 76, 77, 5, 101, 0, 0, 77, 78, 5, 58, 0, 0, 78, 16, 1, 0, 0, 0, 79,
		81, 8, 0, 0, 0, 80, 79, 1, 0, 0, 0, 81, 82, 1, 0, 0, 0, 82, 80, 1, 0, 0,
		0, 82, 83, 1, 0, 0, 0, 83, 18, 1, 0, 0, 0, 84, 86, 7, 1, 0, 0, 85, 84,
		1, 0, 0, 0, 86, 87, 1, 0, 0, 0, 87, 85, 1, 0, 0, 0, 87, 88, 1, 0, 0, 0,
		88, 89, 1, 0, 0, 0, 89, 90, 6, 9, 0, 0, 90, 20, 1, 0, 0, 0, 91, 93, 5,
		13, 0, 0, 92, 91, 1, 0, 0, 0, 92, 93, 1, 0, 0, 0, 93, 94, 1, 0, 0, 0, 94,
		95, 5, 10, 0, 0, 95, 22, 1, 0, 0, 0, 4, 0, 82, 87, 92, 1, 6, 0, 0,
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

// ZpoolStatusLexerInit initializes any static state used to implement ZpoolStatusLexer. By default the
// static state used to implement the lexer is lazily initialized during the first call to
// NewZpoolStatusLexer(). You can call this function if you wish to initialize the static state ahead
// of time.
func ZpoolStatusLexerInit() {
	staticData := &ZpoolStatusLexerLexerStaticData
	staticData.once.Do(zpoolstatuslexerLexerInit)
}

// NewZpoolStatusLexer produces a new lexer instance for the optional input antlr.CharStream.
func NewZpoolStatusLexer(input antlr.CharStream) *ZpoolStatusLexer {
	ZpoolStatusLexerInit()
	l := new(ZpoolStatusLexer)
	l.BaseLexer = antlr.NewBaseLexer(input)
	staticData := &ZpoolStatusLexerLexerStaticData
	l.Interpreter = antlr.NewLexerATNSimulator(l, staticData.atn, staticData.decisionToDFA, staticData.PredictionContextCache)
	l.channelNames = staticData.ChannelNames
	l.modeNames = staticData.ModeNames
	l.RuleNames = staticData.RuleNames
	l.LiteralNames = staticData.LiteralNames
	l.SymbolicNames = staticData.SymbolicNames
	l.GrammarFileName = "ZpoolStatus.g4"
	// TODO: l.EOF = antlr.TokenEOF

	return l
}

// ZpoolStatusLexer tokens.
const (
	ZpoolStatusLexerPOOL_KV   = 1
	ZpoolStatusLexerSTATE_KV  = 2
	ZpoolStatusLexerSCAN_KV   = 3
	ZpoolStatusLexerCONFIG_KV = 4
	ZpoolStatusLexerERRORS_KV = 5
	ZpoolStatusLexerSTATUS_KV = 6
	ZpoolStatusLexerACTION_KV = 7
	ZpoolStatusLexerSEE_KV    = 8
	ZpoolStatusLexerATOM      = 9
	ZpoolStatusLexerWS        = 10
	ZpoolStatusLexerNL        = 11
)
