package main

import (
	"time"
	"fmt"
)

func main() {
	now := time.Now()
	fmt.Println(now.Local().Unix())
	fmt.Println(now.Unix())
}
