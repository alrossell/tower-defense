package main

import (
	"errors"
	"fmt"

	// "math/rand/v2"
	"strings"
	// "text/template"
	// "net"
	"reflect"
	"time"

	"example.com/tower-defense/cord"
	"example.com/tower-defense/player"
)

const boardSize int = 14

var board [boardSize][boardSize]int
var round = 0

var startCord *cord.Cord = nil 
var endCord *cord.Cord = nil

var numberOfPlayers = 0;
var playerList []*player.Player

var gameMapPathCords []cord.Cord;

var previousDisplayLength = boardSize + 2

// --------------------------------
// |        Print Functions       |
// --------------------------------

func printTile(row int, col int) {

    for _, currentPlayer := range playerList {
        if currentPlayer.CurrentCord.Row == row && currentPlayer.CurrentCord.Col == col {
     	    fmt.Printf("\033[48;5;2m  \033[0m")
            return;
        }
    }

    if board[row][col] == 1 {
	    fmt.Printf("\033[48;5;1m  \033[0m")
    } else {
	    fmt.Printf("\033[48;5;255m  \033[0m")
    }
}

func initPrintBoard() {
    fmt.Print("Board: Round 0\n")

    for Row := (boardSize - 1); Row >= 0; Row-- {
        for Col := 0; Col < boardSize; Col++ {
            printTile(Row, Col)
        }
        fmt.Println()
    }

    fmt.Print("Number of Player: 0\n")
}

func printBoard() {
    // Resets the board
    fmt.Printf("\033[%dF2K\r", previousDisplayLength)

    previousDisplayLength = boardSize + 2 + len(playerList)

    fmt.Printf("Board: Round %d\n", round)

    for Row := (boardSize - 1); Row >= 0; Row-- {
        for Col := 0; Col < boardSize; Col++ {
            printTile(Row, Col)
        }
        fmt.Println()
    }

    fmt.Printf("Number of Player: %d\n", len(playerList))
    for _, currentPlayer := range playerList {
        fmt.Printf("Row: %d, Col: %d\n", currentPlayer.CurrentCord.Row, currentPlayer.CurrentCord.Col)
    } 
}

// ---------------------------------------
// |        Game Initialization          |
// ---------------------------------------

func isInGameMapPathCords(currTile *cord.Cord) (bool) {
    for _, value := range gameMapPathCords {
        if value == *currTile {
            return true
        }
    }

    return false
}

func processTile(row int, col int, index int, rawGameMapPath string) (error) {
    if(rawGameMapPath[index] == 'S') {
        if startCord != nil {
            return errors.New("Duplicated start tiles defined")
        }

        startCord =  &cord.Cord { Row: row, Col: col } 
        board[row][col] = 1
    } else if (rawGameMapPath[index] == 'E') {
        if endCord != nil {
            return errors.New("Duplicated end tiles defined")
        }

        endCord = &cord.Cord { Row: row, Col: col } 
        board[row][col] = 1
    } else if(rawGameMapPath[index] == '1') {
        board[row][col] = 1
    } else if (rawGameMapPath[index] == '0') {
        board[row][col] = 0
    }

    return nil
}

func getGamePath() {
    currTile := startCord
    gameMapPathCords = append(gameMapPathCords, *currTile)

    nextTiles := []cord.Cord { { Row: 0, Col: 1}, { Row: 1, Col: 0 }, { Row: 0, Col: -1 }, { Row: -1, Col: 0 } }

    for !reflect.DeepEqual(currTile, endCord) {
        for _, nextTile := range nextTiles {    
            currTile.Row += nextTile.Row
            currTile.Col += nextTile.Col

            if(currTile.Row >= boardSize || currTile.Row < 0 || currTile.Col >= boardSize || currTile.Col < 0) {
                currTile.Row -= nextTile.Row
                currTile.Col -= nextTile.Col
                continue
            }

            if(board[currTile.Row][currTile.Col] == 1 && !isInGameMapPathCords(currTile)) {
                gameMapPathCords = append(gameMapPathCords, *currTile)
                break
            }
            currTile.Row -= nextTile.Row
            currTile.Col -= nextTile.Col
        }
    }
}

func initBoard() (error) {
    gameMapPath := strings.TrimSpace(`
    0 0 0 0 0 0 0 0 0 0 0 0 0 0 
    S 1 1 1 1 1 0 0 0 0 0 0 0 0 
    0 0 0 0 0 1 0 0 0 0 0 0 0 0
    0 1 1 1 1 1 0 0 0 0 0 0 0 0 
    0 1 0 0 0 0 0 0 0 0 0 0 0 0
    0 1 0 0 0 0 0 0 0 0 0 0 0 0
    0 1 1 1 1 1 1 1 1 1 1 0 0 0
    0 0 0 0 0 0 0 0 0 0 1 0 0 0
    0 0 0 0 0 0 0 0 0 0 1 0 0 0 
    0 1 1 1 1 1 1 1 1 1 1 0 0 0
    0 1 0 0 0 0 0 0 0 0 0 0 0 0
    0 1 1 1 1 1 1 1 1 1 1 1 1 E
    0 0 0 0 0 0 0 0 0 0 0 0 0 0
    0 0 0 0 0 0 0 0 0 0 0 0 0 0 
    `)

    rawGameMapPath := strings.Join(strings.Fields(gameMapPath), "")

    if len(rawGameMapPath) != (boardSize * boardSize) {
        return errors.New("Game map size does not match board size")
    }

    for Row := (boardSize - 1); Row >= 0; Row-- {
        for Col := 0; Col < boardSize; Col++ {
            index := Col + ((boardSize - Row - 1) * boardSize)
            processTile(Row, Col, index, rawGameMapPath)
        }
    }

    getGamePath()

    return nil
}

// --------------------------------------
// |        Game Logic Functions        |
// --------------------------------------

func updateBoard() {
    playerLen := len(playerList)
    for index := 0; index < playerLen; {
        if playerList[index].CurrentCord == *startCord {
            playerList[index] = playerList[playerLen-1]
            playerLen-- 
        } else {
            index++
        }
    }
    playerList = playerList[:playerLen] 

    for _, currentPlayer := range playerList {
        currentPlayer.CurrentCordIndex++
        currentPlayer.CurrentCord = gameMapPathCords[currentPlayer.CurrentCordIndex]
    }

    round++
}

func gameMainLoop() {
    initPrintBoard()
   
    for {
        updateBoard()
      
        if len(playerList) < 5 {
            newPlayer := player.Player {
                CurrentCord: cord.Cord{
                    Row: gameMapPathCords[0].Row,
                    Col: gameMapPathCords[0].Col,
                },
                CurrentCordIndex: 0,
            }

            playerList = append(playerList, &newPlayer)
        }

        time.Sleep(1000 * time.Millisecond)
        printBoard()

    }
}

func main() {
    initBoard()
    gameMainLoop()
}
