package main

import (
	"fmt"
	"github.com/maxlambrecht/odatasql"
	"log"
)

// Demonstrates how operator precedence and explicit parentheses are preserved.
func main() {
	sql, err := odatasql.FilterToSQL("not (age gt 25 and status eq 'active') or premium eq true")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(sql)
}
