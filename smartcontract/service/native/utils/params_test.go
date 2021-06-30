package utils

import (
	"fmt"
	"testing"
)

func TestFilmAddr(t *testing.T) {
	fmt.Println("FilmContractAddress", FilmContractAddress.ToBase58())
}
