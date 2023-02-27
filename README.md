# Domain

This provides basic definitions for building application domain models in Go
following a DDD approach.

Things included are:

- Entity identifiers generator.
- Domain event recording traits.
- Domain event dispatching and publishing definitions.
- Continuation token pagination primitives.

## Installation

```bash
go get github.com/tangelo-labs/go-domain
```

## Examples

### Defining Aggregate Roots

```go
package main

import (
    "github.com/tangelo-labs/go-domain"
)

type NameChangedEvent struct {
	ID      domain.ID
    Before  string
    After   string
}

type User struct {
    id      domain.ID
    name    string
    age     int

    events.BaseRecorder
}

func NewUser(id domain.ID) *User {
    return &User{
        id: id,
    }
}

func (u *User) ID() domain.ID {
    return u.id
}

func (u *User) ChangeName(n string) {   
    u.Record(NameChangedEvent{
        ID: u.ID(),
        Before: u.name,
        After: n,
    })
    
     u.name = n    
}
```

See the "examples" directory for more complete example.
