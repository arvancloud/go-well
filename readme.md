## Make your go imports ***well*** imported!

### Describing The Problem
There is a common convention when writing go applications for importing different external or builtin packages. Something like:
```go
import (
    "net/url"
    "strings"
    
    someAlias "github.com/an-external-package"
    "github.com/another-external-package"
)
```
This convention helps with better reading the source code and prevents from importing wrong packages by accident.
There are different tools that can make import section of your source code like this, But each one has a pitfall sort of speaking.
One of most popular one of them is `goimports` which automatically separates different imports like above sample. But it also disorganizes the import part sometimes. Something like this:
```go
import (
    "net/url"
    "strings"
  
    "github.com/disorganized-external-package"
  
    someAlias "github.com/an-external-package"
    "github.com/another-external-package"
)
```
### The Solution
We made `well` cli tool with ❤️ to help you having less disorganized imports. You just need to run `well` in your project's root and you're good to go. `well` takes care of well formatting your imports in each `go` file in your project, Recursively.

### Installation
You just need to build the tool by:

`$ go build .`

### Usage
Run the built binary in your project's root:

`$ ./well`
