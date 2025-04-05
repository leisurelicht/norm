# norm - Go SQL Query Builder and ORM

`norm` is a lightweight, flexible SQL query builder and ORM for Go applications. It provides a fluent interface for building and executing SQL queries with minimal boilerplate code.

## Features

- Chainable query building API with intuitive methods
- Rich set of filter operators for complex query conditions
- Simple and efficient CRUD operations
- Automatic struct-to-table mapping with customizable tags
- Transaction support
- Customizable query operators for different database engines
- Optimized query generation with minimal allocations
- Built-in pagination support

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
    "time"
    
    "github.com/username/norm"
    "github.com/zeromicro/go-zero/core/stores/sqlx"
)

// Define your model
type User struct {
    Id        int64     `db:"id"`
    Username  string    `db:"username"`
    Email     string    `db:"email"`
    CreatedAt time.Time `db:"created_at"`
    IsDeleted bool      `db:"is_deleted"`
}

func main() {
    // Connect to database
    conn := sqlx.NewMysql("user:password@tcp(localhost:3306)/dbname")
    
    // Create controller
    userCtrl := norm.NewController(conn, mysqlOp.NewOperator(), User{})
    
    // Use controller with context
    ctx := context.Background()
    ctrl := userCtrl(ctx)
    
    // Create a new user
    user := User{
        Username:  "johndoe",
        Email:     "john@example.com",
        CreatedAt: time.Now(),
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

QuerySet provides a fluent API for building SQL queries with conditions, ordering, limits, etc. It handles the construction of SQL statements and parameter binding.

### Operator

Operators handle the actual database operations, allowing for customization of how queries are executed for different database systems.

## API Reference

### Creating a Controller

```go
// Create a controller factory for a specific model
userCtrl := norm.NewController(conn, mysqlOp.NewOperator(), User{})

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
    "created_at": time.Now(),
    "is_deleted": false,
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

// Update with exclude - update all fields except specified ones
num, err := ctrl.Exclude(norm.AND{"updated_at": time.Now()}).Update(userData)

// Modify - automatically excludes the filter fields from update data
num, err := ctrl.Modify(map[string]any{"status": "active", "id": 1}) // Won't update the "id" field
```

#### Delete

```go
// Soft delete (setting is_deleted to true)
num, err := ctrl.Filter(norm.AND{"id": 1}).Delete()

// Hard delete (removing from database)
num, err := ctrl.Filter(norm.AND{"id": 1}).Remove()
```

### Advanced Query Operations

#### Get or Create

```go
// Try to find a record, create it if not exist
result, err := ctrl.GetOrCreate(map[string]any{
    "username": "johndoe",
    "email": "john@example.com",
})
```

#### Create or Update

```go
// Create a record if it doesn't exist, otherwise update it
created, id, err := ctrl.CreateOrUpdate(
    map[string]any{"email": "updated@example.com"}, // data to insert/update
    norm.AND{"username": "johndoe"}, // filter criteria
)
```

#### Create If Not Exist

```go
// Only create if record doesn't exist
id, created, err := ctrl.CreateIfNotExist(map[string]any{
    "username": "unique_user",
    "email": "unique@example.com",
})
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
| len | String length equals | `"username__len": 8` |
| in | In a list | `"status__in": []string{"active", "pending"}` |
| between | Between two values | `"age__between": []int{18, 65}` |
| contains | Contains string (case sensitive) | `"username__contains": "john"` |
| icontains | Case-insensitive contains | `"username__icontains": "john"` |
| startswith | Starts with string (case sensitive) | `"username__startswith": "j"` |
| istartswith | Case-insensitive starts with | `"username__istartswith": "j"` |
| endswith | Ends with string (case sensitive) | `"email__endswith": ".com"` |
| iendswith | Case-insensitive ends with | `"email__iendswith": ".COM"` |

#### Using NOT Operators

All operators can be negated by prefixing them with `not_`:

```go
// Find users whose username doesn't contain "admin"
result, err := ctrl.Filter(norm.AND{"username__not_contains": "admin"}).FindAll()

// Find users who aren't in the specified ID list
result, err := ctrl.Filter(norm.AND{"id__not_in": []int64{1, 2, 3}}).FindAll()
```

#### Using AND/OR Conditions

```go
// Complex filtering with AND/OR combinations
result, err := ctrl.Filter(
    "AND", // Default conjunction is AND
    norm.AND{"username__startswith": "john", "is_deleted": false},
    norm.OR{"email__contains": "@example.com", "email__contains": "@company.com"},
).FindAll()

// Converting map keys to use OR logic
orConditions := norm.EachOR(norm.AND{
    "username__contains": "john",
    "email__contains": "john",
}) // Both conditions will use OR instead of AND
result, err := ctrl.Filter(orConditions).FindAll()

// Convert a single key to use OR logic
result, err := ctrl.Filter(norm.AND{
    "username": "john",
    norm.ToOR("email"): "john@example.com",
}).FindAll() // username=john OR email=john@example.com
```

### Working with Direct SQL

```go
// Use raw WHERE condition
results, err := ctrl.Where("username = ? OR email LIKE ?", "john", "%@example.com%").FindAll()
```

### Ordering, Limit, and Pagination

```go
// Order by fields (prefix with - for descending)
results, err := ctrl.OrderBy([]string{"created_at", "-username"}).FindAll()

// String-based ordering
results, err := ctrl.OrderBy("created_at ASC, username DESC").FindAll()

// Limit and pagination (pageSize, pageNumber)
results, err := ctrl.Limit(10, 1).FindAll() // 10 records per page, first page

// Combined example with filtering, ordering, and pagination
results, err := ctrl.Filter(norm.AND{"is_deleted": false})
                   .OrderBy([]string{"created_at", "-username"})
                   .Limit(10, 2) // Second page
                   .FindAll()
```

### Selecting Specific Columns

```go
// Select specific columns
results, err := ctrl.Select([]string{"id", "username", "email"}).FindAll()

// String-based select
results, err := ctrl.Select("id, username, COUNT(*) as user_count").FindAll()

// Select with joins
results, err := ctrl.Select("u.id, u.username, p.name as profile_name")
                   .Where("u.id = p.user_id")
                   .FindAll()
```

### Group By and Having

```go
// Group by with having clause
results, err := ctrl.GroupBy([]string{"status"})
                   .Having("COUNT(*) > ?", 5)
                   .FindAll()

// String-based grouping
results, err := ctrl.GroupBy("status, created_at")
                   .Having("COUNT(*) > ? AND MAX(score) > ?", 5, 80)
                   .FindAll()
```

### Transactions

```go
session := sqlx.NewSession(conn)
err := session.TransactCtx(context.Background(), func(ctx context.Context, s sqlx.Session) error {
    ctrl := userCtrl(ctx).WithSession(s)
    
    // Perform operations within transaction
    id, err := ctrl.Insert(map[string]any{"username": "transaction_user"})
    if err != nil {
        return err // Transaction will be rolled back
    }
    
    // Update another record in the same transaction
    _, err = ctrl.Reset().Filter(norm.AND{"id": id}).Update(map[string]any{
        "email": fmt.Sprintf("tx_%d@example.com", id)
    })
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

// Get column-to-column mapping (useful for lookups)
// Returns a map where keys are values from column1 and values are from column2
map, err := ctrl.GetC2CMap("id", "username")
// Result: map[1:"john" 2:"jane" 3:"bob"]

// List with pagination (returns total count and paginated data)
total, data, err := ctrl.Filter(norm.AND{"is_deleted": false})
                     .OrderBy([]string{"created_at"})
                     .Limit(10, 1)
                     .List()

// GetOrCreate finds a record matching the criteria or creates it if not found
result, err := ctrl.GetOrCreate(map[string]any{
    "username": "johndoe",
    "email": "john@example.com",
})

// CreateOrUpdate creates a record if it doesn't exist, otherwise updates it
// Returns: created (bool), id/affected rows (int64), error
created, id, err := ctrl.CreateOrUpdate(
    map[string]any{"email": "updated@example.com"}, // data to insert/update
    norm.AND{"username": "johndoe"}, // filter criteria
)

// CreateIfNotExist only creates a record if one matching the criteria doesn't exist
// Returns: id (int64), created (bool), error
id, created, err := ctrl.CreateIfNotExist(map[string]any{
    "username": "unique_user",
    "email": "unique@example.com",
})

// Modify - automatically excludes the filter fields from update data
// Useful for updating records while preventing certain fields from being changed
num, err := ctrl.Modify(map[string]any{
    "status": "active", 
    "id": 1, // The id field will be used as a filter and not updated
})
```

### Transaction with Multiple Operations

For more complex transactions that involve multiple operations:

```go
err := session.TransactCtx(context.Background(), func(ctx context.Context, s sqlx.Session) error {
    // Create controller with session
    ctrl := userCtrl(ctx).WithSession(s)
    
    // 1. Insert a new user
    userData := map[string]any{
        "username": "new_user",
        "email": "new@example.com",
        "created_at": time.Now(),
    }
    userId, err := ctrl.Insert(userData)
    if err != nil {
        return err
    }
    
    // 2. Create related user profile
    profileCtrl := profileCtrl(ctx).WithSession(s)
    profileId, err := profileCtrl.Insert(map[string]any{
        "user_id": userId,
        "display_name": "New User",
        "created_at": time.Now(),
    })
    if err != nil {
        return err
    }
    
    // 3. Update user with profile reference
    _, err = ctrl.Reset().Filter(norm.AND{"id": userId}).Update(map[string]any{
        "profile_id": profileId,
        "status": "active",
    })
    if err != nil {
        return err
    }
    
    // All operations succeeded - transaction will commit
    return nil
})

if err != nil {
    // Handle transaction error
    log.Printf("Transaction failed: %v", err)
}
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
    log.Printf("Executing insert query: %s with args: %v", query, args)
    // ...
    return op.Operator.Insert(ctx, conn, query, args...)
}
```

## Best Practices

1. **Use struct tags:** Always define `db` tags on your struct fields for proper mapping.
2. **Validate input:** Check for errors on all database operations.
3. **Use transactions:** For operations that require atomicity.
4. **Use Reset:** Call `Reset()` before reusing a controller for a new query chain.
5. **Avoid unnecessary queries:** Use methods like `Exist()` instead of retrieving and checking records.
6. **Leverage batch operations:** Use `BulkInsert` and similar methods for better performance.
7. **Monitor query performance:** Use the testing utilities to identify slow queries.

## Common Patterns

### Repository Pattern

```go
// UserRepository handles all database operations for users
type UserRepository struct {
    ctrl func(context.Context) norm.Controller
}

func NewUserRepository(conn sqlx.SqlConn) *UserRepository {
    return &UserRepository{
        ctrl: norm.NewController(conn, mysqlOp.NewOperator(), User{}),
    }
}

func (r *UserRepository) FindByUsername(ctx context.Context, username string) (*User, error) {
    var user User
    err := r.ctrl(ctx).Filter(norm.AND{"username": username}).FindOneModel(&user)
    if err != nil {
        return nil, err
    }
    return &user, nil
}

func (r *UserRepository) Create(ctx context.Context, user *User) (int64, error) {
    return r.ctrl(ctx).InsertModel(user)
}
```

### Working with Time Fields

```go
// Find records created in the last 24 hours
yesterday := time.Now().AddDate(0, 0, -1)
results, err := ctrl.Filter(norm.AND{"created_at__gte": yesterday}).FindAll()

// Find records between two dates
start := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
end := time.Date(2023, 12, 31, 23, 59, 59, 0, time.UTC)
results, err := ctrl.Filter(
    norm.AND{"created_at__between": []time.Time{start, end}}
).FindAll()
```

## Testing

For information about testing and benchmarking, see [README_TEST.md](README_TEST.md).

## License

[MIT License](LICENSE)
