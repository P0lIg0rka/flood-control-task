package main

import (
	"context"
	"fmt"
	floodcontrol "task/fc"
)

type res struct {
	user int64
	r    bool
	err  error
}

func main() {
	checker := floodcontrol.CreateFC(1, 2)
	ctx := context.Background()
	ch := make(chan res)
	for i := 0; i < 15; i++ {
		go func() {
			check, err := checker.Check(ctx, 1)
			ch <- res{
				user: 1,
				r:    check,
				err:  err,
			}
		}()
	}
	for i := 0; i < 15; i++ {
		r := <-ch
		if r.err != nil {
			fmt.Printf("Error: %s\n", r.err)
			continue
		}
		if r.r {
			fmt.Printf("User %d is flooding\n", r.user)
		} else {
			fmt.Printf("User %d is good\n", r.user)
		}
	}
	close(ch)
}
