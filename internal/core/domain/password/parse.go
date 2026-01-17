package password

import (
	"regexp"
)

var passwordRegEx = regexp.MustCompile("^[a-zA-Z0-9#@!-]{3,19}$")

// Parse parses the string value and returns a password if the value complies
// with the rules for a password.
func Parse(value string) (Password, error) {
	if !passwordRegEx.MatchString(value) {
		return Password{}, errInvalidPass
	}

	return Password{value}, nil
}

// MustParse parses the string value and returns a password if the value
// complies with the rules for a password. If an error occurs the function panics.
func MustParse(value string) Password {
	password, err := Parse(value)
	if err != nil {
		panic(err)
	}

	return password
}

func ParseConfirm(pass string, confirm string) (Password, error) {
	p, err := Parse(pass)
	if err != nil {
		return Password{}, err
	}

	if pass != confirm {
		return Password{}, errPassNoMatch
	}

	return p, nil
}

func ParseConfirmPointers(pass *string, confirm *string) (Password, error) {
	if pass == nil || confirm == nil {
		return Password{}, errPassNoMatch
	}

	return ParseConfirm(*pass, *confirm)
}
