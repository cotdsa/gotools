package main

import "github.com/segmentio/go-prompt"

func get_user_password_interactive() (string, string, error) {
	var vault_username string
	var vault_password string

	if vault_username == "" {
		vault_username = prompt.String("Enter Username (VAULT_USER)")
	}

	if vault_password == "" {
		vault_password = prompt.PasswordMasked("(%s)Enter Password", vault_username)
	}

	return vault_username, vault_password, nil
}
