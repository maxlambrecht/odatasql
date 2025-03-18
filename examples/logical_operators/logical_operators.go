package main

import (
	"fmt"
	"github.com/maxlambrecht/odatasql"
	"log"
)

func main() {
	sql, err := odatasql.Convert("age ge 30 or (status eq 'active' and premium eq true)")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(sql)
}
