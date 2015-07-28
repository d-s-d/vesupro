package vesupro

type Token int

const (
    ILLEGAL Token = iota
    EOF
    WS     // Whitespace

    IDENT  // identifier

    DEF_OP // :=
    STRING // "..."
    FLOAT  // e.g., 0.1
    INT    // e.g., 1
    BOOL   // true || false
    SEMI   // ;
    DOT    // .
    COMMA  // ,

    OPEN_PAREN     // '('
    CLOSE_PAREN    // ')'
    OPEN_BRACKET   // '['
    CLOSE_BRACKET  // ']'

    TRUE  // true
    FALSE // false
    NULL  // null

    JSON  // fast scan JSON
)
