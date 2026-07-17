//go:build !js && !wasm

package stdlib

import (
	_ "github.com/mat-dgruber/Harpia/stdlib/bd"
	_ "github.com/mat-dgruber/Harpia/stdlib/soquete"
)
