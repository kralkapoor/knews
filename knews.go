package main

import (
	"errors"
	"fmt"
	"log"
)

func main() {
	fmt.Println("What's going on out there?")
	message, err := hello("")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(message)
}

func hello(name string) (string, error) {
	if name == "" {
		return "", errors.New("empty name")
	}

	message := fmt.Sprintf("Hi, %v. Welcome!", name)
	return message, nil
}
