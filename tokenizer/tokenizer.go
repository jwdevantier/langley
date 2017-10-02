package tokenizer

// Lexer => tokenizer
// Item => Token
// ItemType => TokenType
//"github.com/golang-collections/collections/stack"

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

const eof = -1
const newline = '\n'

// Token represents a token/string returned from the tokenizer
type Token struct {
	typ  TokenType // general type/category of token
	val  string    // literal value
	pos  Pos       // starting position, in bytes, of this token in input string
	line Pos       // the line number at the start of this token
}

//Pos - a positive int
type Pos int

type StateFn func(*Tokenizer) StateFn

// Tokenizer contains the state of the tokenizer instance
type Tokenizer struct {
	name  string  // Used for error reporting
	input string  // String being scanned
	state StateFn // the next state function to enter

	start  Pos //start position of this item
	end    Pos //end position of this item
	pos    Pos //current position into input
	oldPos Pos // position, in bytes, of most recent token returned

	width  Pos        //width of last rune read
	tokens chan Token //channel of scanned items
	line   Pos

	//state StateFn
}

// New creates a new Tokenizer
func New(name string, initialState StateFn, input string) *Tokenizer {
	t := &Tokenizer{
		name:   name,
		input:  input,
		tokens: make(chan Token),
		state:  initialState,
	}

	go t.run()
	return t
}

// NextRune returns the next rune (UTF-8 character) in the input
func (t *Tokenizer) NextRune() rune {
	if t.pos > Pos(len(t.input)) {
		t.width = 0
		return eof
	}
	r, w :=
		utf8.DecodeRuneInString(t.input[t.pos:])
	t.width = Pos(w)
	t.pos += t.width
	if r == newline {
		t.line++
	}
	return r

}

// Peek returns the next rune of the input stream without consuming it
func (t *Tokenizer) Peek() rune {
	r := t.NextRune()
	t.Backup()
	return r
}

// Backup steps the tokenizer back one run in the input stream
func (t *Tokenizer) Backup() {
	t.pos -= t.width
	if t.width == 1 && t.input[t.pos] == newline {
		t.line--
	}
}

// Emit extracts the string between the last token end and the current
// position in the input stream and emits it as a token on the token
// channel of type 'tt'.
func (t *Tokenizer) Emit(tt TokenType) {
	t.tokens <- Token{tt, t.input[t.start:t.pos], t.start, t.line}
	if tt.CountNewLines() {
		t.line += Pos(strings.Count(t.input[t.start:t.pos], "\n"))
	}
	t.start = t.pos // end of token becomes new starting pos
}

// Ignore skips over the pending input before this point
func (t *Tokenizer) Ignore() {
	t.line += Pos(strings.Count(t.input[t.start:t.pos], "\n"))
	t.start = t.pos
}

// MatchOne matches next rune iff contained in alphabet
// returns a bool, true iff next rune is contained in the alphabet
func (t *Tokenizer) MatchOne(alphabet string) bool {
	if strings.ContainsRune(alphabet, t.NextRune()) {
		return true
	}
	t.Backup()
	return false
}

// MatchMany greedily matches runes contained in the alphabet
func (t *Tokenizer) MatchMany(alphabet string) {
	for strings.ContainsRune(alphabet, t.NextRune()) {
	}
	t.Backup()
}

// Errorf returns an error token and terminates the tokenizer
// by returning 'nil' as the next StateFn, terminating t.Next()
func (t *Tokenizer) Errorf(format string, args ...interface{}) StateFn {
	t.tokens <- Token{
		Error,
		fmt.Sprintf(format, args...),
		t.start,
		t.line,
	}
	return nil
}

// Next returns the next token in the token stream
func (t *Tokenizer) Next() Token {
	tkn := <-t.tokens
	t.oldPos = t.pos
	return tkn
}

// Drain consumes all tokens waiting in the channel so
// the tokenization goroutine will exit.
// To be called by the parser, not the tokenizing goroutine.
func (t *Tokenizer) Drain() {
	for range t.tokens {
	}
}

// run executes each lex/tokenization state in turn
// until no more states are returned indicating an end
// of the underlying input stream or an error.
// (called during initialization of the Tokenizer)
func (t *Tokenizer) run() {
	for t.state != nil {
		t.state = t.state(t)
	}
	close(t.tokens) //no more tokens delivered
}
