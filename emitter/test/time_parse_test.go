package test

import (
	"fmt"
	"testing"
	"time"
)

func TestTimeParse(t *testing.T) {
	parse, err := time.Parse("01-02 15-04-05", "09-27 22-22-04")
	parse = parse.AddDate(2022, 0, 0)
	fmt.Println(parse, err)
}
