package vesupro

import (
    "io"
    "fmt"
)

type VesuproObject interface {
    Dispatch(c *MethodCall) (VesuproObject, error)
    MarshalJSON()([]byte, error)
}

func Evaluate(output io.Writer, program io.Reader,
symTable map[string]VesuproObject) error {
    var err error

    t := NewTokenizer(program)
    defs, err := ParseDefinitions(t)

    if err != nil { return err }

    first := true
    output.Write([]byte{'{'})

    for _, def := range defs {
        if first {
            first = false
            output.Write([]byte{'"'})
        } else {
            output.Write([]byte(",\n\""))
        }
        output.Write([]byte(def.TargetName))
        output.Write([]byte(`":`))

        rcvObj, found := symTable[def.ReceiverName]
        if !found {
            return fmt.Errorf("Receiver not found %s.", def.TargetName)
        }

        for _, call := range def.MethodCalls {
            rcvObj, err = rcvObj.Dispatch(call)
            if err != nil { return err }
        }
        jsonOut, err := rcvObj.MarshalJSON()
        if err != nil { return err }
        output.Write(jsonOut)
    }

    output.Write([]byte{'}'})

    return nil
}
