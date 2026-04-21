package utils

import (
	"encoding/json"
	"fmt"
	"os"
)

func Dd(data interface{}) {
	b, _ := json.MarshalIndent(data, "", "  ")
	fmt.Println(string(b))
	os.Exit(1)
}

func Dump(data interface{}) {
	b, _ := json.MarshalIndent(data, "", "  ")
	fmt.Println(string(b))
}
