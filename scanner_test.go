package vesupro_test

import (
    "./"
    "testing"
    "bytes"
)

func TestScanner_Scan(t *testing.T) {
    var tests = []struct {
		s   string
		tok vesupro.Token
		lit string
        ignoreWS bool
	}{
		// Special tokens (EOF, ILLEGAL, WS)
		{s: ``, tok: vesupro.EOF},
		{s: `#`, tok: vesupro.ILLEGAL, lit: `#`},
		{s: ` `, tok: vesupro.WS, lit: " "},
		{s: "\t", tok: vesupro.WS, lit: "\t"},
		{s: "\n", tok: vesupro.WS, lit: "\n"},

        {s: "someIdent", tok: vesupro.IDENT, lit: "someIdent"},
        {s: "illegal-Ident", tok: vesupro.IDENT, lit: "illegal"},
        {s: "-illegalIdent", tok: vesupro.ILLEGAL, lit: "-i"},

        {s: "12.123", tok: vesupro.FLOAT, lit: "12.123"},
        {s: "0.1", tok: vesupro.FLOAT, lit: "0.1"},
        {s: "0.1e-1", tok: vesupro.FLOAT, lit: "0.1e-1"},
        {s: "10E-1", tok: vesupro.FLOAT, lit: "10E-1"},
        {s: "0.1e-100", tok: vesupro.FLOAT, lit: "0.1e-100"},
        {s: "0.0", tok: vesupro.FLOAT, lit: "0.0"},
        {s: "001.", tok: vesupro.FLOAT, lit: "001."}, // no leading zeros
        {s: "0.1ee-100", tok: vesupro.ILLEGAL, lit: "0.1ee"},
        {s: "0.1e-", tok: vesupro.ILLEGAL, lit: "0.1e-"},
        {s: "0.1e", tok: vesupro.ILLEGAL, lit: "0.1e"},

        {s: "123", tok: vesupro.INT, lit: "123"},
        {s: "0", tok: vesupro.INT, lit: "0"},

        {s: `"aaa"`, tok: vesupro.STRING, lit: `"aaa"`},

        {s: `true`, tok: vesupro.TRUE, lit: `true`},
        {s: `false`, tok: vesupro.FALSE, lit: `false`},
        {s: `null`, tok: vesupro.NULL, lit: `null`},


        {s: `"someString"`, tok: vesupro.STRING, lit: `"someString"`},
        {s: `"some\u0020String"`, tok: vesupro.STRING, lit: `"some\u0020String"`},
        {s: `"some\ua020String"`, tok: vesupro.STRING, lit: `"some\ua020String"`},
        {s: `"some\\String"`, tok: vesupro.STRING, lit: `"some\\String"`},
        {s: `"some\uString"`, tok: vesupro.ILLEGAL, lit: `"some\uS`},


        {s: `  "ignoreWS"`, tok: vesupro.STRING, lit: `"ignoreWS"`, ignoreWS: true},
    }

    for i, tt := range tests {
		s := vesupro.NewTokenizer(bytes.NewBufferString(tt.s))
		tok := vesupro.Scan(s, tt.ignoreWS)
        lit := string(s.CurrentToken())
		if tt.tok != tok {
			t.Errorf("%d. %q token mismatch: exp=%d got=%d <%q>", i, tt.s,
            tt.tok, tok, lit)
		} else if tt.lit != lit {
			t.Errorf("%d. %q literal mismatch: exp=%q got=%q", i, tt.s,
            tt.lit, lit)
		}
	}
}
