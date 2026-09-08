package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"

	"golang.org/x/term"
)

var ErrPasswordMismatch = errors.New("passwords do not match")

func promptPassword(prompt string) ([]byte, error) {
	fmt.Print(prompt)
	password, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	if err != nil {
		return nil, fmt.Errorf("reading password: %w", err)
	}

	return password, nil
}

func promptNewPassword() ([]byte, error) {
	password, err := promptPassword("new master password: ")
	if err != nil {
		return nil, err
	}

	confirm, err := promptPassword("confirm master password: ")
	if err != nil {
		return nil, err
	}

	if !bytes.Equal(password, confirm) {
		return nil, ErrPasswordMismatch
	}

	return password, nil
}
