# ODataSQL - Convert OData Filters to SQL

**ODataSQL** is a Go library that converts **OData filter expressions into SQL WHERE clauses** while preserving operator
precedence.

## 🔹 Usage

```go
func main() {
    sql, err := odatasql.Convert("name eq 'Alice' and age gt 30")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(sql) // Output: name = 'Alice' AND age > 30
}
```

## 🛠 Supported Operators

| OData | SQL   | Example OData                      | SQL Output                       |
|-------|-------|------------------------------------|----------------------------------|
| eq    | `=`   | `firtsName eq 'Bob'`               | `first_name = 'Bob'`             |
| ne    | `!=`  | `status ne 'active'`               | `status != 'active'`             |
| gt    | `>`   | `age gt 18`                        | `age > 18`                       |
| ge    | `>=`  | `height ge 170`                    | `height >= 170`                  |
| lt    | `<`   | `score lt 50`                      | `score < 50`                     |
| le    | `<=`  | `price le 99.99`                   | `price <= 99.99`                 |
| and   | `AND` | `age gt 18 and status eq 'active'` | `age > 18 AND status = 'active'` |
| or    | `OR`  | `age lt 18 or premium eq true`     | `age < 18 OR premium = true`     |
| not   | `NOT` | `not age gt 18`                    | `NOT age > 18`                   |
| in    | `IN`  | `color in ('red', 'blue')`         | `color IN ('red', 'blue')`       |

## 📂 Running Examples

```sh
go run examples/basic/basic.go
go run examples/in_operator/in_operator.go
go run examples/logical_operators/logical_operators.go
go run examples/precedence/precedence.go
```
