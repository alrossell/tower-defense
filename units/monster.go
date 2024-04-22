package units 

import (
    "example.com/tower-defense/cord"
)

type Monster struct {
    CurrentCord cord.Cord
    CurrentCordIndex int
    Health int
}

