package vesupro

import (
    "bufio"
    "io"
    "unicode/utf8"
    "unicode"
    "bytes"
)


// INTERFACES
type Tokenizer interface {
    Read() rune
    Unread()

    StartToken()
    CurrentToken() []byte
}

// TYPES
type BufferedRuneStream struct {
    start int
    off int
    lastSize int

    data []byte
}

type RuneStream struct {
    reader *bufio.Reader
    buf *bytes.Buffer
    lastSize int
}

func (s* RuneStream) Read() rune {
    ch, lastSize, err := s.reader.ReadRune()
	if err != nil {
		return eof
	}
    s.lastSize = lastSize
    s.buf.WriteRune(ch)
	return ch
}

func (s* RuneStream) Unread() {
    if s.lastSize > 0 {
        err := s.reader.UnreadRune()
        if err != nil {
            s.buf.Truncate(s.buf.Len() - s.lastSize)
        }
    }
}

func (s* RuneStream) StartToken() {
    s.buf = &bytes.Buffer{}
}

func (s* RuneStream) CurrentToken() []byte {
    return s.buf.Bytes()
}

func (s* BufferedRuneStream) Read() rune {
    ch, lastSize := utf8.DecodeRune(s.data[s.off:])
    s.lastSize = lastSize
    s.off = s.off + lastSize
    return ch
}

func (s* BufferedRuneStream) Unread() {
    s.off -= s.lastSize
}

func (s* BufferedRuneStream) StartToken() {
    s.start = s.off
}

func (s* BufferedRuneStream) CurrentToken() []byte {
    return s.data[s.start:s.off]
}

func NewTokenizer(r io.Reader) Tokenizer {
    byteStream, isBytesBuffer := r.(*bytes.Buffer)
    if isBytesBuffer {
        return &BufferedRuneStream{start:0, off:0,data:byteStream.Bytes()}
    }
    return &RuneStream{ reader: bufio.NewReader(r), buf: &bytes.Buffer{} }
}

func Scan(t Tokenizer, ignoreWS bool) (tok Token) {
    t.StartToken()
    ch := t.Read()
    tok = ILLEGAL

    if isWhitespace(ch) {
        // consume whitespace
        tok = scanWhitespace(t)
        if ignoreWS {
            return Scan(t, false)
        }
    } else if isIdentStart(ch) {
        // consume ident
        tok = scanIdent(t)
    } else if isDigit(ch) {
        tok = scanNumber(t, ch)
    } else {
        switch(ch) {
        case '"': tok = scanString(t)
        case '.': tok = DOT
        case ',': tok = COMMA
        case '-': tok = scanNumber(t, ch)
        case ';': tok = SEMI
        case '{': tok = OPEN_BRACE
        case '}': tok = CLOSE_BRACE
        case '[': tok = OPEN_BRACKET
        case ']': tok = CLOSE_BRACKET
        case '(': tok = OPEN_PAREN
        case ')': tok = CLOSE_PAREN
        case ':':
            ch = t.Read()
            if ch == '=' {
                tok = DEF_OP
            }
        case eof: tok = EOF
        }
    }
    return
}

func scanWhitespace(t Tokenizer) Token {
    ch := t.Read()

    for isWhitespace(ch) {
        ch = t.Read()
    }

    t.Unread()

    return WS
}

func scanIdent(t Tokenizer) Token {
    ch := t.Read()

    for isIdentLetter(ch) {
        ch = t.Read()
    }

    t.Unread()

    switch string(t.CurrentToken()) {
    case "false":
        return FALSE
    case "true":
        return TRUE
    case "null":
        return NULL
    }

    return IDENT
}

func scanDigits(t Tokenizer) {
    ch := t.Read()
    for isDigit(ch) {
        ch = t.Read()
    }

    t.Unread()
}

func scanNumber(t Tokenizer, ch rune) (tok Token) {
    const (
        Start = iota
        SignificantStart
        Significant
        Fractional
        ExponentSign
        ExponentFirstDigit
        ExponentDigit
        End
    )

    tok = INT
    state := Start
    for ;;ch = t.Read() {
        switch state {
        case Start:
            switch {
            case ch == '-':
                state = SignificantStart
            case isDigit(ch):
                state = Significant
            default:
                tok = ILLEGAL
            }
        case SignificantStart:
            if !isDigit(ch) {
                tok = ILLEGAL
            } else {
                state = Significant
            }
        case Significant:
            switch {
            case isDigit(ch):
            case ch == '.':
                state = Fractional
                tok = FLOAT
            case ch == 'e' || ch == 'E':
                state = ExponentSign
                tok = FLOAT
            default:
                state = End
            }
        case Fractional:
            switch {
            case isDigit(ch):
            case ch == 'e' || ch == 'E':
                state = ExponentSign
            default:
                state = End
            }
        case ExponentSign:
            switch {
            case isDigit(ch):
                state = ExponentDigit
            case ch == '-' || ch == '+':
                state = ExponentFirstDigit
            default:
                tok = ILLEGAL
            }
        case ExponentFirstDigit:
            if !isDigit(ch) {
                tok = ILLEGAL
            } else {
                state = ExponentDigit
            }
        case ExponentDigit:
            if !isDigit(ch) {
                state = End
            }
        }
        if tok == ILLEGAL || state == End {
            break;
        }
    } // loop over ch

    return
}

func scanString(t Tokenizer) Token {
    const (
        InString = iota
        Esc
        EscU
        EscU1
        EscU12
        EscU123
        End
        Error
    )

    state := InString

    for ch := t.Read(); ch != eof; ch = t.Read() {
        switch(state) {
        case InString:
            switch {
            case ch == '"':
                state = End
                return STRING
            case ch == '\\':
                state = Esc
            case ch < 0x20:
                return ILLEGAL
            }
        case Esc:
            switch ch {
            case 'b', 'f', 'n', 'r', 't', '\\', '/', '"':
                state = InString
            case 'u':
                state = EscU
            default:
                return ILLEGAL;
            }
        case EscU, EscU1, EscU12, EscU123:
            if '0' <= ch && ch <= '9' || 'a' <= ch &&
            ch <= 'f' || 'A' <= ch && ch <= 'F' {
                if state != EscU123 {
                    state += 1
                } else {
                    state = InString
                }
            } else {
                return ILLEGAL;
            }
        }
        if state >= End {
            break;
        }
    }

    if state != End {
        return ILLEGAL
    }
    t.Unread()
    return STRING
}

// FastScanJSON scans a json object as one token
// this is useful when using an third-party json parser which expects
// a byte-slice as input (such as json, ffjson, etc.)
func FastScanJSON(t Tokenizer) (tok Token) {
    var nestingLevel int = 0
    // the closing brace may only appear within a string
    for ch := t.Read(); (ch != '}'|| nestingLevel > 0) && ch != eof; ch = t.Read() {
        switch ch {
        case '"':
            scanString(t)
        case '{':
            nestingLevel += 1
        case '}':
            nestingLevel -= 1
        }
    }
    return JSON
}


func isIdentStart( ch rune ) bool {
    return  'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' ||
    ch == '_'
}

func isIdentLetter( ch rune ) bool {
    return  'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' ||
    ch >= '0' && ch <= '9' || ch == '_'
}

func isWhitespace( r rune ) bool { return unicode.IsSpace(r) }
func isDigit( r rune ) bool { return unicode.IsDigit(r) }
func isLetter( r rune ) bool { return unicode.IsLetter(r) }

var eof = utf8.RuneError
