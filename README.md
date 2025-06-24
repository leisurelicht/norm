# norm - Go SQL Query Builder and ORM

`norm` is a lightweight, flexible SQL query builder and ORM for Go applications. It provides a fluent interface for building and executing SQL queries with minimal boilerplate code.

## Features

- **Chainable Query Building**: Intuitive fluent API for building complex queries
- **Rich Filter Operators**: Comprehensive set of operators for complex query conditions
- **Multiple Database Support**: MySQL, ClickHouse with extensible operator system
- **CRUD Operations**: Simple and efficient Create, Read, Update, Delete operations
- **Automatic Struct Mapping**: Struct-to-table mapping with customizable `db` tags
- **Transaction Support**: Built-in transaction handling
- **Pagination Support**: Easy pagination with Limit and OrderBy
- **Aggregation**: GroupBy, Having clauses for complex queries
- **Bulk Operations**: Efficient bulk insert and update operations

## Installation

```bash
go get github.com/leisurelicht/norm
```

## Quick Start

### Define Your Model

```go
package main

import (
    "context"
    "time"
    
    "github.com/leisurelicht/norm"
    "github.com/zeromicro/go-zero/core/stores/sqlx"
    go_zero "github.com/leisurelicht/norm/operator/mysql/go-zero"
)

// User model with db tags
type User struct {
    ID          int64     `db:"id"`
    Name        string    `db:"name"`
    Email       string    `db:"email"`
    Age         int       `db:"age"`
    IsActive    bool      `db:"is_active"`
    CreatedAt   time.Time `db:"created_at"`
    UpdatedAt   time.Time `db:"updated_at"`
    IsDeleted   bool      `db:"is_deleted"`
}
```

### Initialize Controller

```go
func main() {
    // Initialize database connection
    db := sqlx.NewMysql("user:password@tcp(localhost:3306)/database")
    
    // Create controller with go-zero operator
    userController := norm.NewController(db, go_zero.NewOperator(), User{})
    
    ctx := context.Background()
    
    // Now you can use the controller
    users, err := userController(ctx).Filter(norm.Cond{"is_active": true}).FindAll()
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Found %d active users\n", len(users))
}
```

## Core Concepts

### 1. Conditions (Cond, AND, OR)

```go
// Basic condition
userController(ctx).Filter(norm.Cond{"name": "John"})

// Multiple conditions with AND (default)
userController(ctx).Filter(
    norm.Cond{"age__gte": 18},
    norm.AND{"is_active": true},
)

// OR conditions
userController(ctx).Filter(
    norm.Cond{"name": "John"},
    norm.OR{"email__contains": "gmail"},
)

// Complex nested conditions
userController(ctx).Filter(
    norm.Cond{"age__between": []int{18, 65}},
    norm.AND{"is_active": true},
    norm.OR{"name__startswith": "J"},
)
```

### 2. Filter Operators

| Operator | Description | Example |
|----------|-------------|---------|
| `exact` (default) | Exact match | `{"name": "John"}` |
| `exclude` | Not equal | `{"name__exclude": "John"}` |
| `iexact` | Case-insensitive exact | `{"name__iexact": "john"}` |
| `gt` | Greater than | `{"age__gt": 18}` |
| `gte` | Greater than or equal | `{"age__gte": 18}` |
| `lt` | Less than | `{"age__lt": 65}` |
| `lte` | Less than or equal | `{"age__lte": 65}` |
| `in` | In list | `{"id__in": []int{1,2,3}}` |
| `not_in` | Not in list | `{"id__not_in": []int{1,2,3}}` |
| `between` | Between values | `{"age__between": []int{18,65}}` |
| `not_between` | Not between values | `{"age__not_between": []int{18,65}}` |
| `contains` | String contains | `{"name__contains": "oh"}` |
| `not_contains` | String does not contain | `{"name__not_contains": "oh"}` |
| `icontains` | Case-insensitive contains | `{"name__icontains": "OH"}` |
| `not_icontains` | Case-insensitive not contains | `{"name__not_icontains": "OH"}` |
| `startswith` | String starts with | `{"name__startswith": "Jo"}` |
| `not_startswith` | String does not start with | `{"name__not_startswith": "Jo"}` |
| `istartswith` | Case-insensitive starts with | `{"name__istartswith": "jo"}` |
| `not_istartswith` | Case-insensitive not starts with | `{"name__not_istartswith": "jo"}` |
| `endswith` | String ends with | `{"name__endswith": "hn"}` |
| `not_endswith` | String does not end with | `{"name__not_endswith": "hn"}` |
| `iendswith` | Case-insensitive ends with | `{"name__iendswith": "HN"}` |
| `not_iendswith` | Case-insensitive not ends with | `{"name__not_iendswith": "HN"}` |
| `len` | String/field length | `{"name__len": 4}` |

### 3. Special Condition Features

```go
// Using SortKey to control field order in SQL generation
userController(ctx).Filter(norm.Cond{
    norm.SortKey: []string{"name", "email"},
    "name": "John",
    "email": "john@example.com",
})

// Using OR prefix (| ) to make specific field OR within same condition
userController(ctx).Filter(norm.Cond{
    "name": "John",
    "| email": "john@example.com", // name = 'John' OR email = 'john@example.com'
})

// EachOR - Convert all fields in condition to OR
userController(ctx).Filter(norm.EachOR(norm.Cond{
    "name": "John",
    "email": "john@example.com", // name = 'John' OR email = 'john@example.com'
}))
```

## CRUD Operations

### Create

```go
// Create single record
id, err := userController(ctx).Create(map[string]any{
    "name": "John Doe",
    "email": "john@example.com",
    "age": 25,
    "is_active": true,
})

// Create from struct
user := User{
    Name: "Jane Doe",
    Email: "jane@example.com",
    Age: 28,
    IsActive: true,
}
id, err := userController(ctx).Create(&user)

// Bulk create
users := []map[string]any{
    {"name": "User1", "email": "user1@example.com", "age": 20},
    {"name": "User2", "email": "user2@example.com", "age": 25},
}
count, err := userController(ctx).Create(users)
```

### Read

```go
// Find single record
user, err := userController(ctx).Filter(norm.Cond{"id": 1}).FindOne()

// Find into struct
var user User
err := userController(ctx).Filter(norm.Cond{"id": 1}).FindOneModel(&user)

// Find all records
users, err := userController(ctx).Filter(norm.Cond{"is_active": true}).FindAll()

// Find all into struct slice
var users []User
err := userController(ctx).Filter(norm.Cond{"is_active": true}).FindAllModel(&users)

// Count records
count, err := userController(ctx).Filter(norm.Cond{"is_active": true}).Count()

// Check existence
exists, err := userController(ctx).Filter(norm.Cond{"email": "john@example.com"}).Exist()
```

### Update

```go
// Update records
count, err := userController(ctx).
    Filter(norm.Cond{"id": 1}).
    Update(map[string]any{"name": "Updated Name"})

// Update multiple records
count, err := userController(ctx).
    Filter(norm.Cond{"is_active": false}).
    Update(map[string]any{"is_active": true})
```

### Delete

```go
// Soft delete (sets is_deleted = true)
count, err := userController(ctx).Filter(norm.Cond{"id": 1}).Delete()

// Hard delete (removes from database)
count, err := userController(ctx).Filter(norm.Cond{"id": 1}).Remove()
```

## Advanced Operations

### Pagination and Ordering

```go
// Order by single field
users, err := userController(ctx).
    OrderBy("name").
    FindAll()

// Order by multiple fields
users, err := userController(ctx).
    OrderBy([]string{"age", "-created_at"}). // ASC, DESC
    FindAll()

// Pagination (requires OrderBy)
users, err := userController(ctx).
    OrderBy("id").
    Limit(10, 1). // page size 10, page 1
    FindAll()
```

### Selection and Grouping

```go
// Select specific columns
users, err := userController(ctx).
    Select([]string{"name", "email"}).
    FindAll()

// Select with string
users, err := userController(ctx).
    Select("name, email, age").
    FindAll()

// Group by with aggregation
type AgeGroup struct {
    Age   int `db:"age"`
    Count int `db:"count"`
}

var ageGroups []AgeGroup
err := userController(ctx).
    Select("age, COUNT(*) as count").
    GroupBy("age").
    FindAllModel(&ageGroups)

// Group by with having
err := userController(ctx).
    Select("age, COUNT(*) as count").
    GroupBy("age").
    Having("COUNT(*) > ?", 5).
    FindAllModel(&ageGroups)
```

### Exclude Conditions

```go
// Exclude is the opposite of Filter
users, err := userController(ctx).
    Exclude(norm.Cond{"is_active": false}).
    FindAll()

// Equivalent to
users, err := userController(ctx).
    Filter(norm.Cond{"is_active__exclude": false}).
    FindAll()
```

### Raw SQL Conditions

```go
// Use Where for raw SQL
users, err := userController(ctx).
    Where("age > ? AND name LIKE ?", 18, "%John%").
    FindAll()

// Note: Where and Filter/Exclude cannot be used together
```

### Advanced CRUD Operations

```go
// Get or Create - Uses data map fields as filter conditions automatically
user, err := userController(ctx).GetOrCreate(map[string]any{
    "email": "new@example.com",
    "name": "New User",
    "age": 25,
})

// Create or Update - Requires existing filter conditions
created, count, err := userController(ctx).
    Filter(norm.Cond{"email": "user@example.com"}).
    CreateOrUpdate(map[string]any{
        "name": "Updated Name",
        "age": 30,
    })

// Create if not exists -Uses data map fields as filter conditions automatically
id, created, err := userController(ctx).CreateIfNotExist(map[string]any{
    "email": "unique@example.com",
    "name": "Unique User",
})

// List with count and data
total, users, err := userController(ctx).
    Filter(norm.Cond{"is_active": true}).
    OrderBy("name").
    List()
```

## Database Support

### MySQL with go-zero

```go
import (
    "github.com/zeromicro/go-zero/core/stores/sqlx"
    go_zero "github.com/leisurelicht/norm/operator/mysql/go-zero"
)

db := sqlx.NewMysql("connection_string")
controller := norm.NewController(db, go_zero.NewOperator(), YourModel{})
```

### MySQL with sqlx

```go
import (
    sqlx_op "github.com/leisurelicht/norm/operator/mysql/sqlx"
)

// Use with database/sql or jmoiron/sqlx
controller := norm.NewController(db, sqlx_op.NewOperator(), YourModel{})
```

### ClickHouse

```go
import (
    clickhouse_op "github.com/leisurelicht/norm/operator/clickhouse"
)

controller := norm.NewController(db, clickhouse_op.NewOperator(), YourModel{})
```

## Error Handling

```go
import (
    "errors"
    "github.com/leisurelicht/norm"
)

// Check for specific errors
user, err := userController(ctx).Filter(norm.Cond{"id": 999}).FindOne()
if errors.Is(err, norm.ErrNotFound) {
    // Handle not found
    fmt.Println("User not found")
} else if errors.Is(err, norm.ErrDuplicateKey) {
    // Handle duplicate key
    fmt.Println("Duplicate key error")
} else if err != nil {
    // Handle other errors
    fmt.Printf("Error: %v\n", err)
}

// Note: FindOne returns empty map[string]any{} when no record is found (no error)
// ErrNotFound is only returned by FindOneModel and FindAllModel when used with struct pointers
```

## Struct Tags

Use `db` tags to map struct fields to database columns:

```go
type User struct {
    ID          int64      `db:"id"`
    Name        string     `db:"name"`
    Email       string     `db:"email"`
    IgnoreField string     `db:"-"`           // Ignored field
    CustomName  string     `db:"custom_col"`  // Custom column name
    WithOptions string     `db:"col,type=varchar,length=100"` // With options
}
```

## Best Practices

1. **Always use context**: Pass context to controller functions
2. **Handle errors**: Check and handle errors appropriately
3. **Use transactions**: For operations that need atomicity
4. **Validate input**: Validate data before database operations
5. **Use struct models**: Prefer struct models over maps for type safety
6. **Order dependencies**: Call OrderBy before Limit for pagination
7. **Filter vs Where**: Cannot use Filter/Exclude and Where in the same query
8. **Method restrictions**: Some methods like GroupBy, Select are not supported for certain operations (Create, Update, Delete)

## Method Restrictions

Some methods have restrictions when used with certain operations:

### Create Operations

- **Not supported**: Filter, Exclude, Where, Select, OrderBy, GroupBy, Having, Limit

### Update/Delete/Remove Operations  

- **Update**: Not supported - Select, GroupBy, Having
- **Delete**: Not supported - GroupBy, Select, OrderBy
- **Remove**: Not supported - Select, GroupBy, Having

### Query Operations

- **FindOne/FindAll**: Not supported - Select and GroupBy together, Having without GroupBy
- **FindOneModel/FindAllModel**: Supports all query methods
- **Exist**: Not supported - GroupBy, Select
- **Count**: Supports all filter methods

### Advanced Operations

- **GetOrCreate**: Not supported - Select, GroupBy, Having. Uses data map fields as filter conditions.
- **CreateOrUpdate**: Not supported - Select, GroupBy, Having. Requires existing filter conditions.
- **CreateIfNotExist**: Not supported - Select, GroupBy, Having. Automatically adds filter from data map.

### Important Notes

1. **Limit dependency**: Limit can only be used after OrderBy
2. **Where vs Filter/Exclude**: Cannot use Where with Filter or Exclude in the same query
3. **Column validation**: Select, OrderBy, and GroupBy validate column names against model fields
4. **FindOne vs FindOneModel**: FindOne returns empty map when not found, FindOneModel returns ErrNotFound

## Examples

### Complete CRUD Example

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/leisurelicht/norm"
    "github.com/zeromicro/go-zero/core/stores/sqlx"
    go_zero "github.com/leisurelicht/norm/operator/mysql/go-zero"
)

type User struct {
    ID        int64     `db:"id"`
    Name      string    `db:"name"`
    Email     string    `db:"email"`
    Age       int       `db:"age"`
    IsActive  bool      `db:"is_active"`
    CreatedAt time.Time `db:"created_at"`
    IsDeleted bool      `db:"is_deleted"`
}

func main() {
    // Initialize
    db := sqlx.NewMysql("user:pass@tcp(localhost:3306)/testdb")
    userCtl := norm.NewController(db, go_zero.NewOperator(), User{})
    ctx := context.Background()

    // Create user
    id, err := userCtl(ctx).Create(map[string]any{
        "name":       "John Doe",
        "email":      "john@example.com",
        "age":        25,
        "is_active":  true,
        "created_at": time.Now(),
    })
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Created user with ID: %d\n", id)

    // Find user
    var user User
    err = userCtl(ctx).Filter(norm.Cond{"id": id}).FindOneModel(&user)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Found user: %+v\n", user)

    // Update user
    count, err := userCtl(ctx).
        Filter(norm.Cond{"id": id}).
        Update(map[string]any{"age": 26})
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Updated %d records\n", count)

    // Find active users with pagination
    var users []User
    err = userCtl(ctx).
        Filter(norm.Cond{"is_active": true}).
        OrderBy("name").
        Limit(10, 1).
        FindAllModel(&users)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Found %d active users\n", len(users))

    // Soft delete
    count, err = userCtl(ctx).Filter(norm.Cond{"id": id}).Delete()
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Soft deleted %d records\n", count)
}
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License.
