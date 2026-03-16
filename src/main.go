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
	Player            *Player
	Levels            []Level // все 21 уровень
	CurrentLevel      *Level  // текущий уровень
	CurrentLevelIndex int     // индекс текущего уровня
	Messages          []string
	Seed              uint64       // чтобы при запуске игры не генерировались одни и те же параметры, одна и та же последовательность
	RNG               *rand.Rand   // для генерации
	Running           bool         // тру - игра активна , если нет - программа заверщается
	StateGame         int          // сотояние игры
	SaveManager       *SaveManager // Добавляем менеджер сохранений
}

// сотояния игры
const (
	YouLose  int = 1 // ты поиграл
	YouWin   int = 2 // ты выиграл
	YouGame  int = 3 // ты играешь
	MainMenu int = 4 // главное меню
)

// Этот конструктор создаст Game для нас новый объект, который в данный момент пуст, но будет расширяться по мере дальнейшего выполнения кода.
func NewGame(state int) *Game {
	g := &Game{
		Seed:        (uint64)(time.Now().UnixNano()),
		Running:     true,
		StateGame:   state,
		SaveManager: NewSaveManager(), // Инициализируем
	}

	// Если это не меню, инициализируем уровни и игрока
	if state != MainMenu {
		g.RNG = rand.New(rand.NewPCG(g.Seed, 10))
		// Создаем уровни
		for i := 0; i < CountLevels; i++ {
			level := NewLevel()
			level.Number = i
			g.Levels = append(g.Levels, level) // Теперь игра загрузит нашу карту в качестве первого уровня. Использовть это, если мы будем делать слайс Level в структуре Game
		}
		// Инициализируем уровень
		g.CurrentLevelIndex = 0
		g.CurrentLevel = &g.Levels[0]
		// поместить игрока в рандомное место в первой комнате
		// startX, startY := g.CurrentLevel.Rooms[0].RandomPos()
		startX, startY := g.CurrentLevel.GetPos()
		player := NewPlayer(startX, startY, g.CurrentLevelIndex)
		g.Player = &player

		// Добавляем сообщения
		g.AddMessage("Start game!")
		g.AddMessage("Use WASD for action, q for exit")
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
	// Перемещаем игрока в стартовую позицию
	startX, startY := g.CurrentLevel.Rooms[0].RandomPos()
	g.Player.PosX = startX
	g.Player.PosY = startY

	g.AddMessage("I'm going up a level" + strconv.Itoa(g.CurrentLevelIndex+1))
}

// Добавление сообщения
func (g *Game) AddMessage(msg string) {
	g.Messages = append(g.Messages, msg)
	// Ограничиваем количество сообщений
	if len(g.Messages) > 100 {
		g.Messages = g.Messages[len(g.Messages)-100:]
	}
}

// если будем менять логику, то понадобится, пока не нужно
// func (g *Game) Update() error {
// 	HandleInput(g)
// 	return nil
// }

func main() {
	// Создаем игру в состоянии главного меню
	game := NewGame(MainMenu)

	render := Renderer{}
	if err := render.Init(); err != nil {
		log.Fatalf("Failed to initialize renderer: %v", err)
	}
	controller := NewController(game, render.GameWindow)
	defer render.Cleanup()

	for game.Running {
		render.Render(game)
		controller.HandleInput()

		// Если игра перешла в состояние игры, обрабатываем врагов
		if game.StateGame == YouGame {
			controller.EnemyFOV()
			controller.EnemyTurn()
		}
	}
}

// Для сохранения
func (g *Game) LoadFromSave(saveData *SaveData) error {
	if saveData == nil {
		return fmt.Errorf("saveData is nil")
	}

	// Загружаем данные игрока
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

	// Восстанавливаем рюкзак
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

	// Проверка и инициализация временных эффектов
	if g.Player.TemporaryEffects == nil {
		g.Player.TemporaryEffects = make(map[string]int)
	}

	// Восстанавливаем временные эффекты
	g.Player.TemporaryEffects = saveData.Player.TemporaryEffects

	// Создаем новый уровень
	g.CurrentLevel = &Level{
		Tiles:   make([][]Tile, ScreenWidth),
		Rooms:   make([]Room, len(saveData.CurrentLevel.Rooms)),
		Tunnels: make([]Tunnel, len(saveData.CurrentLevel.Tunnels)),
		Enemies: make([]Enemy, len(saveData.CurrentLevel.Enemies)),
		Number:  saveData.CurrentLevel.Number,
	}

	// Инициализируем матрицу тайлов
	for x := 0; x < ScreenWidth; x++ {
		g.CurrentLevel.Tiles[x] = make([]Tile, ScreenHeight)
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
		// Восстанавливаем указатель на комнату
		var currentRoom *Room
		if enemyData.CurrentRoom >= 0 && enemyData.CurrentRoom < len(g.CurrentLevel.Rooms) {
			currentRoom = &g.CurrentLevel.Rooms[enemyData.CurrentRoom]
		} else {
			currentRoom = nil
		}

		// Создаем врага со всеми полями
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
