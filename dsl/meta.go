package dsl

import (
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

// Meta defines a set of key/value pairs that can be assigned to an object. Each
// value consists of a slice of strings so that multiple invocation of the Meta
// function on the same target using the same key builds up the slice.
//
// Meta may appear in attributes, result types, endpoints, responses, services
// and API definitions.
//
// While keys can have any value the following names have special meanings:
//
// - "type:generate:force" forces the code generation for the type it is defined
// on. By default goa only generates types that are used explicitly by the
// service methods. The value is a slice of strings that lists the names of the
// services for which to generate the struct. The struct is generated for all
// services if left empty.
//
//	package design
//
//	var _ = Service("service1", func() { ... })
//	var _ = Service("service2", func() { ... })
//
//	var Unused = Type("Unused", func() {
//	    Attribute("name", String)
//	    Meta("type:generate:force", "service1", "service2")
//	})
//
// - "struct:error:name" DEPRECATED, use ErrorName instead.
//
// - "struct:pkg:path" overrides where the Go type generated for the enclosing
// user type definition is generated. Both the file location and Go package name
// are overridden. The file location is computed by appending the location of
// the gen folder with the string given as second argument to Meta. So for
// example Meta("struct:pkg:path", "foo/bar") will generate the Go type in the
// file "gen/foo/bar/<name>.go" where <name> is the name of the type.
// Additionally the package name is computed by taking the base path of the
// provided location ("bar" in the prior example). The example below causes the
// Go type declaration for MyType to live in the file gen/types/mytype.go with
// the package name "types":
//
//	package design
//
//	var MyType = Type("MyType", func() {
//	    Attribute("name")
//	    Meta("struct:pkg:path", "types")
//	})
//
// Note: If that meta tag is used more that once in the same design, but with
// different values in the meta statement (ex. one type has Meta("struct:pkg:path", "types1")
// and another has Meta("struct:pkg:path", "types2")) then those two types cannot
// both contain a field of the same user type.
// For the same reason, you may not set a different custom package in a user type than
// the one set on a containing user type.
//
// - "struct:field:name" overrides the Go struct field name generated by default
// by goa. Applicable to attributes only.
//
//	var MyType = Type("MyType", func() {
//	    Attribute("ssn", String, "User SSN", func() {
//	        Meta("struct:field:name", "SSN")
//	    })
//	})
//
// - "struct:field:type" overrides the Go struct field type specified in the
// design, with one caveat; if the type would have been a pointer (such as its
// not Required) the new type will also be a pointer.  Applicable to attributes
// only. The import path of the type should be passed in as the second
// parameter, if needed.  If the default imported package name conflicts with
// another, you can override that as well with the third parameter.
//
//	var MyType = Type("BigOleMessage", func() {
//	    Attribute("type", String, "Type of big payload")
//	    Attribute("bigPayload", String, "Don't parse it if you don't have to",func() {
//	        Meta("struct:field:type","json.RawMessage","encoding/json")
//	     })
//	     Attribute("id", String, func() {
//	         Meta("struct:field:type","bison.ObjectId", "github.com/globalsign/mgo/bson", "bison")
//	     })
//	})
//
// - "struct:field:proto" overrides the generated protobuf field type. If the
// type is defined in a separate proto file, the last three elements define the
// proto file import path, Go type name and Go import path respectively.
//
//	var Timestamp = Type("Timestamp", func() {
//	    Description("Google timestamp compatible design")
//	    Field(1, "seconds", Int64, "Unix timestamp in seconds", func() {
//	        Meta("struct:field:proto", "int64") // Goa generates sint64 by default
//	    })
//	    Field(2, "nanos", Int32, "Unix timestamp in nanoseconds", func() {
//	        Meta("struct:field:proto", "int32") // Goa generates sint32 by default
//	    })
//	})
//
//	var MyType = Type("MyType", func() {
//	    Field(1, "created_at", Timestamp, func() {
//	        Meta("struct:field:proto", "google.protobuf.Timestamp", "google/protobuf/timestamp.proto", "Timestamp", "google.golang.org/protobuf/types/known/timestamppb")
//	    })
//	})
//
// - "struct:tag:xxx" sets a generated Go struct field tag and overrides tags
// that Goa would otherwise set. If the metadata value is a slice then the
// strings are joined with the space character as separator. Applicable to
// attributes only.
//
//	var MyType = Type("MyType", func() {
//	    Attribute("ssn", String, "User SSN", func() {
//	        Meta("struct:tag:json", "SSN,omitempty")
//	        Meta("struct:tag:xml", "SSN,omitempty")
//	    })
//	})
//
// - "protoc:include" provides the list of import paths used to invoke protoc.
// Applicable to API and service definitions only. If used on an API definition
// the include paths are used for all services.
//
//	var _ = API("myapi", func() {
//	    Meta("protoc:include", "/usr/include", "/usr/local/include")
//	})
//
//	var _ = Service("service1", func() {
//	    Meta("protoc:include", "/usr/local/include/google/protobuf")
//	})
//
// - "swagger:generate" DEPRECATED, use "openapi:generate" instead.
//
// - "openapi:generate" specifies whether OpenAPI specification should be
// generated. Defaults to true. Applicable to Server, services, methods and file
// servers.
//
//	var _ = Service("MyService", func() {
//	    Meta("openapi:generate", "false")
//	})
//
// - "swagger:summary" DEPRECATED, use "openapi:summary" instead
//
// - "openapi:summary" sets the OpenAPI operation summary field. The special
// value "{path}" is replaced with the method HTTP path. Applicable to methods
// or to API .
//
//	var _ = Service("MyService", func() {
//	    Method("MyMethod", func() {
//	           Meta("openapi:summary", "Summary of MyMethod")
//	    })
//	})
//
// - "openapi:operationId" sets the OpenAPI operationId field format. The following
// special values will be replaced with operation-specific information:
//
//	"{service}" is replaced with the name of the service
//
//	"{method}" is replaced with the name of the method
//
//	"(#{routeIndex})" is replaced with the index of the path in cases where a
//	 method has more than one route associated with it. The index will never be added
//	 when only one route exists. The # character may be swapped for any content you
//	 wish to use as a spacer between the preceding content and the route index.
//
// If you wish to specify a static operationId, omitting any of the above special values
// will render the operationId as a literal.
//
// Defaults to "{service}#{method}(#{routeIndex})". Applicable to methods, services, or
// to API.
//
//	var _ = Service("MyService", func() {
//	    Method("MyMethod", func() {
//	           // Generates MyService.MyMethod
//	           Meta("openapi:operationId", "{service}.{method}(.{routeIndex})")
//	    })
//	})
//
// - "swagger:example" DEPRECATED, use "openapi:example" instead
//
// - "openapi:example" specifies whether to generate random example. Defaults to
// true. Applicable to API (applies to all attributes) or individual attributes.
//
//	var _ = API("MyAPI", func() {
//	    Meta("openapi:example", "false")
//	})
//
// - "swagger:tag:xxx" DEPRECATED, use "openapi:tag:xxx" instead
//
// - "openapi:tag:xxx" sets the OpenAPI object field tag xxx. Applicable to
// HTTP services and methods. Tags are defined on services and used by methods.
//
//	var _ = Service("MyService", func() {
//	    HTTP(func() {
//	    	Meta("openapi:tag:Backend:desc", "Description of Backend")
//	    	Meta("openapi:tag:Backend:url", "http://example.com")
//	    	Meta("openapi:tag:Backend:url:desc", "See more docs here")
//	    	Meta("openapi:tag:Backend:extension:x-data", `{"foo":"bar"}`)
//	    })
//	    Method("MyMethod", func() {
//	        HTTP(func() {
//	     	   Meta("openapi:tag:Backend")
//	    	})
//	    })
//	})
//
// - "swagger:extension:xxx" DEPRECATED, use "openapi:extension:xxx" instead
//
// - "openapi:extension:xxx" sets the OpenAPI extensions xxx. The value can be
// any valid JSON. Applicable to API (OpenAPI info and tag objects), Service
// (OpenAPI paths object), Method (OpenAPI path-item object), Route (OpenAPI
// operation object), Param (OpenAPI parameter object), Response (OpenAPI
// response object) and Security (OpenAPI security-scheme object). See
// https://github.com/OAI/OpenAPI-Specification/blob/master/guidelines/EXTENSIONS.md.
//
//	var _ = API("MyAPI", func() {
//	    Meta("openapi:extension:x-api", `{"foo":"bar"}`)
//	})
//
// - "openapi:typename" overrides the name of the type generated in the OpenAPI specification.
// Applicable to types (including embedded Payload and Result definitions).
//
//	var Foo = Type("Foo", func() {
//	    Attribute("name", String)
//	    Meta("openapi:typename", "Bar")
//	})
func Meta(name string, value ...string) {
	appendMeta := func(meta expr.MetaExpr, name string, value ...string) expr.MetaExpr {
		if meta == nil {
			meta = make(map[string][]string)
		}
		meta[name] = append(meta[name], value...)
		return meta
	}

	switch e := eval.Current().(type) {
	case *expr.APIExpr:
		e.Meta = appendMeta(e.Meta, name, value...)
	case *expr.ServerExpr:
		e.Meta = appendMeta(e.Meta, name, value...)
	case *expr.AttributeExpr:
		e.Meta = appendMeta(e.Meta, name, value...)
	case *expr.ResultTypeExpr:
		e.Meta = appendMeta(e.Meta, name, value...)
	case *expr.MethodExpr:
		e.Meta = appendMeta(e.Meta, name, value...)
	case *expr.ServiceExpr:
		e.Meta = appendMeta(e.Meta, name, value...)
	case *expr.HTTPServiceExpr:
		e.Meta = appendMeta(e.Meta, name, value...)
	case *expr.HTTPEndpointExpr:
		e.Meta = appendMeta(e.Meta, name, value...)
	case *expr.RouteExpr:
		e.Meta = appendMeta(e.Meta, name, value...)
	case *expr.HTTPFileServerExpr:
		e.Meta = appendMeta(e.Meta, name, value...)
	case *expr.HTTPResponseExpr:
		e.Meta = appendMeta(e.Meta, name, value...)
	case expr.CompositeExpr:
		att := e.Attribute()
		att.Meta = appendMeta(att.Meta, name, value...)
	default:
		eval.IncompatibleDSL()
	}
}
