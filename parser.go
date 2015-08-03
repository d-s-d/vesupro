package vesupro

import (
    "fmt"
    "strconv"
)

const initMethodCall = 4

type ArgumentToken struct {
    TokenType Token
    TokenContent []byte
}

func (arg *ArgumentToken) ToInt64() (int64, error) {
    if arg.TokenType != INT {
        return 0, fmt.Errorf(
            "ToInt64(): Cannot convert Token of type %d to integer.",
            arg.TokenType)
    }
    return strconv.ParseInt(string(arg.TokenContent), 10, 64)
}

func (arg *ArgumentToken) ToFloat64() (float64, error) {
    if arg.TokenType != FLOAT {
        return 0.0, fmt.Errorf(
            "ToFloat64(): Cannot convert Token of type %d to float.",
            arg.TokenType)
    }
    return strconv.ParseFloat(string(arg.TokenContent), 64)
}

func (arg *ArgumentToken) ToBool() (bool, error) {
    if arg.TokenType != TRUE && arg.TokenType != FALSE {
        return false, fmt.Errorf(
            "ToBool(): Cannot convert Token of type %d to bool.",
            arg.TokenType)
    }
    return arg.TokenType == TRUE, nil
}

func (arg *ArgumentToken) ToString() (string, error) {
    if arg.TokenType != STRING {
        return "", fmt.Errorf(
            "ToString(): Cannot convert Token of type %d to string.",
            arg.TokenType)
    }
    return string(arg.TokenContent), nil
}


type MethodCall struct {
    Name string
    Arguments []*ArgumentToken
}

type Definition struct {
    TargetName string
    ReceiverName string
    MethodCalls []*MethodCall
}

type DefSeq struct {
    Definitions []*Definition
}

func NewMethodCall(name string) *MethodCall {
    return &MethodCall{Name: name}
}

func NewDefinition(targetName string, rcvName string,
    methodCalls []*MethodCall) *Definition {
    return &Definition {
        TargetName: targetName,
        ReceiverName: rcvName,
        MethodCalls: methodCalls,
    }
}

func ParseArgumentList(t Tokenizer) ([]*ArgumentToken, error) {

    tok := Scan(t, true)

    if tok == CLOSE_PAREN {
        return []*ArgumentToken{}, nil
    }

    args := make([]*ArgumentToken, 0, 8)

    for {
        switch tok {
        case INT, FLOAT, STRING, TRUE, FALSE, JSON:
        default:
            return nil, fmt.Errorf(
                "Expected argument token, got %d. (rune pos. %d)", tok,
                t.RuneOffset())
        }
        args = append(args, &ArgumentToken{
            TokenType: tok, TokenContent: t.CurrentToken()})

        tok = Scan(t, true)
        switch tok {
        case COMMA:
            tok = Scan(t, true)
        case CLOSE_PAREN:
            return args, nil
        default:
            return nil, fmt.Errorf(
                "Expected COMMA or CLOSE_PAREN, got %d. (rune pos. %d)",
                tok, t.RuneOffset())
        }
    }
}

func ParseDefinitions(t Tokenizer) ([]*Definition, error) {
    var err error

    defs := make([]*Definition, 0, 2)
    def, err := ParseDefinition(t)
    for ;err == nil && def != nil; def, err = ParseDefinition(t) {
        defs = append(defs, def)
    }
    return defs, err
}

func ParseDefinition(t Tokenizer) (*Definition, error) {
    var err error
    tok := Scan(t, true)

    if tok == EOF { return nil, nil }

    methodCalls := make([]*MethodCall, 0, 4)

    // targetName := rcvName.{funcName([Argument [{, Argument}])}
    // ^        ^ 
    targetName := string(t.CurrentToken())

    // targetName := rcvName.{funcName([Argument [{, Argument}])}
    //            ^^
    err = ScanExpTok(t, DEF_OP, true)
    if err != nil { return nil, err }

    // targetName := rcvName.{funcName([Argument [{, Argument}])}
    //               ^     ^
    err = ScanExpTok(t, IDENT, true)
    if err != nil { return nil, err }
    rcvName := string(t.CurrentToken())

    // targetName := rcvName.funcName([Argument [{, Argument}])}
    //                      ^
    err = ScanExpTok(t, DOT, true)
    if err != nil { return nil, err }

    err = ScanExpTok(t, IDENT, true)
    if err != nil { return nil, err }
    curMethodCall := NewMethodCall(string(t.CurrentToken()))
    methodCalls = append(methodCalls, curMethodCall)

    for {
        err = ScanExpTok(t, OPEN_PAREN, true)
        if err != nil { return nil, err }


        curMethodCall.Arguments, err = ParseArgumentList(t)
        if err != nil { return nil, err }

        tok = Scan(t, true)
        if tok != DOT {
            if tok == SEMI {
                return NewDefinition(targetName, rcvName, methodCalls), nil
            }
            return nil, fmt.Errorf(
                "Expected DOT or SEMI, but got %d.", tok)
        }

        err = ScanExpTok(t, IDENT, true)
        if err != nil { return nil, err }
        curMethodCall = NewMethodCall(string(t.CurrentToken() ))
        methodCalls = append(methodCalls, curMethodCall)
    }
}
