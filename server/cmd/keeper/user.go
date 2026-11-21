package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/term"
)

var errUsage = errors.New("usage: keeper user add|passwd|del <login>")

type userManager interface {
	AddUser(ctx context.Context, login, password string) error
	SetPassword(ctx context.Context, login, password string) error
	DeleteUser(ctx context.Context, login string) error
}

func runUserCommand(ctx context.Context, users userManager, args []string) error {
	if len(args) != 3 || args[0] != "user" {
		return errUsage
	}
	cmd, login := args[1], args[2]
	if cmd == "del" {
		return users.DeleteUser(ctx, login)
	}
	if cmd != "add" && cmd != "passwd" {
		return errUsage
	}
	password, err := readPassword()
	if err != nil {
		return fmt.Errorf("failed to read password: %w", err)
	}
	if cmd == "add" {
		return users.AddUser(ctx, login, password)
	}
	return users.SetPassword(ctx, login, password)
}

func readPassword() (string, error) {
	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) {
		line, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil && !errors.Is(err, io.EOF) {
			return "", err
		}
		return strings.TrimSpace(line), nil
	}
	fmt.Print("Password: ")
	password, err := term.ReadPassword(fd)
	fmt.Println()
	return string(password), err
}
