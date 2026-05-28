// Code generated from /home/oleksiybondar/Documents/development/lite-nas/shared/go/parsers/acl/getfacl/grammar/Getfacl.g4 by ANTLR 4.13.2. DO NOT EDIT.

package getfacl

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

type GetfaclLexer struct {
	*antlr.BaseLexer
	channelNames []string
	modeNames    []string
	// TODO: EOF string
}

var GetfaclLexerLexerStaticData struct {
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

func getfacllexerLexerInit() {
	staticData := &GetfaclLexerLexerStaticData
	staticData.ChannelNames = []string{
		"DEFAULT_TOKEN_CHANNEL", "HIDDEN",
	}
	staticData.ModeNames = []string{
		"DEFAULT_MODE",
	}
	staticData.LiteralNames = []string{
		"", "'file'", "'owner'", "'default:'", "'user'", "'group'", "'other'",
		"'mask'", "", "", "'#'", "':'",
	}
	staticData.SymbolicNames = []string{
		"", "", "", "DEFAULT_PREFIX", "USER_TAG", "GROUP_TAG", "OTHER_TAG",
		"MASK_TAG", "PERM", "VALUE_ATOM", "HASH", "COLON", "WS", "NL",
	}
	staticData.RuleNames = []string{
		"T__0", "T__1", "DEFAULT_PREFIX", "USER_TAG", "GROUP_TAG", "OTHER_TAG",
		"MASK_TAG", "PERM", "VALUE_ATOM", "HASH", "COLON", "WS", "NL",
	}
	staticData.PredictionContextCache = antlr.NewPredictionContextCache()
	staticData.serializedATN = []int32{
		4, 0, 13, 94, 6, -1, 2, 0, 7, 0, 2, 1, 7, 1, 2, 2, 7, 2, 2, 3, 7, 3, 2,
		4, 7, 4, 2, 5, 7, 5, 2, 6, 7, 6, 2, 7, 7, 7, 2, 8, 7, 8, 2, 9, 7, 9, 2,
		10, 7, 10, 2, 11, 7, 11, 2, 12, 7, 12, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1,
		1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1,
		2, 1, 2, 1, 2, 1, 3, 1, 3, 1, 3, 1, 3, 1, 3, 1, 4, 1, 4, 1, 4, 1, 4, 1,
		4, 1, 4, 1, 5, 1, 5, 1, 5, 1, 5, 1, 5, 1, 5, 1, 6, 1, 6, 1, 6, 1, 6, 1,
		6, 1, 7, 1, 7, 1, 7, 1, 7, 1, 8, 4, 8, 75, 8, 8, 11, 8, 12, 8, 76, 1, 9,
		1, 9, 1, 10, 1, 10, 1, 11, 4, 11, 84, 8, 11, 11, 11, 12, 11, 85, 1, 11,
		1, 11, 1, 12, 3, 12, 91, 8, 12, 1, 12, 1, 12, 0, 0, 13, 1, 1, 3, 2, 5,
		3, 7, 4, 9, 5, 11, 6, 13, 7, 15, 8, 17, 9, 19, 10, 21, 11, 23, 12, 25,
		13, 1, 0, 5, 2, 0, 45, 45, 114, 114, 2, 0, 45, 45, 119, 119, 2, 0, 45,
		45, 120, 120, 5, 0, 9, 10, 13, 13, 32, 32, 35, 35, 58, 58, 2, 0, 9, 9,
		32, 32, 96, 0, 1, 1, 0, 0, 0, 0, 3, 1, 0, 0, 0, 0, 5, 1, 0, 0, 0, 0, 7,
		1, 0, 0, 0, 0, 9, 1, 0, 0, 0, 0, 11, 1, 0, 0, 0, 0, 13, 1, 0, 0, 0, 0,
		15, 1, 0, 0, 0, 0, 17, 1, 0, 0, 0, 0, 19, 1, 0, 0, 0, 0, 21, 1, 0, 0, 0,
		0, 23, 1, 0, 0, 0, 0, 25, 1, 0, 0, 0, 1, 27, 1, 0, 0, 0, 3, 32, 1, 0, 0,
		0, 5, 38, 1, 0, 0, 0, 7, 47, 1, 0, 0, 0, 9, 52, 1, 0, 0, 0, 11, 58, 1,
		0, 0, 0, 13, 64, 1, 0, 0, 0, 15, 69, 1, 0, 0, 0, 17, 74, 1, 0, 0, 0, 19,
		78, 1, 0, 0, 0, 21, 80, 1, 0, 0, 0, 23, 83, 1, 0, 0, 0, 25, 90, 1, 0, 0,
		0, 27, 28, 5, 102, 0, 0, 28, 29, 5, 105, 0, 0, 29, 30, 5, 108, 0, 0, 30,
		31, 5, 101, 0, 0, 31, 2, 1, 0, 0, 0, 32, 33, 5, 111, 0, 0, 33, 34, 5, 119,
		0, 0, 34, 35, 5, 110, 0, 0, 35, 36, 5, 101, 0, 0, 36, 37, 5, 114, 0, 0,
		37, 4, 1, 0, 0, 0, 38, 39, 5, 100, 0, 0, 39, 40, 5, 101, 0, 0, 40, 41,
		5, 102, 0, 0, 41, 42, 5, 97, 0, 0, 42, 43, 5, 117, 0, 0, 43, 44, 5, 108,
		0, 0, 44, 45, 5, 116, 0, 0, 45, 46, 5, 58, 0, 0, 46, 6, 1, 0, 0, 0, 47,
		48, 5, 117, 0, 0, 48, 49, 5, 115, 0, 0, 49, 50, 5, 101, 0, 0, 50, 51, 5,
		114, 0, 0, 51, 8, 1, 0, 0, 0, 52, 53, 5, 103, 0, 0, 53, 54, 5, 114, 0,
		0, 54, 55, 5, 111, 0, 0, 55, 56, 5, 117, 0, 0, 56, 57, 5, 112, 0, 0, 57,
		10, 1, 0, 0, 0, 58, 59, 5, 111, 0, 0, 59, 60, 5, 116, 0, 0, 60, 61, 5,
		104, 0, 0, 61, 62, 5, 101, 0, 0, 62, 63, 5, 114, 0, 0, 63, 12, 1, 0, 0,
		0, 64, 65, 5, 109, 0, 0, 65, 66, 5, 97, 0, 0, 66, 67, 5, 115, 0, 0, 67,
		68, 5, 107, 0, 0, 68, 14, 1, 0, 0, 0, 69, 70, 7, 0, 0, 0, 70, 71, 7, 1,
		0, 0, 71, 72, 7, 2, 0, 0, 72, 16, 1, 0, 0, 0, 73, 75, 8, 3, 0, 0, 74, 73,
		1, 0, 0, 0, 75, 76, 1, 0, 0, 0, 76, 74, 1, 0, 0, 0, 76, 77, 1, 0, 0, 0,
		77, 18, 1, 0, 0, 0, 78, 79, 5, 35, 0, 0, 79, 20, 1, 0, 0, 0, 80, 81, 5,
		58, 0, 0, 81, 22, 1, 0, 0, 0, 82, 84, 7, 4, 0, 0, 83, 82, 1, 0, 0, 0, 84,
		85, 1, 0, 0, 0, 85, 83, 1, 0, 0, 0, 85, 86, 1, 0, 0, 0, 86, 87, 1, 0, 0,
		0, 87, 88, 6, 11, 0, 0, 88, 24, 1, 0, 0, 0, 89, 91, 5, 13, 0, 0, 90, 89,
		1, 0, 0, 0, 90, 91, 1, 0, 0, 0, 91, 92, 1, 0, 0, 0, 92, 93, 5, 10, 0, 0,
		93, 26, 1, 0, 0, 0, 4, 0, 76, 85, 90, 1, 6, 0, 0,
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

// GetfaclLexerInit initializes any static state used to implement GetfaclLexer. By default the
// static state used to implement the lexer is lazily initialized during the first call to
// NewGetfaclLexer(). You can call this function if you wish to initialize the static state ahead
// of time.
func GetfaclLexerInit() {
	staticData := &GetfaclLexerLexerStaticData
	staticData.once.Do(getfacllexerLexerInit)
}

// NewGetfaclLexer produces a new lexer instance for the optional input antlr.CharStream.
func NewGetfaclLexer(input antlr.CharStream) *GetfaclLexer {
	GetfaclLexerInit()
	l := new(GetfaclLexer)
	l.BaseLexer = antlr.NewBaseLexer(input)
	staticData := &GetfaclLexerLexerStaticData
	l.Interpreter = antlr.NewLexerATNSimulator(l, staticData.atn, staticData.decisionToDFA, staticData.PredictionContextCache)
	l.channelNames = staticData.ChannelNames
	l.modeNames = staticData.ModeNames
	l.RuleNames = staticData.RuleNames
	l.LiteralNames = staticData.LiteralNames
	l.SymbolicNames = staticData.SymbolicNames
	l.GrammarFileName = "Getfacl.g4"
	// TODO: l.EOF = antlr.TokenEOF

	return l
}

// GetfaclLexer tokens.
const (
	GetfaclLexerT__0           = 1
	GetfaclLexerT__1           = 2
	GetfaclLexerDEFAULT_PREFIX = 3
	GetfaclLexerUSER_TAG       = 4
	GetfaclLexerGROUP_TAG      = 5
	GetfaclLexerOTHER_TAG      = 6
	GetfaclLexerMASK_TAG       = 7
	GetfaclLexerPERM           = 8
	GetfaclLexerVALUE_ATOM     = 9
	GetfaclLexerHASH           = 10
	GetfaclLexerCOLON          = 11
	GetfaclLexerWS             = 12
	GetfaclLexerNL             = 13
)
