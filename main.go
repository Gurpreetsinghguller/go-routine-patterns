package main

import (
	"fmt"

	"github.com/Gurpreetsinghguller/go-routine-patterns/fanInFanOut"
	timeoutselect "github.com/Gurpreetsinghguller/go-routine-patterns/timeoutSelect"
	"github.com/Gurpreetsinghguller/go-routine-patterns/workerpool"
)

func main() {
	fmt.Println(fanInFanOut.Manager())
	fmt.Println(workerpool.Manager())
	fmt.Println(timeoutselect.Manager())
}
