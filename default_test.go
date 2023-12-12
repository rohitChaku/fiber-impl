package main

import (
	"fmt"
	"testing"
)

func TestMapDefault1(t *testing.T) {
	var user *User
	err := MapDefault(&user)
	fmt.Print(err, "\n", user)
}

// unknown type -> need pointer
func TestMapDefault2(t *testing.T) {
	var user *User
	err := MapDefault(user)
	fmt.Print(err, "\n", user)
}

func TestMapDefault3(t *testing.T) {
	var user User
	err := MapDefault(&user)
	fmt.Print(err, "\n", user)
}

func TestMapFormDefault(t *testing.T) {
	var user *User
	err := MapFormDefault(&user)
	fmt.Print(err, "\n", user)
}

// unknown type -> need pointer
func TestMapFormDefault2(t *testing.T) {
	var user *User
	err := MapFormDefault(user)
	fmt.Print(err, "\n", user)
}

func TestMapFormDefault3(t *testing.T) {
	var user User
	err := MapFormDefault(&user)
	fmt.Print(err, "\n", user)
}
