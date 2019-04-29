package reader

import (
	"fmt"
	"io"
	"strings"

	log "github.com/sirupsen/logrus"
)

// stateFn state function type
type stateFn func(*LispReader) stateFn

/////////////////////////////////////////////////////////////////////////////
///   Object & Constructor   ////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////

// LispReader embeds the PositionReader struct
type LispReader struct {
	name  string
	atoms []Atom
	token strings.Builder
	*PositionReader
}

// NewLispReader creates a LispReader for the given string and optional
//                   position stack
func NewLispReader(programName string, programData string) *LispReader {
	var token strings.Builder
	return &LispReader{
		programName,
		[]Atom{},
		token,
		NewPositionReader(programData, initPosition()),
	}
}

/////////////////////////////////////////////////////////////////////////////
///   Methods   /////////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////

// Read ...
func (l *LispReader) Read() {
	var rn rune
	for rn = l.read1(); isWhitespace(rn); {
		continue
	}

	if rn == EOF {
		readEndOfFile(l)
		return
	}
	l.WriteToken(rn)

	switch {
	case rn == leftParen:
		readLeftParen(l)
	case rn == rightParen:
		readRightParen(l)
	case rn == leftBracket:
		readLeftVect(l)
	case rn == rightBracket:
		readRightVect(l)
	case rn == doubleQuote:
		readString(l)
	// // case r == semiColon:
	// // 	return readComment
	// // case isCompoundNumber(r):
	// // 	return readNumber
	// // case isAllowedIdentifierRune(r):
	// // 	return readIdentifier
	default:
		log.Panic(fmt.Sprintf(UnsupportedRuneError, rn, string(rn)))
	}
	l.Read()
}

// Atoms ...
func (l *LispReader) Atoms() []Atom {
	return l.atoms
}

// WriteAtom ...
func (l *LispReader) WriteAtom(a Atom) {
	l.atoms = append(l.atoms, a)
}

// WriteToken ...
func (l *LispReader) WriteToken(rn rune) (int, error) {
	return l.token.WriteRune(rn)
}

// ReadToken ...
func (l *LispReader) ReadToken() string {
	return l.token.String()
}

// ResetToken ...
func (l *LispReader) ResetToken() {
	l.token.Reset()
}

// PeekRune ...
func (l *LispReader) PeekRune() rune {
	rn := l.read1()
	l.UnreadRune()
	return rn
}

// read1 performs custom error-wrapping for the ReadRune method
func (l *LispReader) read1() rune {
	rn, _, err := l.ReadRune()
	if err == io.EOF {
		return EOF
	}
	return rn
}

func (l *LispReader) tokenAsAtom(atomType AtomType) {
	atom := NewAtom(l, atomType)
	l.WriteAtom(atom)
	l.ResetToken()
}

/////////////////////////////////////////////////////////////////////////////
///   Support Functions   ///////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////

func readLeftParen(l *LispReader) {
	log.Tracef("Reading left paren and emitting atom ...\n")
	l.tokenAsAtom(AtomLeftParen)
}

func readRightParen(l *LispReader) {
	log.Tracef("Reading right paren and emitting atom ...\n")
	l.tokenAsAtom(AtomRightParen)
}

func readEndOfFile(l *LispReader) {
	log.Tracef("Reading end of file and emitting atom ...\n")
	l.tokenAsAtom(AtomEOF)
}

func readLeftVect(l *LispReader) {
	l.tokenAsAtom(AtomLeftVect)
}

func readRightVect(l *LispReader) {
	l.tokenAsAtom(AtomRightVect)
}

func readString(l *LispReader) {
	// The opening quote was written to .token in the Read method; now we need
	// to remove that and write the rest of the string
	l.ResetToken()
	for rn := l.read1(); !isStringEnd(rn); rn = l.read1() {
		if isUnexpectedStringEnd(rn) {
			log.Error(UnterminatedQuotedStringError)
			return
		}
		l.WriteToken(rn)
	}
	l.tokenAsAtom(AtomString)
}

/////////////////////////////////////////////////////////////////////////////
///   Utility Functions   ///////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////

// isWhitespace reports whether a given rune is considered whitespace
func isWhitespace(rn rune) bool {
	return rn == space || rn == tab || rn == newline || rn == optionalCollectionDelimiter
}

func isStringEnd(rn rune) bool {
	return rn == doubleQuote
}
func isUnexpectedStringEnd(rn rune) bool {
	return rn == EOF || rn == newline
}

// ReadAtomsData ...
func ReadAtomsData(l *LispReader) []string {
	// XXX use string builder here instead
	lexedStrings := []string{}
	for _, atom := range l.atoms {
		if atom.Type == AtomEOF {
			break
		}
		lexedStrings = append(lexedStrings,
			fmt.Sprintf("%-3s: row: %d, col: %2d, abs: %2d, type %s",
				atom.Value, atom.Row(), atom.Column(),
				atom.Absolute(), AtomName(atom.Type)))
	}
	return append(lexedStrings, "")
}

// FormatAtomsData prints the value of String()
func FormatAtomsData(l *LispReader) string {
	atoms := []string{""}
	atoms = append(atoms, ReadAtomsData(l)...)
	atoms = append(atoms, "")
	return strings.Join(atoms, "\n")
}

// PrintAtomsData prints the value of String()
func PrintAtomsData(l *LispReader) {
	fmt.Print(FormatAtomsData(l))
}
