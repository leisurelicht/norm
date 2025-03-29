# norm - Go SQL Query Builder and ORM

`norm` is a lightweight, flexible SQL query builder and ORM for Go applications. It provides a fluent interface for building and executing SQL queries with minimal boilerplate code.

## Features

- Chainable query building API
- Support for complex filters and conditions
- Simple CRUD operations
- Automatic struct-to-table mapping
- Transaction support
- Customizable query operators

## Installation

```go
go get github.com/username/norm
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    
    "github.com/username/norm"
    "github.com/zeromicro/go-zero/core/stores/sqlx"
)

// Define your model
type User struct {
    Id        int64  `db:"id"`
    Username  string `db:"username"`
    Email     string `db:"email"`
    CreatedAt int64  `db:"created_at"`
    IsDeleted bool   `db:"is_deleted"`
}

func main() {
    // Connect to database
    conn := sqlx.NewMysql("user:password@tcp(localhost:3306)/dbname")
    
    // Create controller
    userCtrl := norm.NewController(conn, norm.MysqlOperator{}, User{})
    
    // Use controller with context
    ctx := context.Background()
    ctrl := userCtrl(ctx)
    
    // Create a new user
    user := User{
        Username:  "johndoe",
        Email:     "john@example.com",
        CreatedAt: time.Now().Unix(),
        IsDeleted: false,
    }
    
    id, err := ctrl.InsertModel(user)
    if err != nil {
        panic(err)
    }
    fmt.Printf("Created user with ID: %d\n", id)
    
    // Find users with complex filters
    result, err := ctrl.Filter(
        norm.AND{"username__startswith": "john", "is_deleted": false},
        norm.OR{"email__contains": "@example.com"},
    ).FindAll()
    
    if err != nil {
        panic(err)
    }
    
    for _, user := range result {
        fmt.Printf("Found user: %v\n", user)
    }
}
```

## Core Concepts

### Controller

The Controller is the main entry point for database operations. It provides methods for querying, inserting, updating, and deleting records.

### QuerySet

QuerySet provides a fluent API for building SQL queries with conditions, ordering, limits, etc.

## API Reference

### Creating a Controller

```go
// Create a controller factory for a specific model
userCtrl := norm.NewController(conn, norm.MysqlOperator{}, User{})

// Create a controller instance with context
ctrl := userCtrl(context.Background())
```

### Basic CRUD Operations

#### Insert

```go
// Insert using a map
id, err := ctrl.Insert(map[string]any{
    "username": "johndoe",
    "email": "john@example.com",
})

// Insert using a struct
id, err := ctrl.InsertModel(user)
```

#### Find

```go
// Find all records
results, err := ctrl.FindAll()

// Find all records and map to structs
var users []User
err := ctrl.FindAllModel(&users)

// Find one record
result, err := ctrl.FindOne()

// Find one record and map to struct
var user User
err := ctrl.FindOneModel(&user)
```

#### Update

```go
// Update records matching the filter
num, err := ctrl.Filter(norm.AND{"id": 1}).Update(map[string]any{
    "username": "newusername",
})
```

#### Delete

```go
// Soft delete (setting is_deleted to true)
num, err := ctrl.Filter(norm.AND{"id": 1}).Delete()

// Hard delete (removing from database)
num, err := ctrl.Filter(norm.AND{"id": 1}).Remove()
```

### Filters and Conditions

`norm` supports a variety of filter operators:

| Operator | Description | Example |
|----------|-------------|---------|
| exact | Exact match (default) | `"username": "johndoe"` |
| iexact | Case-insensitive exact match | `"username__iexact": "johndoe"` |
| gt | Greater than | `"age__gt": 18` |
| gte | Greater than or equal | `"age__gte": 18` |
| lt | Less than | `"age__lt": 65` |
| lte | Less than or equal | `"age__lte": 65` |
| in | In a list | `"status__in": []string{"active", "pending"}` |
| between | Between two values | `"age__between": []int{18, 65}` |
| contains | Contains string | `"username__contains": "john"` |
| icontains | Case-insensitive contains | `"username__icontains": "john"` |
| startswith | Starts with string | `"username__startswith": "j"` |
| istartswith | Case-insensitive starts with | `"username__istartswith": "j"` |
| endswith | Ends with string | `"email__endswith": ".com"` |
| iendswith | Case-insensitive ends with | `"email__iendswith": ".COM"` |

#### Using AND/OR conditions

```go
// Complex filtering
result, err := ctrl.Filter(
    "AND", // Default conjunction is AND
    norm.AND{"username__startswith": "john", "is_deleted": false},
    norm.OR{"email__contains": "@example.com", "email__contains": "@company.com"},
).FindAll()

// Using NOT operator
result, err := ctrl.Filter(
    norm.AND{"username__not_in": []string{"admin", "system"}},
).FindAll()
```

### Ordering, Limit, and Pagination

```go
// Order by fields (prefix with - for descending)
results, err := ctrl.OrderBy([]string{"created_at", "-username"}).FindAll()

// Limit and pagination (pageSize, pageNumber)
results, err := ctrl.Limit(10, 1).FindAll()
```

### Selecting Specific Columns

```go
// Select specific columns
results, err := ctrl.Select([]string{"id", "username", "email"}).FindAll()
```

### Group By and Having

```go
// Group by with having clause
results, err := ctrl.GroupBy([]string{"status"}).Having("COUNT(*) > ?", 5).FindAll()
```

### Transactions

```go
session := sqlx.NewSession(conn)
err := session.TransactCtx(context.Background(), func(ctx context.Context, s sqlx.Session) error {
    ctrl := userCtrl(ctx).WithSession(s)
    
    // Perform operations within transaction
    _, err := ctrl.Insert(map[string]any{"username": "transaction_user"})
    if err != nil {
        return err // Transaction will be rolled back
    }
    
    return nil // Transaction will be committed
})
```

### Other Utility Methods

```go
// Check if a record exists
exists, err := ctrl.Filter(norm.AND{"username": "johndoe"}).Exist()

// Get count of records
count, err := ctrl.Filter(norm.AND{"is_deleted": false}).Count()

// Get or create record
result, err := ctrl.GetOrCreate(map[string]any{
    "username": "johndoe",
    "email": "john@example.com",
})

// Create or update record
created, id, err := ctrl.CreateOrUpdate(
    map[string]any{"email": "updated@example.com"},
    norm.AND{"username": "johndoe"},
)
```

## Extending Norm

### Custom Operators

You can implement the `Operator` interface to customize how SQL operations are executed:

```go
type CustomOperator struct {
    norm.Operator
    // Add custom fields
}

func (op *CustomOperator) Insert(ctx context.Context, conn any, query string, args ...any) (int64, error) {
    // Custom insert logic
    // ...
    return op.Operator.Insert(ctx, conn, query, args...)
}
```

## Best Practices

1. **Use struct tags:** Always define `db` tags on your struct fields for proper mapping.
2. **Validate input:** Check for errors on all database operations.
3. **Use transactions:** For operations that require atomicity.
4. **Use Reset:** Call `Reset()` before reusing a controller for a new query chain.

## Testing

For information about testing and benchmarking, see [README_TEST.md](README_TEST.md).

## License

[MIT License](LICENSE)
