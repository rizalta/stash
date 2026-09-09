package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/rizalta/stash/internal/vault"
	"golang.org/x/term"
)

var stdinReader = bufio.NewReader(os.Stdin)

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

func openVaultFromPrompt() (*vault.Vault, error) {
	password, err := promptPassword("master password: ")
	if err != nil {
		return nil, err
	}

	v, err := vault.Open(vaultPath, password)
	if err != nil {
		if errors.Is(err, vault.ErrWrongPassword) {
			return nil, errors.New("incorrect password")
		}
		return nil, fmt.Errorf("opening vault: %w", err)
	}
	fmt.Println()

	return v, nil
}

func promptField(prompt string) (string, error) {
	fmt.Print(prompt)
	val, err := stdinReader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("reading field: %w", err)
	}

	return strings.TrimSpace(val), nil
}

func promptFieldWithDefault(prompt string, current string) (string, error) {
	val, err := promptField(fmt.Sprintf("%s [%s]: ", prompt, current))
	if err != nil {
		return "", err
	}

	if val == "" {
		return current, nil
	}

	return val, nil
}
