package main

//Здесь обрабатывается ввод
import (
	"fmt"
	"log"
	"math"
	"math/rand/v2"
	"time"

	"github.com/rthornton128/goncurses"
)

type Controller struct {
	Game       *Game
	GameWindow *goncurses.Window
}

// конструктор
func NewController(game *Game, gameWindow *goncurses.Window) Controller {
	return Controller{
		Game:       game,
		GameWindow: gameWindow,
	}
}

// Обработка ввода
func (controller *Controller) HandleInput() {
	if controller.Game.StateGame == MainMenu {
		controller.HandleMenuInput()
		return
	}

	ch := controller.GameWindow.GetChar()
	dx := 0
	dy := 0
	switch ch {
	case 'q', 'Q':
		// Сохраняем игру перед выходом
		controller.SaveOnExit()
		controller.Game.Running = false
	case 'w', 'W':
		dy = -1
	case 's', 'S':
		dy = 1
	case 'a', 'A':
		dx = -1
	case 'd', 'D':
		dx = 1
	}
	controller.MovePlayer(dx, dy)
	if dx != 0 || dy != 0 {
		controller.Game.Player.CountTile = controller.Game.Player.CountTile + 1
	}
}

// HandleMenuInput - обработка ввода в главном меню
func (controller *Controller) HandleMenuInput() {
	ch := controller.GameWindow.GetChar()
	switch ch {
	case '1':
		// Продолжить последнее сохранение
		controller.LoadLastSave()
	case '2':
		// Начать новую игру
		controller.StartNewGame()
	case 'q', 'Q':
		// Выйти из игры
		controller.Game.Running = false
	}
}

// LoadLastSave - загрузка последнего сохранения
func (c *Controller) LoadLastSave() {
	saveData, err := c.Game.SaveManager.LoadGame("last_save.json")
	if err != nil {
		c.Game.AddMessage(fmt.Sprintf("No save found: %v", err))
		// Если сохранения нет, начинаем новую игру
		c.StartNewGame()
		return
	}

	// Инициализируем игру (уровни, игрока)
	c.Game.RNG = rand.New(rand.NewPCG(c.Game.Seed, 10))
	// Создаем уровни
	for i := 0; i < CountLevels; i++ {
		level := NewLevel()
		level.Number = i
		c.Game.Levels = append(c.Game.Levels, level)
	}

	// Устанавливаем текущий уровень (временный)
	c.Game.CurrentLevelIndex = 0
	c.Game.CurrentLevel = &c.Game.Levels[0]

	// Создаем временного игрока (поля будут перезаписаны в LoadFromSave)
	startX, startY := c.Game.CurrentLevel.GetPos()
	player := NewPlayer(startX, startY, c.Game.CurrentLevelIndex)
	c.Game.Player = &player

	// Загружаем данные сохранения
	err = c.Game.LoadFromSave(saveData)
	if err != nil {
		c.Game.AddMessage(fmt.Sprintf("Failed to load save: %v", err))
		// Если не удалось загрузить, начинаем новую игру
		c.StartNewGame()
	} else {
		c.Game.StateGame = YouGame
		c.Game.AddMessage("Game loaded from last save")
	}
}

// StartNewGame - начать новую игру
func (c *Controller) StartNewGame() {
	// Инициализируем игру
	c.Game.RNG = rand.New(rand.NewPCG(c.Game.Seed, 10))
	// Создаем уровни
	for i := 0; i < CountLevels; i++ {
		level := NewLevel()
		level.Number = i
		c.Game.Levels = append(c.Game.Levels, level)
	}
	// Инициализируем уровень
	c.Game.CurrentLevelIndex = 0
	c.Game.CurrentLevel = &c.Game.Levels[0]
	// Создаем игрока
	startX, startY := c.Game.CurrentLevel.GetPos()
	player := NewPlayer(startX, startY, c.Game.CurrentLevelIndex)
	c.Game.Player = &player

	// Добавляем сообщения
	c.Game.AddMessage("Start new game!")
	c.Game.AddMessage("Use WASD for action, q for exit")

	c.Game.StateGame = YouGame
}

func (controller *Controller) MovePlayer(dx int, dy int) {
	tile := controller.Game.CurrentLevel.Tiles[controller.Game.Player.PosX+dx][controller.Game.Player.PosY+dy] //никогда не выйдет за границы комнаты
	if !tile.Blocked {
		controller.Game.Player.PosX += dx
		controller.Game.Player.PosY += dy
		if tile.Symbol == '%' {
			if controller.Game.CurrentLevelIndex < CountLevels-1 {
				controller.Game.UpgradeLevel()
			} else {
				controller.Game.StateGame = YouWin
				// Сохраняем статистику при победе
				controller.SaveStatistics(true)
			}
		}
	}
	controller.UpdateTiles()
}

// меняем состояние игры, для этого нужно обновить тайлы. Когда добавляем новые объекты в поле, обновляем тайлы здесь, в рендере не должно быть никакой логики, там только отрисовываются обновленные тайлы
func (controller *Controller) UpdateTiles() {

	// мы проходим по всем комнатам уровня и формируем их интерьер: стены, пол
	for _, room := range controller.Game.CurrentLevel.Rooms {
		x1, x2, y1, y2 := room.Interior()
		for x := x1; x <= x2; x++ {
			for y := y1; y <= y2; y++ {
				floor, err := NewTile(x, y, TileFloor)
				if err != nil {
					log.Fatal(err)
				}
				if controller.Game.CurrentLevel.Tiles[x][y].Symbol != '%' {
					controller.Game.CurrentLevel.Tiles[x][y] = floor
				}

			}
		}
	}
	//строим клетки с коридорами
	for _, path := range controller.Game.CurrentLevel.Tunnels {
		for _, coordPath := range path.Path {
			tunnels, err := NewTile(coordPath[0], coordPath[1], TileTunnel)
			if err != nil {
				log.Fatal(err)
			}
			controller.Game.CurrentLevel.Tiles[coordPath[0]][coordPath[1]] = tunnels
		}
	}
	//обновляем клетки с врагами
	for _, enemy := range controller.Game.CurrentLevel.Enemies {
		tile, err := NewTile(enemy.PosXEnemy, enemy.PosYEnemy, enemy.TypeEnemy)
		if err != nil {
			log.Fatal(err)
		}
		controller.Game.CurrentLevel.Tiles[enemy.PosXEnemy][enemy.PosYEnemy] = tile
	}

	//обновляем клетку объектов
	for _, object := range controller.Game.CurrentLevel.Objects {
		symbol := GetObjectSymbol(object.TypeObject)
		tile, err := NewTile(object.PosX, object.PosY, string(symbol))
		if err != nil {
			log.Fatal(err)
		}
		controller.Game.CurrentLevel.Tiles[object.PosX][object.PosY] = tile
	}

	//обновляем клетку игрока
	player, err := NewTile(controller.Game.Player.PosX, controller.Game.Player.PosY, TilePlayer)
	if err != nil {
		log.Fatal(err)
	}
	controller.Game.CurrentLevel.Tiles[controller.Game.Player.PosX][controller.Game.Player.PosY] = player
}

// область видимости врагов, меняет режим врага на преследование
func (controller *Controller) EnemyFOV() {
	for i := 0; i < len(controller.Game.CurrentLevel.Enemies); i++ {
		x1, x2, y1, y2 := controller.Game.CurrentLevel.Enemies[i].CurrentRoom.Interior()
		if controller.Game.Player.PosX >= x1 && controller.Game.Player.PosX <= x2 && controller.Game.Player.PosY >= y1 && controller.Game.Player.PosY <= y2 {
			dX := controller.Game.Player.PosX - controller.Game.CurrentLevel.Enemies[i].PosXEnemy
			dY := controller.Game.Player.PosY - controller.Game.CurrentLevel.Enemies[i].PosYEnemy
			len := math.Sqrt(math.Pow(float64(dX), 2) + math.Pow(float64(dY), 2))
			if int(len) <= controller.Game.CurrentLevel.Enemies[i].HostilityEnemy {
				controller.Game.CurrentLevel.Enemies[i].Mode = Chasing
			} else {
				controller.Game.CurrentLevel.Enemies[i].Mode = Roaming
			}

		} else {
			controller.Game.CurrentLevel.Enemies[i].Mode = Roaming
		}

	}
}

func (controller *Controller) EnemyTurn() {
	for i := 0; i < len(controller.Game.CurrentLevel.Enemies); i++ {
		x, y := controller.Game.CurrentLevel.Enemies[i].PosXEnemy, controller.Game.CurrentLevel.Enemies[i].PosYEnemy
		if controller.Game.CurrentLevel.Enemies[i].Mode == Roaming {
			x, y = controller.Game.CurrentLevel.Enemies[i].EnemyMove()

		} else if controller.Game.CurrentLevel.Enemies[i].Mode == Chasing {
			x, y = controller.Game.CurrentLevel.Enemies[i].ChaseTarget(controller.Game.Player.PosX, controller.Game.Player.PosY)
		}

		tile := controller.Game.CurrentLevel.Tiles[x][y]
		if !tile.BlockedForEnemy {
			controller.Game.CurrentLevel.Enemies[i].PosXEnemy = x
			controller.Game.CurrentLevel.Enemies[i].PosYEnemy = y
			controller.UpdateTiles()
		}
	}
}

// HandleSave - сохранение игры
func (c *Controller) HandleSave() {
	// Автоматическое имя сохранения
	saveName := "last_save"
	err := c.Game.SaveManager.SaveGame(c.Game, saveName)
	if err != nil {
		c.Game.AddMessage(fmt.Sprintf("Save failed: %v", err))
	} else {
		c.Game.AddMessage(fmt.Sprintf("Game saved as %s", saveName))
	}
}

// HandleLoad - загрузка игры
func (c *Controller) HandleLoad() {
	// Получаем список сохранений
	saves, err := c.Game.SaveManager.ListSaves()
	if err != nil {
		c.Game.AddMessage(fmt.Sprintf("Failed to list saves: %v", err))
		return
	}

	if len(saves) == 0 {
		c.Game.AddMessage("No saves found")
		return
	}

	// Загружаем последнее сохранение
	saveData, err := c.Game.SaveManager.GetLatestSave()
	if err != nil {
		c.Game.AddMessage(fmt.Sprintf("Failed to load save: %v", err))
		return
	}

	// Загружаем данные в игру
	err = c.Game.LoadFromSave(saveData)
	if err != nil {
		c.Game.AddMessage(fmt.Sprintf("Failed to load game: %v", err))
	} else {
		c.Game.AddMessage("Game loaded successfully")
	}
}

// SaveOnExit - сохранение при выходе
func (c *Controller) SaveOnExit() {
	c.Game.AddMessage("Saving game before exit...")
	c.HandleSave()
	// Также сохраняем статистику текущего прохождения
	c.SaveStatistics(false)
}

// SaveStatistics - сохранение статистики при завершении игры
func (c *Controller) SaveStatistics(isCompleted bool) {
	// Вычисляем общее время игры (примерно)
	totalPlayTime := int(time.Since(time.Unix(0, int64(c.Game.Seed))).Seconds())
	if totalPlayTime < 0 {
		totalPlayTime = 0
	}

	// Сохраняем статистику
	c.Game.SaveManager.AddStatistic(
		"Player", // Можно добавить ввод имени игрока позже
		c.Game.Player.CurrentLevelIndex,
		c.Game.Player.CountEnemy,
		c.Game.Player.Treasure,
		c.Game.Player.CountTile,
		totalPlayTime,
		isCompleted,
	)

	c.Game.AddMessage(fmt.Sprintf("Statistics saved. Reached level: %d", c.Game.Player.CurrentLevelIndex))
}
