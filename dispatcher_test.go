package vesupro_test

import (
    "./"
    "testing"
    "bytes"
    "errors"
    "fmt"
)

type MockObject struct {
    OutString *bytes.Buffer
    firstDispatch bool
}

func (m *MockObject) Dispatch(methodName string,
t vesupro.Tokenizer) (vesupro.VesuproObject, error) {

    if !m.firstDispatch {
        m.OutString.WriteString("mockObject")
        m.firstDispatch = true
    }

    m.OutString.WriteRune('.')
    m.OutString.WriteString(methodName)
    m.OutString.WriteRune('(')

    tok := vesupro.Scan(t, true)
    if tok != vesupro.CLOSE_PAREN {
        for {
            m.OutString.WriteString(fmt.Sprintf("%d:", tok))
            m.OutString.Write(t.CurrentToken())

            tok = vesupro.Scan(t, true)
            if tok == vesupro.CLOSE_PAREN {
                break;
            }
            if tok != vesupro.COMMA {
                return nil, errors.New(fmt.Sprintf(
                    "Expected COMMA but got token id %d (%q).",
                    tok, t.CurrentToken()))
            }
            tok = vesupro.Scan(t, true)
            m.OutString.WriteString(", ")
        }
    }
    m.OutString.WriteRune(')')

    return m, nil
}

func (m *MockObject) MarshalJSON() ([]byte, error) {
    out := bytes.Buffer{}

    out.WriteString(`{"OutString": `)
    out.WriteString(fmt.Sprintf("%q", m.OutString.Bytes() ))
    out.WriteRune('}')
    m.OutString = &bytes.Buffer{}
    m.firstDispatch = false
    return out.Bytes(), nil
}

func TestDispatcher(t *testing.T) {
    tests := []struct {
        in string
        out string
    }{
        {in: `v1 := mockObject.test();`,
        out: `{"v1":{"OutString": "mockObject.test()"}}`},

        {in: `v1 := mockObject.test(1);`,
        out: fmt.Sprintf(
            `{"v1":{"OutString": "mockObject.test(%d:1)"}}`, vesupro.INT)},

        {in: `v1 := mockObject.test(1, "string");`,
        out: fmt.Sprintf(
            `{"v1":{"OutString": "mockObject.test(%d:1, %d:\"string\")"}}`,
            vesupro.INT, vesupro.STRING)},

        {in: `v1 := mockObject.test({"a": {"bla": 1}});`,
        out: fmt.Sprintf(
            `{"v1":{"OutString": "mockObject.test(%d:{\"a\": {\"bla\": 1}})"}}`,
            vesupro.JSON)},

        {in: `v1 := mockObject.foo(0.1).bar(true);`,
        out: fmt.Sprintf(
            `{"v1":{"OutString": "mockObject.foo(%d:0.1).bar(%d:true)"}}`,
            vesupro.FLOAT, vesupro.TRUE)},

            {in: `v1 := mockObject.foo(0.1).bar(true);` +
            `v2 := mockObject.test("foobar");`,
        out: fmt.Sprintf(
            `{"v1":{"OutString": "mockObject.foo(%d:0.1).bar(%d:true)"},` +
            "\n" + `"v2":{"OutString": "mockObject.test(%d:\"foobar\")"}}`,
            vesupro.FLOAT, vesupro.TRUE, vesupro.STRING)},
    }


    for i, tt := range tests {
        in := bytes.NewBufferString(tt.in)
        symTable := map[string]vesupro.VesuproObject{
            "mockObject": &MockObject{OutString: &bytes.Buffer{}},
        }
        out := &bytes.Buffer{}

        err := vesupro.Dispatch(out, in, symTable)

        if err != nil {
            t.Errorf("%d. error: %q", i, err)
        } else if tt.out != string(out.Bytes()) {
            t.Errorf("%d. in/out mismatch %q != %q.",
            i, tt.out, string(out.Bytes()))
        }
    }
}
