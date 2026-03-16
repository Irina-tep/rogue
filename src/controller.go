package main

// Здесь обрабатывается ввод
import (
	"fmt"
	"log"
	"math"
	"math/rand/v2"
	"strconv"
	"time"

	"github.com/rthornton128/goncurses"
)

type Controller struct {
	Game       *Game
	GameWindow *goncurses.Window
}

// конструктор .
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
	controller.Game.AddMessage(fmt.Sprintf("Key pressed: %c", ch)) // Добавляем сообщение о нажатой клавише
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
	case 'h', 'H': // обработка клавиш для использования предметов (h для оружия, j для еды, k для эликсиров, e для свитков)
		controller.UseWeapon()
		return
	case 'j', 'J':
		controller.UseFood()
	case 'k', 'K':
		controller.UseElixir()
	case 'e', 'E':
		controller.UseScroll()
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

func (c *Controller) MovePlayer(dx, dy int) {
	if c.Game.Player.IsSleeping {
		c.Game.Player.IsSleeping = false
		c.Game.AddMessage("Player is sleeping, cannot move this turn!")
		return
	}

	newX := c.Game.Player.PosX + dx
	newY := c.Game.Player.PosY + dy

	if newX < 0 || newX >= ScreenWidth || newY < 0 || newY >= ScreenHeight {
		c.Game.AddMessage("Cannot move outside the level!")
		return
	}

	tile := c.Game.CurrentLevel.Tiles[newX][newY]
	if !tile.Blocked {
		c.Game.Player.PosX = newX
		c.Game.Player.PosY = newY
		c.Game.Player.CountTile++

		if tile.Symbol == '%' {
			if c.Game.CurrentLevelIndex < CountLevels-1 {
				c.Game.UpgradeLevel()
			} else {
				c.Game.StateGame = YouWin
				c.SaveStatistics(true)
			}
		}

		c.PickUpObject()
		c.UpdateTiles()
	} else {
		c.Game.AddMessage("Cannot move there!")
	}
}

// func (controller *Controller) MovePlayer(dx int, dy int) {
// 	newX, newY := controller.Game.Player.PosX+dx, controller.Game.Player.PosY+dy

// 	// Проверяем, что новая позиция в пределах уровня
// 	// if newX >= 0 && newX < ScreenWidth && newY >= 0 && newY < ScreenHeight {
// 	tile := controller.Game.CurrentLevel.Tiles[newX][newY]
// 	if !tile.Blocked {
// 		controller.Game.Player.PosX = newX
// 		controller.Game.Player.PosY = newY
// 		controller.Game.Player.CountTile++

// 		// Проверяем, не наступили ли на портал
// 		if tile.Symbol == '%' {
// 			if controller.Game.CurrentLevelIndex < CountLevels-1 {
// 				controller.Game.UpgradeLevel()
// 			} else {
// 				controller.Game.StateGame = YouWin
// 				controller.SaveStatistics(true)
// 			}
// 		}

// 		// Подбираем предмет, если он есть на клетке
// 		controller.PickUpObject()
// 		controller.UpdateTiles()
// 	}
// }

//}

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
	// строим клетки с коридорами
	for _, path := range controller.Game.CurrentLevel.Tunnels {
		for _, coordPath := range path.Path {
			tunnels, err := NewTile(coordPath[0], coordPath[1], TileTunnel)
			if err != nil {
				log.Fatal(err)
			}
			controller.Game.CurrentLevel.Tiles[coordPath[0]][coordPath[1]] = tunnels
		}
	}
	// обновляем клетки с врагами
	for _, enemy := range controller.Game.CurrentLevel.Enemies {
		tile, err := NewTile(enemy.PosXEnemy, enemy.PosYEnemy, enemy.TypeEnemy)
		if err != nil {
			log.Fatal(err)
		}
		controller.Game.CurrentLevel.Tiles[enemy.PosXEnemy][enemy.PosYEnemy] = tile
	}

	// обновляем клетку объектов
	for _, object := range controller.Game.CurrentLevel.Objects {
		symbol := GetObjectSymbol(object.TypeObject)
		tile, err := NewTile(object.PosX, object.PosY, string(symbol))
		if err != nil {
			log.Fatal(err)
		}
		controller.Game.CurrentLevel.Tiles[object.PosX][object.PosY] = tile
	}

	// обновляем клетку игрока
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

func (c *Controller) EnemyTurn() {
	for i := 0; i < len(c.Game.CurrentLevel.Enemies); i++ {
		enemy := &c.Game.CurrentLevel.Enemies[i]

		// Проверяем, видит ли враг игрока
		if c.isPlayerInRange(enemy) {
			enemy.Mode = Chasing
		} else {
			enemy.Mode = Roaming
		}

		// Перемещаем врага
		newX, newY := enemy.PosXEnemy, enemy.PosYEnemy
		if enemy.Mode == Roaming {
			newX, newY = enemy.EnemyMove(c.Game.CurrentLevel)
		} else if enemy.Mode == Chasing {
			newX, newY = enemy.ChaseTarget(c.Game.CurrentLevel, c.Game.Player.PosX, c.Game.Player.PosY)
		}

		// Проверяем, можно ли переместиться на новую позицию
		if !c.Game.CurrentLevel.Tiles[newX][newY].Blocked {
			enemy.PosXEnemy = newX
			enemy.PosYEnemy = newY
		}

		// Проверяем, находится ли игрок на той же клетке, что и враг
		if enemy.PosXEnemy == c.Game.Player.PosX && enemy.PosYEnemy == c.Game.Player.PosY {
			enemy.Attack(c.Game.Player)
			if c.Game.Player.HP <= 0 {
				c.Game.StateGame = YouLose
			}
		}
	}
}

func (c *Controller) isPlayerInRange(enemy *Enemy) bool {
	dx := c.Game.Player.PosX - enemy.PosXEnemy
	dy := c.Game.Player.PosY - enemy.PosYEnemy
	distance := math.Sqrt(float64(dx*dx + dy*dy))
	return distance <= float64(enemy.HostilityEnemy)
}

// func (controller *Controller) EnemyTurn() {
// 	for i := 0; i < len(controller.Game.CurrentLevel.Enemies); i++ {
// 		x, y := controller.Game.CurrentLevel.Enemies[i].PosXEnemy, controller.Game.CurrentLevel.Enemies[i].PosYEnemy
// 		if controller.Game.CurrentLevel.Enemies[i].Mode == Roaming {
// 			x, y = controller.Game.CurrentLevel.Enemies[i].EnemyMove()
// 		} else if controller.Game.CurrentLevel.Enemies[i].Mode == Chasing {
// 			x, y = controller.Game.CurrentLevel.Enemies[i].ChaseTarget(controller.Game.Player.PosX, controller.Game.Player.PosY)
// 		}

// 		tile := controller.Game.CurrentLevel.Tiles[x][y]
// 		if !tile.BlockedForEnemy {
// 			controller.Game.CurrentLevel.Enemies[i].PosXEnemy = x
// 			controller.Game.CurrentLevel.Enemies[i].PosYEnemy = y
// 			controller.UpdateTiles()
// 		}
// 	}
// }

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

// метод для подбора предметов с поля
func (c *Controller) PickUpObject() {
	for i, obj := range c.Game.CurrentLevel.Objects {
		if obj.PosX == c.Game.Player.PosX && obj.PosY == c.Game.Player.PosY {
			if c.Game.Player.Backpack.AddObject(obj) {
				c.Game.AddMessage(fmt.Sprintf("Picked up %s", GetObjectSymbol(obj.TypeObject)))
				// Удаляем предмет с уровня
				c.Game.CurrentLevel.Objects = append(
					c.Game.CurrentLevel.Objects[:i],
					c.Game.CurrentLevel.Objects[i+1:]...,
				)
				c.UpdateTiles()
				return
			} else {
				c.Game.AddMessage("Backpack is full for this object type!")
				return
			}
		}
	}
}

// метод для добавления сокровищ игроку
func (controller *Controller) AddTreasure(amount int) {
	controller.Game.Player.Treasure += amount
	controller.Game.AddMessage(fmt.Sprintf("Gained %d treasure!", amount))
}

// Использование еды
func (c *Controller) UseFood() {
	objects := c.Game.Player.Backpack.GetObjects(FOOD)
	if len(objects) == 0 {
		c.Game.AddMessage("No food in backpack!")
		return
	}

	c.Game.AddMessage("Select food (1-" + strconv.Itoa(len(objects)) + "):")
	for i, object := range objects {
		c.Game.AddMessage(fmt.Sprintf("%d. %s", i+1, GetObjectSymbol(object.TypeObject)))
	}

	// Здесь можно добавить логику выбора предмета через ввод пользователя.
	// Для простоты возьмем первый предмет.
	object, _ := c.Game.Player.Backpack.UseObject(FOOD, 0)
	c.ApplyObjectEffect(object)
}

// Использование эликсира
func (c *Controller) UseElixir() {
	objects := c.Game.Player.Backpack.GetObjects(ELEXIR)
	if len(objects) == 0 {
		c.Game.AddMessage("No elixirs in backpack!")
		return
	}

	c.Game.AddMessage("Select elixir (1-" + strconv.Itoa(len(objects)) + "):")
	for i, object := range objects {
		c.Game.AddMessage(fmt.Sprintf("%d. %s", i+1, GetObjectSymbol(object.TypeObject)))
	}

	object, _ := c.Game.Player.Backpack.UseObject(ELEXIR, 0)
	c.ApplyObjectEffect(object)
}

// Использование свитка
func (c *Controller) UseScroll() {
	objects := c.Game.Player.Backpack.GetObjects(SCROL)
	if len(objects) == 0 {
		c.Game.AddMessage("No scrolls in backpack!")
		return
	}

	c.Game.AddMessage("Select scroll (1-" + strconv.Itoa(len(objects)) + "):")
	for i, object := range objects {
		c.Game.AddMessage(fmt.Sprintf("%d. %s", i+1, GetObjectSymbol(object.TypeObject)))
	}

	object, _ := c.Game.Player.Backpack.UseObject(SCROL, 0)
	c.ApplyObjectEffect(object)
}

// Использование оружия
func (c *Controller) UseWeapon() {
	objects := c.Game.Player.Backpack.GetObjects(WEAPON)
	if len(objects) == 0 {
		c.Game.AddMessage("No weapons in backpack!")
		return
	}

	// Выбираем оружие (например, первое)
	selectedObject, _ := c.Game.Player.Backpack.UseObject(WEAPON, 0)

	// Если у игрока уже было оружие, бросаем его на пол
	if c.Game.Player.CurrenWeapon != "1d1" {
		oldWeapon := &Object{
			TypeObject: WEAPON,
			Damage:     c.Game.Player.CurrenWeapon,
		}
		// Бросаем оружие на соседнюю клетку
		dropX, dropY := c.Game.Player.PosX+1, c.Game.Player.PosY
		if dropX >= ScreenWidth {
			dropX = c.Game.Player.PosX - 1
		}
		oldWeapon.PosX, oldWeapon.PosY = dropX, dropY
		c.Game.CurrentLevel.Objects = append(c.Game.CurrentLevel.Objects, oldWeapon)
		c.Game.AddMessage(fmt.Sprintf("Dropped old weapon %s", oldWeapon.Damage))
	}

	// Одеваем новое оружие
	if selectedObject != nil {
		c.Game.Player.CurrenWeapon = selectedObject.Damage
		c.Game.AddMessage(fmt.Sprintf("Equipped %s (Damage: %s)", GetObjectSymbol(selectedObject.TypeObject), selectedObject.Damage))
	} else {
		c.Game.Player.CurrenWeapon = "1d1"
		c.Game.AddMessage("Unequipped weapon")
	}
}

// Применение эффектов предмета
func (c *Controller) ApplyObjectEffect(object *Object) {
	if object == nil {
		return
	}

	switch object.TypeObject {
	case ELEXIR:
		// Временное увеличение характеристик
		if c.Game.Player.TemporaryEffects == nil {
			c.Game.Player.TemporaryEffects = make(map[string]int)
		}
		c.Game.Player.TemporaryEffects["dexterity"] = object.Dexterity
		c.Game.Player.TemporaryEffects["strength"] = object.Strength
		c.Game.Player.CountElixir++
		c.Game.AddMessage(fmt.Sprintf("Drank elixir. Dexterity: %d, Strength: %d (temporary)", c.Game.Player.Dexterity, c.Game.Player.Strength))
	}
}

// В конце каждого хода уменьшаем эффекты
func (c *Controller) DecreaseTemporaryEffects() {
	if c.Game.Player.TemporaryEffects != nil {
		for effect, value := range c.Game.Player.TemporaryEffects {
			if value > 0 {
				c.Game.Player.TemporaryEffects[effect] = value - 1
			} else {
				delete(c.Game.Player.TemporaryEffects, effect)
			}
		}
	}
}
