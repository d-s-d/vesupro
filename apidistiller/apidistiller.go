package apidistiller

import (
    "fmt"
    "regexp"
    "go/ast"
)

// BasicTypes maps basic go types to the corresponding vesupro tokens.
var BasicTypes = map[string][]string{
    "uint": []string{"vesupro.INT"},
    "uint8": []string{"vesupro.INT"},
    "uint16": []string{"vesupro.INT"},
    "uint32": []string{"vesupro.INT"},
    "uint64": []string{"vesupro.INT"},
    "byte": []string{"vesupro.INT"},
    "int": []string{"vesupro.INT"},
    "int8": []string{"vesupro.INT"},
    "int16": []string{"vesupro.INT"},
    "int32": []string{"vesupro.INT"},
    "int64": []string{"vesupro.INT"},
    "rune": []string{"vesupro.INT"},
    "float32": []string{"vesupro.FLOAT"},
    "float64": []string{"vesupro.FLOAT"},
    "complex32": []string{"vesupro.FLOAT"},
    "complex64": []string{"vesupro.FLOAT"},
    "bool": []string{"vesupro.TRUE", "vesupro.FALSE"},
    "string": []string{"vesupro.STRING"},
}

// VesuproRegexp is the regular expression used to identify functions which
// belong to the Api.
var VesuproRegexp = regexp.MustCompile("^//\\s*vesupro:\\s*export.*$")

// # API Representation #

// Parameter represents a formal parameter of an exported method.
type Parameter struct {
    Position uint   // position of the argument
    TypeName string // name of the type 'A' for '*A' for instance

    // Currently, only two type of parameters are supported:
    // 1. basic non-pointers types (int, uint, ...)
    // 2. pointer to struct types
    // if IsStruct is true, the type is *A, where A is a struct type
    IsStruct bool
}

// Method represents an exported method of the API.
type Method struct {
    Name string
    Params []*Parameter
}

// API represents the api
type API struct {
    // maps receiver type to functions
    Methods map[string] []*Method
    PackageName string
}

// NewAPI Creates a new Api.
func NewAPI(pkgName string) *API {
    return &API{make(map[string][]*Method, 0), pkgName}
}

// DistillFromAstFile distills the API from the provided ast.file
func (api *API) DistillFromAstFile(f *ast.File) error {
    if f.Name.Name != api.PackageName {
        return fmt.Errorf("Including methods from different packages in the " +
            "same API is not supported (api package %q, file package %q).",
            api.PackageName, f.Name.Name)
    }

    // iterate through declarations
    for _, decl := range f.Decls {
        fDecl, ok := decl.(*ast.FuncDecl)
        // we ignore functions which do not have a receiver
        if !ok || fDecl.Recv == nil { continue; }

        // check whether function is supposed to be exported
        match := false
        if fDecl.Doc != nil {
            for _, comment := range fDecl.Doc.List {
                match = VesuproRegexp.MatchString(comment.Text)
                if match { break }
            }
        }
        if !match { continue }

        var receiverTypeName string

        switch t := fDecl.Recv.List[0].Type.(type) {
        case (*ast.StarExpr):
            if ident, ok := t.X.(*ast.Ident); ok {
                receiverTypeName = ident.Name
            }
        case (*ast.Ident):
            receiverTypeName = t.Name
        }

        _, exists := api.Methods[receiverTypeName]
        if !exists {
            api.Methods[receiverTypeName] = make([]*Method, 0)
        }

        // parse methods
        methodCall := &Method{Name: fDecl.Name.Name}
        actualPos := 0
        // parse parameters
        for _, paramField := range fDecl.Type.Params.List {
            parameterTemplate := &Parameter{}

            switch t := paramField.Type.(type) {
            case (*ast.Ident):
                parameterTemplate.TypeName = t.Name
                // check whether it's a basic type
                _, found := BasicTypes[t.Name]
                if !found {
                    return fmt.Errorf("Unsupported Type %s.", t.Name)
                }
            case (*ast.StarExpr):
                // we just assume that this is a struct
                ident, ok := t.X.(*ast.Ident)
                if !ok {
                    return fmt.Errorf(
                        "Error when parsing StarExpr: %s", t)
                }
                _, found := BasicTypes[ident.Name]
                if found {
                    return fmt.Errorf(
                        "Pointers to basic types are not supported (*%s at "+
                        "position %d of method %s).", ident.Name, actualPos,
                        fDecl.Name.Name)
                }
                parameterTemplate.TypeName = ident.Name
                parameterTemplate.IsStruct = true

            } // switch parameter type

            // iterate over names
            var curParam  *Parameter
            for _ = range paramField.Names {
                curParam = &Parameter{}
                *curParam = *parameterTemplate
                curParam.Position = uint(actualPos)
                actualPos++
                methodCall.Params = append(methodCall.Params,
                    curParam)
            }
        } // for parameter field

        api.Methods[receiverTypeName] = append(
            api.Methods[receiverTypeName], methodCall)
    }
    return nil
}
