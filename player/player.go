package player 

import (
    "example.com/tower-defense/cord"
)

type Player struct {
    CurrentCord cord.Cord
    CurrentCordIndex int
    Health int
}

