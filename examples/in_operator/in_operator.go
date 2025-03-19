package main

import (
	"fmt"
	"github.com/maxlambrecht/odatasql"
	"log"
)

func main() {
	sql, err := odatasql.FilterToSQL("color in ('red', 'blue') and category eq 'electronics'")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(sql)
}
