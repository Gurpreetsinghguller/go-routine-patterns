package main

import (
	"fmt"

	"github.com/Gurpreetsinghguller/go-routine-patterns/fanInFanOut"
	"github.com/Gurpreetsinghguller/go-routine-patterns/workerpool"
)

func main() {
	fmt.Println(fanInFanOut.Manager())
	fmt.Println(workerpool.Manager())
}
