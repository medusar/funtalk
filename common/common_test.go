package common

import (
	"fmt"
	"testing"
)

func TestNewId(t *testing.T) {
	fmt.Println(NewId())
	fmt.Println(NewId().ID())
}

func TestNewIdStr(t *testing.T) {
	fmt.Println(NewIdStr())
}
