// The MIT License
//
// Copyright (c) 2020 Temporal Technologies Inc.  All rights reserved.
//
// Copyright (c) 2020 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/format"
	"go/types"
	"log"
	"os"
	"strings"
	"text/template"

	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/go/types/typeutil"
)

const generatedFile = "grpc_interceptor.go"

const codeTemplateText = `
// The MIT License
//
// Copyright (c) 2022 Temporal Technologies Inc.  All rights reserved.
//
// Copyright (c) 2020 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

// Code generated by go generate; DO NOT EDIT.

package converter

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	
	batchpb "go.temporal.io/api/batch/v1"
	commonpb "go.temporal.io/api/common/v1"
	historypb "go.temporal.io/api/history/v1"
	commandpb "go.temporal.io/api/command/v1"
	failurepb "go.temporal.io/api/failure/v1"
	interactionpb "go.temporal.io/api/interaction/v1"
	workflowpb "go.temporal.io/api/workflow/v1"
	querypb "go.temporal.io/api/query/v1"
	schedulepb "go.temporal.io/api/schedule/v1"
	workflowservicepb "go.temporal.io/api/workflowservice/v1"
)

// PayloadCodecGRPCClientInterceptorOptions holds interceptor options.
// Currently this is just the list of codecs to use.
type PayloadCodecGRPCClientInterceptorOptions struct {
	Codecs []PayloadCodec
}

// NewPayloadCodecGRPCClientInterceptor returns a GRPC Client Interceptor that will mimic the encoding
// that the SDK system would perform when configured with a matching EncodingDataConverter.
// Note: This approach does not support use cases that rely on the ContextAware DataConverter interface as
// workflow context is not available at the GRPC level.
func NewPayloadCodecGRPCClientInterceptor(options PayloadCodecGRPCClientInterceptorOptions) (grpc.UnaryClientInterceptor, error) {
	s := serviceInterceptor{codecs: options.Codecs}

	return func(ctx context.Context, method string, req, response interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		err := s.process(true, req)
		if err != nil {
			return err
		}

		err = invoker(ctx, method, req, response, cc, opts...)
		if err != nil {
			return err
		}

		return s.process(false, response)
	}, nil
}

type serviceInterceptor struct {
	codecs []PayloadCodec
}

func (s *serviceInterceptor) encodePayload(payload *commonpb.Payload) error {
	var err error
	payloads := []*commonpb.Payload{payload}
	for i := len(s.codecs) - 1; i >= 0; i-- {
		if payloads, err = s.codecs[i].Encode(payloads); err != nil {
			return err
		}
	}
	if len(payloads) != 1 {
		return fmt.Errorf("received %d payloads from codec, expected 1", len(payloads))
	}
	*payload = *payloads[0]
	return nil
}

func (s *serviceInterceptor) encodePayloads(payloadspb *commonpb.Payloads) error {
	var err error
	payloads := payloadspb.Payloads
	for i := len(s.codecs) - 1; i >= 0; i-- {
		if payloads, err = s.codecs[i].Encode(payloads); err != nil {
			return err
		}
	}
	payloadspb.Payloads = payloads
	return nil
}

func (s *serviceInterceptor) decodePayload(payload *commonpb.Payload) error {
	var err error
	payloads := []*commonpb.Payload{payload}
	for _, codec := range s.codecs {
		if payloads, err = codec.Decode(payloads); err != nil {
			return err
		}
	}
	if len(payloads) != 1 {
		return fmt.Errorf("received %d payloads from codec, expected 1", len(payloads))
	}
	*payload = *payloads[0]
	return nil
}

func (s *serviceInterceptor) decodePayloads(payloadspb *commonpb.Payloads) error {
	var err error
	payloads := payloadspb.Payloads
	for _, codec := range s.codecs {
		if payloads, err = codec.Decode(payloads); err != nil {
			return err
		}
	}
	payloadspb.Payloads = payloads
	return nil
}

func (s *serviceInterceptor) process(encode bool, objs ...interface{}) error {
	for _, obj := range objs {
		switch o := obj.(type) {
			case *commonpb.Payload:
				if o == nil { continue }
				if encode {
					if err := s.encodePayload(o); err != nil { return err }
				} else {
					if err := s.decodePayload(o); err != nil { return err }
				}
			case *commonpb.Payloads:
				if o == nil { continue }
				if encode {
					if err := s.encodePayloads(o); err != nil { return err }
				} else {
					if err := s.decodePayloads(o); err != nil { return err }
				}
			case map[string]*commonpb.Payload:
				for _, x := range o { if err := s.process(encode, x); err != nil { return err } }
{{range $type, $record := .}}
		{{if $record.Slice}}
			case []{{$type}}:
				for _, x := range o { if err := s.process(encode, x); err != nil { return err } }
		{{end}}
		{{if $record.Map}}
			case map[string]{{$type}}:
				for _, x := range o { if err := s.process(encode, x); err != nil { return err } }
		{{end}}
			case {{$type}}:
				if o == nil { continue }
				if err := s.process(
					encode,
					{{range $record.Methods -}}
						o.{{.}}(),
					{{end}}
				); err != nil { return err }
{{end}}
		}
	}

	return nil
}

`

var codeTemplate = template.Must(template.New("process").Parse(codeTemplateText))

// TypeRecord holds the state for a type referred to by the workflow service
type TypeRecord struct {
	Methods []string // List of methods on this type that can eventually lead to Payload(s)
	Slice   bool     // The API refers to slices of this type
	Map     bool     // The API refers to maps with this type as the value
	Matches bool     // We found methods on this type that can eventually lead to Payload(s)
}

var records = map[string]*TypeRecord{}

// isSlice returns true if a type is slice, false otherwise
func isSlice(t types.Type) bool {
	switch t.(type) {
	case *types.Slice:
		return true
	}
	return false
}

// isMap returns true if a type is map, false otherwise
func isMap(t types.Type) bool {
	switch t.(type) {
	case *types.Map:
		return true
	}
	return false
}

// elemType returns the elem (value) type for a slice or map
func elemType(t types.Type) types.Type {
	switch typ := t.(type) {
	case *types.Slice:
		return typ.Elem()
	case *types.Map:
		return typ.Elem()
	}
	return t
}

// typeName returns a normalized path for a type
func typeName(t types.Type) string {
	return types.TypeString(elemType(t), func(p *types.Package) string {
		return p.Name() + "pb"
	})
}

// typeMatches returns true if a type:
// Is equal to or is a pointer to any of the desired types
// Is a slice or slice of pointers to any of the desired types
// Is a map where the value is any of the desired types or a pointer to any of the desired types
func typeMatches(t types.Type, desired ...types.Type) bool {
	resolved := resolveType(t).String()
	for _, f := range desired {
		if resolved == f.String() {
			return true
		}
	}

	return false
}

// resolveType returns the underlying type for pointers, slices and maps
func resolveType(t types.Type) types.Type {
	switch typ := t.(type) {
	case *types.Pointer:
		return resolveType(typ.Elem())
	case *types.Slice:
		return resolveType(typ.Elem())
	case *types.Map:
		return resolveType(typ.Elem())
	}

	return t
}

func pruneRecords(input map[string]*TypeRecord) map[string]*TypeRecord {
	result := map[string]*TypeRecord{}

	for typ, record := range input {
		if record.Matches {
			result[typ] = record
		}
	}

	return result
}

// walk iterates the methods on a type and returns whether any of them can eventually lead to Payload(s)
// The return type for each method on this type is walked recursively to decide which methods can lead to Payload(s)
func walk(desired []types.Type, skip []types.Type, typ types.Type) bool {
	// If this type is a slice then walk the underlying type and then make a note we need to encode slices of this type
	if isSlice(typ) {
		result := walk(desired, skip, elemType(typ))
		record := records[typeName(typ)]
		record.Slice = true
		return result
	}

	// If this type is a map then walk the underlying type and then make a note we need to encode maps with values of this type
	if isMap(typ) {
		result := walk(desired, skip, elemType(typ))
		record := records[typeName(typ)]
		record.Map = true
		return result
	}

	typeName := typeName(typ)

	// If we've walked this type before, return the previous result
	if record, ok := records[typeName]; ok {
		return record.Matches
	}

	record := TypeRecord{}
	records[typeName] = &record

	for _, meth := range typeutil.IntuitiveMethodSet(elemType(typ), nil) {
		// Ignore non-exported methods
		if !meth.Obj().Exported() {
			continue
		}

		methodName := meth.Obj().Name()
		// The protobuf types have a Get.. method for every protobuf they refer to
		// We walk only these methods to avoid cycles or other nasty issues
		if !strings.HasPrefix(methodName, "Get") {
			continue
		}

		sig := meth.Obj().Type().(*types.Signature)
		// All the Get... methods return the relevant protobuf as the first result
		resultType := sig.Results().At(0).Type()

		// Skip this method if it returns a type we want to skip
		if typeMatches(resultType, skip...) {
			continue
		}

		// Check if this method returns a Payload(s) or if it leads (eventually) to a Type which refers to a Payload(s)
		if typeMatches(resultType, desired...) || walk(desired, skip, resultType) {
			record.Matches = true
			record.Methods = append(record.Methods, methodName)
		}
	}

	// Return whether this Type can (eventually) lead to Payload(s)
	// This is used in the walk logic above so that our encoding handles intermediate Types between our Request/Response objects and Payload(s)
	return record.Matches
}

type config struct {
	verifyOnly bool
}

func main() {
	var cfg config
	flag.BoolVar(&cfg.verifyOnly, "verifyOnly", false,
		"don't automatically write interceptor, just verify it has not changed")
	flag.Parse()

	conf := &packages.Config{Mode: packages.NeedImports | packages.NeedTypes | packages.NeedTypesInfo}
	pkgs, err := packages.Load(conf, "go.temporal.io/api/workflowservice/v1")
	if err != nil {
		log.Fatal(err)
	}

	servicePkg := pkgs[0]

	pkgs, err = packages.Load(conf, "go.temporal.io/api/common/v1")
	if err != nil {
		log.Fatal(err)
	}

	commonPkg := pkgs[0]
	scope := commonPkg.Types.Scope()
	payloadTypes := []types.Type{
		scope.Lookup("Payloads").Type(),
		scope.Lookup("Payload").Type(),
	}
	skipTypes := []types.Type{
		scope.Lookup("SearchAttributes").Type(),
	}

	scope = servicePkg.Types.Scope()
	// UnimplementedWorkflowServiceServer is auto-generated via our API package
	// The methods on this type refer to all possible Request/Response types so we can use this to walk through all of our protobuf types
	service := scope.Lookup("UnimplementedWorkflowServiceServer")
	if _, ok := service.(*types.TypeName); ok {
		for _, meth := range typeutil.IntuitiveMethodSet(service.Type(), nil) {
			if !meth.Obj().Exported() {
				continue
			}

			sig := meth.Obj().Type().(*types.Signature)
			walk(payloadTypes, skipTypes, sig.Params().At(1).Type())
			walk(payloadTypes, skipTypes, sig.Results().At(0).Type())
		}
	}

	buf := &bytes.Buffer{}

	err = codeTemplate.Execute(buf, pruneRecords(records))
	if err != nil {
		log.Fatal(err)
	}
	src, err := format.Source(buf.Bytes())
	if err != nil {
		log.Fatal(err)
	}

	if cfg.verifyOnly {
		currentSrc, err := os.ReadFile(generatedFile)
		if err != nil {
			log.Fatal(err)
		}

		if bytes.Compare(src, currentSrc) != 0 {
			log.Fatal(fmt.Errorf("generated file does not match existing file: %s", generatedFile))
		}

		return
	}

	err = os.WriteFile(generatedFile, src, 0666)
	if err != nil {
		log.Fatal(err)
	}
}
