package helpers

import (
	"errors"
)

func RemoveFromList(s []string, i int) ([]string, error) {
	if i < 0 || i > len(s) - 1 {
		return []string{}, errors.New("Index provided is out of bounds!")
	} 
    s[i] = s[len(s)-1]
    return s[:len(s)-1], nil
}