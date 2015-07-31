package vesupro

import (
    "io"
    "errors"
    "fmt"
)

type VesuproObject interface {
    Dispatch(methodName string, t Tokenizer) (VesuproObject, error)
    MarshalJSON()([]byte, error)
}

func newParseError(expected string, gotTokenId Token, gotToken []byte) error {
    return errors.New(fmt.Sprintf(
        "Parse Error: Expected %s, got token id %d (%q).", expected,
        gotTokenId, string(gotToken)))
}

func parseCallChain(t Tokenizer, symTable map[string]VesuproObject,
receiver VesuproObject) (VesuproObject, error) {
    for {
        tok := Scan(t, true);

        switch tok {
        case SEMI:
            return receiver, nil
        case DOT:
            tok = Scan(t, true)
            if tok != IDENT {
                return nil, newParseError("IDENT", tok, t.CurrentToken())
            }

            methodName := string(t.CurrentToken())

            tok = Scan(t, true)
            if tok != OPEN_PAREN {
                return nil, newParseError("OPEN_PAREN", tok, t.CurrentToken())
            }

            newRcvr, err := receiver.Dispatch(methodName, t)
            if err != nil {
                return nil, err
            }

            tok = Scan(t, true)
            if tok != CLOSE_PAREN {
                return nil, newParseError("CLOSE_PAREN", tok, t.CurrentToken())
            }

            return parseCallChain(t, symTable, newRcvr)
        default:
            return nil, newParseError("SEMI or DOT", tok, t.CurrentToken())
        }
    }
}

func Dispatch(output io.Writer, program io.Reader,
symTable map[string]VesuproObject) error {
    t := NewTokenizer(program)
    first := true
    output.Write([]byte{'{'})
    // parse definitions
    tok := Scan(t, true)
    for ; tok == IDENT; tok = Scan(t, true) {
        // beginning of definition target := ...
        if first {
            first = false
            output.Write([]byte{'"'})
        } else {
            output.Write([]byte(",\n\""))
        }
        output.Write(t.CurrentToken())
        output.Write([]byte{'"', ':'})

        tok = Scan(t, true)
        if tok != DEF_OP {
            return newParseError("DEF_OP", tok, t.CurrentToken())
        }

        tok = Scan(t, true)
        rootName := string(t.CurrentToken())
        rootObject, exists := symTable[rootName]
        if !exists {
            return errors.New(fmt.Sprintf(
                "No object with identifier %q", rootName))
        }

        result, err := parseCallChain(t, symTable, rootObject)
        if err != nil {
            output.Write([]byte(fmt.Sprintf(`{"error":` +
            `"Parse Error: %q"`, err.Error() )))
            return err
        }

        jsonBuf, err := result.MarshalJSON()
        if err != nil {
            output.Write([]byte(fmt.Sprintf(`{"error":` +
            `"Marshalling Error: %q"`, err.Error() )))
        } else {
            output.Write(jsonBuf)
        }
    }
    output.Write([]byte{'}'})
    return nil
}
