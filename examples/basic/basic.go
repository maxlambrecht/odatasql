package main

import (
	"fmt"
	"log"

	"github.com/maxlambrecht/odatasql"
)

// Basic example of converting an OData filter into an SQL WHERE clause.
func main() {
	sql, err := odatasql.Convert("name eq 'Alice' and age gt 30")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(sql)
}
