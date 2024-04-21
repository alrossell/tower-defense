package main

import (
	"errors"
	"fmt"
    "log"
	"strings"
	"reflect"
	"time"
    "os"

	"example.com/tower-defense/cord"
	"example.com/tower-defense/player"

    "github.com/eiannone/keyboard"
)

const boardSize int = 14
const boardRowSize int = boardSize * 2 + 3

var board [boardSize][boardSize]int
var round = 0

var startCord *cord.Cord = nil 
var endCord *cord.Cord = nil

var numberOfPlayers = 0;
var monsterList []*player.Player

var gameMapPathCords []cord.Cord;

var userCursor *cord.Cord = &cord.Cord{ Row: 1, Col: 1 } 

var towerPlacement [boardSize][boardSize]int
var towerHitList [boardSize][boardSize]int

var playerHealth = 100

var tileChar = strings.Repeat(" ", 4)  
var nextCursorX = 0
var nextCursorY = 0

var rowOffset = 0

var orange = "200"
var blue = "3"
var green = "46"
var red = "1"
var white = "255"
var cyan = "6"

// --------------------------------
// |        Print Functions       |
// --------------------------------

func printBoardTile(row int, col int, tileColor string) {
    
    // Have to inverse the row because it is stored inverted
    row = boardSize - row - 1

    fmt.Printf("\033[%d;%dH", row * 2 + 1 + rowOffset, col * 4 + 1)
    fmt.Printf("\033[48;5;%vm%v\033[0m", tileColor, "    ")
    fmt.Println()
    fmt.Printf("\033[%d;%dH", row * 2 + 2 + rowOffset, col * 4 + 1)
    fmt.Printf("\033[48;5;%vm%v\033[0m", tileColor, "    ")
}

func clearScreen() {
    fmt.Print("\033[H")
    fmt.Print("\033[2J")
}

func printTile(row int, col int) {
    if towerPlacement[row][col] == 1 {
        printBoardTile(row, col, blue)
    } else if board[row][col] == 1 {
        printBoardTile(row, col, red)
    } else {
        printBoardTile(row, col, white)
    }

}

func printBuildTile(row int, col int) {
    if userCursor.Row == row && userCursor.Col == col {
        printBoardTile(row, col, green)
        return
    } 

    printTile(row, col)
}

func printActionTile(row int, col int) {
    for _, currentPlayer := range monsterList {
        if currentPlayer.CurrentCord.Row == row && currentPlayer.CurrentCord.Col == col {
            printBoardTile(row, col, cyan)
            return;
        }
    }

    printTile(row, col)
}

func printHeader() {
    clearScreen()

    fmt.Printf("Player Health: %d\n", playerHealth)
    fmt.Printf("Board: Round %d\n", round)
    rowOffset = 2
}


func initPrintBoard() {
    // Erase the screen
    fmt.Print("\033[2J")
    // Move the cursor to the top left
    fmt.Print("\033[H")

    printDamageBoard()

    printHeader()

    for Row := (boardSize - 1); Row >= 0; Row-- {
        for Col := 0; Col < boardSize; Col++ {
            printTile(Row, Col)
        }
        fmt.Println()
    }
}

func printBoard() {
    printHeader()

    for Row := (boardSize - 1); Row >= 0; Row-- {
        for Col := 0; Col < boardSize; Col++ {
            printActionTile(Row, Col)
        }
        fmt.Println()
    }
    fmt.Printf("\033[%d;%dH", boardRowSize, 0)

    fmt.Printf("Number of Player: %d\n", len(monsterList))
    for _, currentMonster := range monsterList {
        fmt.Printf("Row: %d, Col: %d, Health: %d\n", currentMonster.CurrentCord.Row, currentMonster.CurrentCord.Col, currentMonster.Health)
    }
}

func buildPrintBoard() {
    printHeader()
    
    for Row := (boardSize - 1); Row >= 0; Row-- {
        for Col := 0; Col < boardSize; Col++ {
            printBuildTile(Row, Col)
        }
        fmt.Println()
    }

    fmt.Printf("\033[%d;%dH", boardRowSize, 0)
    fmt.Printf("Row: %d, Col: %d \n", userCursor.Row, userCursor.Col)

    printDamageBoard()
}

func printDamageBoard() {
    for Row := (boardSize - 1); Row >= 0; Row-- {
        for Col := 0; Col < boardSize; Col++ {
            fmt.Printf("%2d ", towerHitList[Row][Col]) 
        }
        fmt.Println()
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

func placeTower(row int, col int) {
    towerPlacement[row][col] = 1

    currTile := cord.Cord { Row: row, Col: col } 

    nextTiles := []cord.Cord { 
        { Row: 0, Col: 0}, { Row: 0, Col: 1}, { Row: 1, Col: 0 }, { Row: 0, Col: -1 }, { Row: -1, Col: 0 },
    }

    
    for _, nextTile := range nextTiles {    
        currTile.Row += nextTile.Row
        currTile.Col += nextTile.Col

        if currTile.Row < boardSize && currTile.Row >= 0 && currTile.Col < boardSize && currTile.Col >= 0 {
            log.Println("thing2")
            towerHitList[currTile.Row][currTile.Col] += 10 
        }

        currTile.Row -= nextTile.Row
        currTile.Col -= nextTile.Col
    }
}

func handleInput() (string) {
    
    if err := keyboard.Open(); err != nil {
        log.Fatal(err)
    }

    defer keyboard.Close()

    for {
        rune, key, err := keyboard.GetKey()
        if err != nil {
            log.Fatal(err)
        }
    
        if key == keyboard.KeyEsc {
            return ""
        }

        keyChar := string(rune)

        if (key == keyboard.KeyEnter || key == keyboard.KeySpace) && !isInGameMapPathCords(userCursor) {
            placeTower(userCursor.Row, userCursor.Col)
        // Up
        } else if (key == keyboard.KeyArrowUp || keyChar == "k") && userCursor.Row < boardSize - 1 {
            userCursor.Row++
        // Left
        } else if (key == keyboard.KeyArrowLeft || keyChar == "h") && userCursor.Col > 0 {
            userCursor.Col--
        // Down
        } else if (key == keyboard.KeyArrowDown || keyChar == "j") && userCursor.Row > 0 {
            userCursor.Row--
        // Right
        } else if (key == keyboard.KeyArrowRight || keyChar == "l") && userCursor.Col < boardSize - 1 {
            userCursor.Col++
        }

        buildPrintBoard()
    }
}

func addPlayers() {
    if len(monsterList) < 5 {
        newPlayer := player.Player {
            CurrentCord: cord.Cord{
                Row: gameMapPathCords[0].Row,
                Col: gameMapPathCords[0].Col,
            },
            CurrentCordIndex: 0,
            Health: 100,
        }

        monsterList = append(monsterList, &newPlayer)
    }
}

func updateBoard() {
    // Move all the monsters
    for _, currentMonster := range monsterList {
        currentMonster.CurrentCordIndex++
        currentMonster.CurrentCord = gameMapPathCords[currentMonster.CurrentCordIndex]

        damage := towerHitList[currentMonster.CurrentCord.Row][currentMonster.CurrentCord.Col]

        fmt.Println("Monster at: ", currentMonster.CurrentCord.Row, currentMonster.CurrentCord.Col, " took damage: ", damage)

        currentMonster.Health -= damage
    }

    // Clears all the monster that are at the finish line
    monsterLen := len(monsterList)
    for index := 0; index < monsterLen; {
        if monsterList[index].Health <= 0 {
            monsterList[index] = monsterList[monsterLen-1]
            monsterLen-- 
        } else if monsterList[index].CurrentCord == *startCord {
            playerHealth -= 10
            monsterList[index] = monsterList[monsterLen-1]
            monsterLen-- 
        } else {
            index++
        }
    }

    monsterList = monsterList[:monsterLen] 
    round++
}

func gameMainLoop() {
    initPrintBoard()

    actionPhase := false 

    for (playerHealth > 0) {
        if !actionPhase {
            handleInput()  
            actionPhase = true
        } else {
            updateBoard()
            addPlayers()
            time.Sleep(1000 * time.Millisecond)
            printBoard()
        }
    }
}

func main() {

    file, err := os.OpenFile("app.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
    if err != nil {
        log.Fatal(err)
    }
    log.SetOutput(file)

    initBoard()
    gameMainLoop()
}
