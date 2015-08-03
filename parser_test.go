package vesupro_test

import (
    "./"
    "testing"
    "bytes"
    "reflect"
)

func TestParseDefinition(t *testing.T) {
    tests := []struct {
        in string
        out []*vesupro.Definition
    }{
    {in: "v1 := target1.f1(1);",
     out: []*vesupro.Definition{&vesupro.Definition{
            TargetName: "v1",
            ReceiverName: "target1",
            MethodCalls: []*vesupro.MethodCall{&vesupro.MethodCall{
                Name: "f1",
                Arguments: []*vesupro.ArgumentToken {&vesupro.ArgumentToken{
                                TokenType: vesupro.INT,
                                TokenContent: []byte{'1'},
                            },
                        },
                    },
                }, // method calls
            },
        },
    },
    }

    for i, tt := range tests {
        in := bytes.NewBufferString(tt.in)
        /*
        symTable := map[string]vesupro.VesuproObject{
            "mockObject": &MockObject{OutString: &bytes.Buffer{}},
        }*/
        //out := &bytes.Buffer{}
        tokzr := vesupro.NewTokenizer(in)

        def, err := vesupro.ParseDefinitions(tokzr)

        if err != nil {
            t.Errorf("%d. error: %q", i, err)
        } else if !reflect.DeepEqual(tt.out, def) {
            t.Errorf("%d. in/out mismatch %q != %q.",
            i, tt.out, def)
        }
    }
}
