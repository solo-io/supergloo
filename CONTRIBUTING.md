# Contributing to Service Mesh Hub

The following is a set of guidelines for contributing to Service Mesh Hub. This should not be considered a set of
hard and fast rules; a particular section of code may not conform to these guidelines because of some extenuating reason.
In those cases, best judgement should be used.

#### Table of Contents

- [Style Guideline](#style-guideline)
    - [Package Layout](#package-layout)  
        * [Naming](#naming)
        * [Interfaces and Mocks](#interfaces-and-mocks)
        * [Common Clients](#common-clients)
    - [Component Structure](#component-structure)
        * [Dependencies](#dependencies)
        * [Component Implementation](#component-implementation)
        * [Externally-Accessible Types](#externally-accessible-types)
    - [Test Code](#test-code)
        * [Test Package Layout](#test-package-layout)
        * [Test Design](#test-design)
        * [Coverage](#coverage)
        * [Test Initialization](#test-initialization)
    - [Code Style](#code-style)
        * [Line Breaks](#line-breaks)
        * [Context Parameters](#context-parameters)
- [Development](#development)
    - [Logging](#logging)
    - [Running Tests](#running-tests)
    

## Style Guideline

### Package Layout

#### Naming
`main` functions should be put in a `cmd` package, with a `Dockerfile` living there as well. A `pkg` directory should
live alongside `cmd`, and should contain the main implementation of the larger component.

You will see packages named `common` scattered throughout the codebase, but usages of a `common` package is
discouraged; instead, there should be an effort to put things in a package whose name conveys the semantics of the package.
For example, if a client implementation is used in several places, it should be put in a `clients` package, for
example, rather than a `common`.

#### Interfaces and Mocks
A package should have a single file named `interfaces.go`, in which all the exported interfaces are placed along with the
`go:generate mockgen` invocation to generate the accompanying mocks. That will normally look something like:

```go
// go:generate mockgen -source ./interfaces.go -destination ./mocks/mock_interfaces.go
NOTE: there is a space before the `go:generate` in the line above;
that is to prevent the build from running this specific line in this file (CONTRIBUTING.md), which will fail
In real usage, you would remove the space between // and go:generate
``` 

#### Common Clients

Simple clients for doing CRUD operations on k8s kinds should live in `pkg/clients`. The constructor for that client
should take a dynamic client (`client.Client`) from `controller-runtime`, and the methods should just forward on the operations to the
underlying dynamic client. 

Clients that are smarter than just CRUD operations should live alongside whatever component needs them. When they become
sufficiently used to warrant moving them (subjective), they should be placed in a package that makes sense as a package
that should be exporting a commonly-used client (but note the point above about avoiding `common` packages).

### Component Structure

#### Dependencies

All of a component's dependencies should be injected, either by coding against an interface rather than an implementation,
or providing your component with a way to build an instance that it needs at runtime (e.g., the factory pattern).
This is to facilitate better component design (ability to swap out implementations (including mock implementations), avoiding subtle dependencies on
heavyweight components) as well as unit tests that have both better coverage and are guaranteed not to require 
heavy dependencies at test runtime.

Examples of dependencies that should be injected are:

- Kubernetes clients
- anything that does file I/O
- other components you've written that your component depends on (i.e., never directly `new` in a component, as then you're bound
to whatever that implementation is) 

#### Component Implementation

There should generally be a one-to-one correspondence between .go files and implementations. A file should have a single
constructor function, prefixed with `New`, that is declared to return an interfaces from `interfaces.go`.

Example:

`interfaces.go`:

```go
package foo

type Bar interface {
    Method()
}
```

`bar.go`:

```go
package foo

func NewBar() Bar {
  return &barImplementation{}
}

type barImplementation struct {
  ...
}

func (b *barImplementation) Method() {
  ...
}
```

#### Externally-Accessible Types

All types that are accessible from outside their package should be explicitly exported by their package. An un-exported
type that the user can get an instance of is hard to work with; for example, they cannot use that type as a field in one
of their own structs, or pass it to a function they have declared. As a concrete example,
the following is discouraged:

```go
package foo

type myStruct{}

func ExportedFunction() *myStruct {
	
}
```

### Test Code

#### Test Package Layout

Code that tests package `foo` should live in the `foo_test` package, and physically alongside the implementation that it covers, in
the same directory. Keeping the test code in its own package ensures that the API of your component is being tested in a way that
mirrors actual usage of the component, and can help ensure that the component you're testing is usable in a real client
situation. If there is a particular function or component that is private to the `foo` package but is sufficiently complex enough
that you'd like to test it in isolation, that is normally a good indicator that that piece should be split off into its own
component, tested independently, and injected into any place that needs it.

#### Test Design

As a result of the above section, unit tests will not have access to a component's struct implementation. As a result, unit tests
should be written against a component's interface type. Only methods on the struct implementation that satisfy the interface definition
should be made public. As a result, unit tests must test the component as it is exported to users/clients of that component. This
guarantees that the behavior of the component under test is the same as the behavior of the component in real use.

Note that this implies that our definition of "unit" that we have adopted is the component level. We do not cover individual functions
from an implementation in a unit test. While this can make tests more cumbersome to write at first, as it will be slightly harder to get
code coverage in all the places you would like, we believe that this practice will, long-term, reduce test code churn and better ensure
that your behavior is tested as a client would see it. Anecdotally, we have seen bugs get caught with this approach that a finer-grained
test philosophy did not catch.

See [this article](https://www.artima.com/suiterunner/private.html) for thoughts on this approach. If your tests are particularly hard to write,
that is a suggestion that your abstraction design is not fine-grained enough, and there is a new component trying to escape from the one
you're writing. However, note that we only recommend splitting a component out to a new one if there are good business-logic-semantics reasons
for doing so; components should not turn into grab bags of random functions that you'd like to test directly.

A good summary from the book "Pragmatic Unit Testing" by Thomas/Hunt:

> In general, you don't want to break any encapsulation for the sake of testing... Most of the time, you should be able to test a class
> by exercising its public methods. If there is significant functionality that is hidden behind private or protected access, that might 
> be a warning sign that there's another class in there struggling to get out. 

#### Coverage

While code coverage is important, we do not strive for 100%. Much of the code we interact with has no business logic
(i.e., it's purely "glue" code) and there is little value in testing it. A good example of this is our simple
per-kind clients. Since they just forward on the operations to the underlying dynamic client, there is no real reason
to test them because no business logic is being exercised.

#### Test Initialization

Say you're working in directory `dir/` and writing a test for component `bar.go`. You can get test files bootstrapped in 
an idiomatic shape by running the following commands:

```bash
cd dir/
ginkgo bootstrap       # if you haven't done this already- creates a dir_suite_test.go file
ginkgo generate bar.go # initializes a test file for covering bar
```

Ensure that the string constants generated into those files make sense. Feel free to change them to something more meaningful.
I.e., we shouldn't have test code that says `Describe("translator")`. Something like `Describe("Istio Traffic Policy Translator")`
may be better.

### Code Style

#### Line Breaks

It is generally a matter of best-judgement for when a line should be broken up; if a line is very long (~120 columns)
or, more vaguely, very dense, consider splitting it up into multiple expressions or just breaking the line. This will
often be encountered when writing constructors or other functions that take many arguments. Those cases should look like:

```go
func ManyParameters(
    param1 int,
    param2 string,
    param3 int,
) ReturnType {
    
    // include a newline after the return type declaration above, then your function can continue as normal
    return &ReturnType{}
}
``` 

#### Context Parameters

It is idiomatic Go style for a `context.Context` parameter to be the first item in a function's parameter list.
In general, also, `context`s should not be held onto by an object for the duration of its lifetime.

You will see an exception to this last point in several places throughout the codebase. This is because those places
mainly serve as receivers of callbacks that do not pass along a `context`, but those places also need to initiate
kubernetes client actions that require a `context`. This is not ideal. Those long-lived `context` references
also often serve as contexts backing `controller-runtime` managers; when those contexts are stopped, that
indicates that the associated cluster has gone offline.  

## Development

### Logging
Logging in Service Mesh Hub uses [zap](https://github.com/uber-go/zap). By default the log level will be set to info,
and the encoder will be a JSON encoder. This means that debug level logs will not be shown, and logs will be outputted 
to `os.Stderr` in a machine readable format.

For debugging purposes this can be controlled with the following ENV variables:
1. `LOG_LEVEL`
    * Must be set to a valid `zapcore.Level`. Options can be found [here](https://pkg.go.dev/go.uber.org/zap/zapcore?tab=doc#Level)
    * If none is provided, or the option is invalid, `Info` will be used by default
2. `DEBUG_MODE`
    * If set, the log encoder will be switched to a human readable format. This means it will not be valid JSON anymore.
    More information on how the data will be formatted can be found [here](https://pkg.go.dev/go.uber.org/zap/zapcore?tab=doc#NewConsoleEncoder)
    * Note: If set, this will override the logging level to `debug`, no matter what `LOG_LEVEL` is set to.

### Running Tests

You can run tests through Goland configurations, but ultimately that will not be how our CI runs the tests, which is through `ginkgo`. Deviations in behavior
can occur in Goland as opposed to `ginkgo`. For example, Goland does not require a `_suite_test.go` file, while `ginkgo` will
refuse to run any of the test cases without a `_suite_test.go` file.

We normally run the tests with:

```bash
ginkgo -r -race -progress -randomizeAllSpecs -randomizeSuites -compilers=4 -failOnPending -p
```

These args ensure that:
* Race detection is turned on
* Test progress is printed out
* Test spec order is randomized, eliminating the possibility of inter-dependence of one spec on another
* Test suite is randomized for the same reason
* Turn down the number of compilers to speed up execution
* Fail on specs marked as pending with `XIt` or `FIt` or similar
* Parallelize the whole thing
