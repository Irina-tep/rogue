package main

import (
	"fmt"
	"log"
	"math/rand/v2"
	"strconv"
	"time"
)

const (
	CountLevels int = 21
)

// главная структура игры
type Game struct {
	Player                 *Player
	Levels                 []Level
	CurrentLevel           *Level
	CurrentLevelIndex      int
	Messages               []string
	Seed                   uint64
	RNG                    *rand.Rand
	Running                bool
	StateGame              int
	SaveManager            *SaveManager
	Renderer               *Renderer
	WaitingForWeaponChoice bool
	WaitingForFoodChoice   bool
	WaitingForElixirChoice bool
	WaitingForScrollChoice bool
}

// сотояния игры
const (
	YouLose  int = 1 // ты поиграл
	YouWin   int = 2 // ты выиграл
	YouGame  int = 3 // ты играешь
	MainMenu int = 4 // главное меню
)

func NewGame(state int) *Game {
	g := &Game{
		Seed:        (uint64)(time.Now().UnixNano()),
		Running:     true,
		StateGame:   state,
		SaveManager: NewSaveManager(),
	}
	return g
}

// переходит на следующий уровень
func (g *Game) UpgradeLevel() {
	if g.CurrentLevelIndex >= CountLevels-1 {
		g.AddMessage("You have reached the final level")
		return
	}
	g.CurrentLevelIndex++
	g.CurrentLevel = &g.Levels[g.CurrentLevelIndex]
	g.Player.CurrentLevelIndex = g.CurrentLevelIndex

	startX, startY := g.CurrentLevel.Rooms[0].RandomPos()
	g.Player.PosX = startX
	g.Player.PosY = startY

	g.AddMessage("I'm going up a level" + strconv.Itoa(g.CurrentLevelIndex+1))
}

// Добавление сообщения
func (g *Game) AddMessage(msg string) {
	g.Messages = append(g.Messages, msg)
	if len(g.Messages) > 100 {
		g.Messages = g.Messages[len(g.Messages)-100:]
	}
}

func main() {
	// Создаем игру в состоянии главного меню
	game := NewGame(MainMenu)
	render := Renderer{}
	if err := render.Init(); err != nil {
		log.Fatalf("Failed to initialize renderer: %v", err)
	}
	game.Renderer = &render
	controller := NewController(game, render.GameWindow)
	defer render.Cleanup()
	flag := true
	for flag {
		render.Render(game)
		flag = controller.HandleMenuInput()
	}

	for game.Running {
		render.Render(game)
		controller.HandleInput()
		if game.StateGame == YouGame {
			controller.EnemyFOV()
			controller.EnemyTurn()
		}
	}
}

// Для загрузки сохраненных данных
func (g *Game) LoadFromSave(saveData *SaveData) error {
	if saveData == nil {
		return fmt.Errorf("saveData is nil")
	}

	// Загружаем данные игрока
	g.Player.Name = saveData.Player.Name
	g.Player.PosX = saveData.Player.PosX
	g.Player.PosY = saveData.Player.PosY
	g.Player.HP = saveData.Player.HP
	g.Player.MaxHP = saveData.Player.MaxHP
	g.Player.Dexterity = saveData.Player.Dexterity
	g.Player.Strength = saveData.Player.Strength
	g.Player.CurrenWeapon = saveData.Player.CurrenWeapon
	g.Player.Treasure = saveData.Player.Treasure
	g.Player.CurrentLevelIndex = saveData.Player.CurrentLevelIndex
	g.Player.CountEnemy = saveData.Player.CountEnemy
	g.Player.CountFood = saveData.Player.CountFood
	g.Player.CountElixir = saveData.Player.CountElixir
	g.Player.CountScrollsRead = saveData.Player.CountScrollsRead
	g.Player.CountHits = saveData.Player.CountHits
	g.Player.CountTile = saveData.Player.CountTile
	g.Player.IsSleeping = saveData.Player.IsSleeping
	g.Player.Name = saveData.Player.Name

	g.Player.Backpack = NewBackpack()
	for objectType, objects := range saveData.Player.Backpack {
		for _, objectData := range objects {
			object := &Object{
				TypeObject:    objectData.TypeObject,
				SubtypeObject: objectData.SubtypeObject,
				Health:        objectData.Health,
				MaxHealth:     objectData.MaxHealth,
				Dexterity:     objectData.Dexterity,
				Strength:      objectData.Strength,
				ValueObject:   objectData.ValueObject,
				Damage:        objectData.Damage,
			}
			g.Player.Backpack.Objects[objectType] = append(g.Player.Backpack.Objects[objectType], object)
		}
	}

	if g.Player.TemporaryEffects == nil {
		g.Player.TemporaryEffects = make(map[string]int)
	}

	g.Player.TemporaryEffects = saveData.Player.TemporaryEffects

	g.CurrentLevel = &Level{
		Tiles:    make([][]Tile, ScreenWidth),
		Rooms:    make([]Room, len(saveData.CurrentLevel.Rooms)),
		Tunnels:  make([]Tunnel, len(saveData.CurrentLevel.Tunnels)),
		Enemies:  make([]Enemy, len(saveData.CurrentLevel.Enemies)),
		Number:   saveData.CurrentLevel.Number,
		Explored: make([][]bool, ScreenWidth),
		Visible:  make([][]bool, ScreenWidth),
	}

	for x := 0; x < ScreenWidth; x++ {
		g.CurrentLevel.Tiles[x] = make([]Tile, ScreenHeight)
	}

	for x := 0; x < ScreenWidth; x++ {
		g.CurrentLevel.Explored[x] = make([]bool, ScreenHeight)
		g.CurrentLevel.Visible[x] = make([]bool, ScreenHeight)
		for y := 0; y < ScreenHeight; y++ {
			if x < len(saveData.CurrentLevel.Explored) && y < len(saveData.CurrentLevel.Explored[x]) {
				g.CurrentLevel.Explored[x][y] = saveData.CurrentLevel.Explored[x][y]
			} else {
				g.CurrentLevel.Explored[x][y] = false
			}
			if x < len(saveData.CurrentLevel.Visible) && y < len(saveData.CurrentLevel.Visible[x]) {
				g.CurrentLevel.Visible[x][y] = saveData.CurrentLevel.Visible[x][y]
			} else {
				g.CurrentLevel.Visible[x][y] = false
			}
		}
	}

	// ВОССТАНАВЛИВАЕМ ТАЙЛЫ - конвертируем из TileData в Tile
	// saveData.CurrentLevel.Tiles является [][]TileData
	for x := 0; x < ScreenWidth && x < len(saveData.CurrentLevel.Tiles); x++ {
		for y := 0; y < ScreenHeight && y < len(saveData.CurrentLevel.Tiles[x]); y++ {
			tileData := saveData.CurrentLevel.Tiles[x][y]
			g.CurrentLevel.Tiles[x][y] = Tile{
				PosX:            tileData.X,
				PosY:            tileData.Y,
				Blocked:         tileData.Blocked,
				Symbol:          tileData.Symbol,
				BlockedForEnemy: tileData.BlockedForEnemy,
			}
		}
	}

	// Загружаем комнаты
	for i, roomData := range saveData.CurrentLevel.Rooms {
		g.CurrentLevel.Rooms[i] = Room{
			X1: roomData.X1,
			Y1: roomData.Y1,
			X2: roomData.X2,
			Y2: roomData.Y2,
		}
	}

	// Загружаем туннели
	for i, tunnelData := range saveData.CurrentLevel.Tunnels {
		g.CurrentLevel.Tunnels[i] = Tunnel{
			Path: tunnelData.Path,
		}
	}

	// ВОССТАНАВЛИВАЕМ ВРАГОВ со всеми полями
	for i, enemyData := range saveData.CurrentLevel.Enemies {
		var currentRoom *Room
		if enemyData.CurrentRoom >= 0 && enemyData.CurrentRoom < len(g.CurrentLevel.Rooms) {
			currentRoom = &g.CurrentLevel.Rooms[enemyData.CurrentRoom]
		} else {
			currentRoom = nil
		}

		g.CurrentLevel.Enemies[i] = Enemy{
			PosXEnemy:      enemyData.PosXEnemy,
			PosYEnemy:      enemyData.PosYEnemy,
			TypeEnemy:      enemyData.TypeEnemy,
			HealthEnemy:    enemyData.HealthEnemy,
			DexterityEnemy: enemyData.DexterityEnemy,
			StrengthEnemy:  enemyData.StrengthEnemy,
			HostilityEnemy: enemyData.HostilityEnemy,
			CurrentRoom:    currentRoom,
			Mode:           enemyData.Mode,
			Treasure:       enemyData.Treasure,
		}
	}

	// Устанавливаем текущий уровень в Levels слайсе
	if saveData.CurrentLevel.Number < CountLevels {
		// Если уровень находится в пределах созданных уровней, заменяем его
		if saveData.CurrentLevel.Number < len(g.Levels) {
			g.Levels[saveData.CurrentLevel.Number] = *g.CurrentLevel
		}
		g.CurrentLevelIndex = saveData.CurrentLevel.Number
	}

	return nil
}
