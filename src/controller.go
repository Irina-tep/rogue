package main

import (
	"fmt"
	"log"
	"math"
	"math/rand/v2"
	"strconv"
	"strings"
	"time"

	"github.com/rthornton128/goncurses"
)

type Controller struct {
	Game       *Game
	GameWindow *goncurses.Window
}

func NewController(game *Game, gameWindow *goncurses.Window) Controller {
	return Controller{
		Game:       game,
		GameWindow: gameWindow,
	}
}

// Обработка ввода
func (controller *Controller) HandleInput() {
	ch := controller.GameWindow.GetChar()
	dx := 0
	dy := 0
	if controller.Game.WaitingForWeaponChoice {
		controller.HandleWeaponChoice(rune(ch))
		return
	}

	if controller.Game.WaitingForFoodChoice {
		controller.HandleFoodChoice(rune(ch))
		return
	}

	if controller.Game.WaitingForElixirChoice {
		controller.HandleElixirChoice(rune(ch))
		return
	}

	if controller.Game.WaitingForScrollChoice {
		controller.HandleScrollChoice(rune(ch))
		return
	}
	switch ch {
	case 113, 81: // 'q' и 'Q'
		controller.SaveOnExit()
		controller.Game.Running = false
	case 119, 87: // 'w' и 'W'
		dy = -1
	case 115, 83: // 's' и 'S'
		dy = 1
	case 97, 65: // 'a' и 'A'
		dx = -1
	case 100, 68: // 'd' и 'D'
		dx = 1
	case 'h', 'H':
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
func (c *Controller) HandleMenuInput() bool {
	ch := c.Game.Renderer.GameWindow.GetChar()

	switch ch {
	case '1':
		// Продолжить последнее сохранение
		c.LoadLastSave()
		return false
	case '2':
		// Начать новую игру
		c.StartNewGame()
		return false
	case '3':
		// Таблица лидеров
		c.ShowLeaderboard()
		return true
	case 'q', 'Q':
		// Выйти из игры
		c.Game.Running = false
		return false
	default:
		return true
	}
}

// LoadLastSave - загрузка последнего сохранения
func (c *Controller) LoadLastSave() {
	saveData, err := c.Game.SaveManager.LoadGame("last_save.json")
	if err != nil {
		c.Game.AddMessage(fmt.Sprintf("No save found: %v", err))
		c.StartNewGame()
		return
	}

	c.Game.RNG = rand.New(rand.NewPCG(c.Game.Seed, 10))
	for i := 0; i < CountLevels; i++ {
		level := NewLevel()
		level.Number = i
		c.Game.Levels = append(c.Game.Levels, level)
	}

	c.Game.CurrentLevelIndex = 0
	c.Game.CurrentLevel = &c.Game.Levels[0]

	startX, startY := c.Game.CurrentLevel.GetPos()
	player := NewPlayer(startX, startY, c.Game.CurrentLevelIndex)
	c.Game.Player = &player

	err = c.Game.LoadFromSave(saveData)
	if err != nil {
		c.Game.AddMessage(fmt.Sprintf("Failed to load save: %v", err))
		c.StartNewGame()
	} else {
		c.Game.StateGame = YouGame
		c.Game.AddMessage("Game loaded from last save")
	}
}

// StartNewGame - начать новую игру
func (c *Controller) StartNewGame() {
	playerName := c.Game.Renderer.EnterPlayerName()
	c.Game.RNG = rand.New(rand.NewPCG(c.Game.Seed, 10))
	for i := 0; i < CountLevels; i++ {
		level := NewLevel()
		level.Number = i
		c.Game.Levels = append(c.Game.Levels, level)
	}
	c.Game.CurrentLevelIndex = 0
	c.Game.CurrentLevel = &c.Game.Levels[0]
	startX, startY := c.Game.CurrentLevel.GetPos()
	player := NewPlayer(startX, startY, c.Game.CurrentLevelIndex)
	player.Name = playerName
	c.Game.Player = &player
	c.UpdateVisibility()
	c.Game.AddMessage("Start new game!")
	c.Game.AddMessage("Use WASD for action, q for exit")

	c.Game.StateGame = YouGame
}

func (controller *Controller) MovePlayer(dx int, dy int) {
	newX, newY := controller.Game.Player.PosX+dx, controller.Game.Player.PosY+dy

	tile := controller.Game.CurrentLevel.Tiles[newX][newY]
	for i := 0; i < len(controller.Game.CurrentLevel.Enemies); i++ {
		if controller.Game.CurrentLevel.Enemies[i].PosXEnemy == newX && controller.Game.CurrentLevel.Enemies[i].PosYEnemy == newY {
			controller.PlayerAttack(&controller.Game.CurrentLevel.Enemies[i])
			return
		}
	}
	if !tile.Blocked {
		controller.Game.Player.PosX = newX
		controller.Game.Player.PosY = newY
		controller.Game.Player.CountTile++

		if tile.Symbol == '%' {
			if controller.Game.CurrentLevelIndex < CountLevels-1 {
				controller.Game.UpgradeLevel()
			} else {
				controller.Game.StateGame = YouWin
				controller.SaveStatistics(true)
			}
		}
		controller.PickUpObject()
		controller.UpdateTiles()
		controller.UpdateVisibility()
	}
}

// меняем состояние игры, для этого нужно обновить тайлы. Когда добавляем новые объекты в поле, обновляем тайлы здесь
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

// UpdateVisibility обновляет видимые клетки на основе позиции игрока
func (c *Controller) UpdateVisibility() {
	level := c.Game.CurrentLevel
	playerX, playerY := c.Game.Player.PosX, c.Game.Player.PosY

	for x := 0; x < ScreenWidth; x++ {
		for y := 0; y < ScreenHeight; y++ {
			level.Visible[x][y] = false
		}
	}

	playerRoom := level.FindRoomContaining(playerX, playerY)
	radius := 8
	level.CalculateVisibility(playerX, playerY, radius)

	for x := 0; x < ScreenWidth; x++ {
		for y := 0; y < ScreenHeight; y++ {
			if playerRoom != nil && x >= playerRoom.X1 && x <= playerRoom.X2 && y >= playerRoom.Y1 && y <= playerRoom.Y2 {
				level.Visible[x][y] = true
				continue
			}

			distance := (x-playerX)*(x-playerX) + (y-playerY)*(y-playerY)
			if distance <= radius*radius && level.HasLineOfSight(playerX, playerY, x, y) { // радиус 3 клетки
				level.Visible[x][y] = true
			}
		}
	}
}

// область видимости врагов, меняет режим врага на преследование
func (controller *Controller) EnemyFOV() {
	for i := 0; i < len(controller.Game.CurrentLevel.Enemies); i++ {
		x1, x2, y1, y2 := controller.Game.CurrentLevel.Enemies[i].CurrentRoom.Interior()
		if controller.Game.Player.PosX >= x1 && controller.Game.Player.PosX <= x2 && controller.Game.Player.PosY >= y1 && controller.Game.Player.PosY <= y2 {
			if controller.isPlayerInRange(&controller.Game.CurrentLevel.Enemies[i]) {
				controller.Game.CurrentLevel.Enemies[i].Mode = Chasing
				controller.Game.AddMessage(fmt.Sprintf("%s is chasing you!", controller.Game.CurrentLevel.Enemies[i].TypeEnemy))
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
		// Проверяем, находится ли монстр рядом с игроком
		dx := math.Abs(float64(enemy.PosXEnemy - c.Game.Player.PosX))
		dy := math.Abs(float64(enemy.PosYEnemy - c.Game.Player.PosY))
		if (dx == 1 && dy == 0) || (dx == 0 && dy == 1) {
			// Монстр атакует игрока
			enemy.Attack(c.Game.Player, c)
			continue // Пропускаем перемещение, если монстр атакует
		}
		// Перемещаем врага
		newX, newY := enemy.PosXEnemy, enemy.PosYEnemy
		if enemy.Mode == Roaming {
			newX, newY = enemy.EnemyMove()
		} else if enemy.Mode == Chasing {
			newX, newY = enemy.ChaseTarget(c.Game.Player.PosX, c.Game.Player.PosY)
		}
		if !c.Game.CurrentLevel.Tiles[newX][newY].BlockedForEnemy {
			enemy.PosXEnemy = newX
			enemy.PosYEnemy = newY
			c.UpdateTiles()
			c.UpdateVisibility()
		}
	}
}

func (c *Controller) isPlayerInRange(enemy *Enemy) bool {
	dx := c.Game.Player.PosX - enemy.PosXEnemy
	dy := c.Game.Player.PosY - enemy.PosYEnemy
	distance := math.Sqrt(float64(dx*dx + dy*dy))
	return distance <= float64(enemy.HostilityEnemy)
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
	reachedLevel := c.Game.Player.CurrentLevelIndex

	playerName := c.Game.Player.Name
	if playerName == "" {
		playerName = "Player"
	}

	c.Game.SaveManager.AddStatistic(
		c.Game.Player.Name,
		reachedLevel,
		c.Game.Player.CountEnemy,
		c.Game.Player.Treasure,
		c.Game.Player.CountFood,
		c.Game.Player.CountElixir,
		c.Game.Player.CountScrollsRead,
		c.Game.Player.CountHits,
		c.Game.Player.TotalHitsTaken,
		c.Game.Player.CountTile,
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

	// Отображаем список еды
	c.Game.AddMessage("Choose food (1-" + strconv.Itoa(len(objects)) + "):")
	for i, object := range objects {
		c.Game.AddMessage(fmt.Sprintf("%d. %s (Health: %d)", i+1, GetObjectSymbol(object.TypeObject), object.Health))
	}

	// Устанавливаем режим ожидания выбора еды
	c.Game.WaitingForFoodChoice = true
}

func (c *Controller) UseElixir() {
	objects := c.Game.Player.Backpack.GetObjects(ELEXIR)
	if len(objects) == 0 {
		c.Game.AddMessage("No elixirs in backpack!")
		return
	}

	// Отображаем список эликсиров
	c.Game.AddMessage("Choose elixir (1-" + strconv.Itoa(len(objects)) + "):")
	for i, object := range objects {
		var effect string
		switch object.SubtypeObject {
		case DEXTERITY_ELIXIR:
			effect = fmt.Sprintf("Dexterity: %d", object.Dexterity)
		case STRENGTH_ELIXIR:
			effect = fmt.Sprintf("Strength: %d", object.Strength)
		}
		c.Game.AddMessage(fmt.Sprintf("%d. %s (%s)", i+1, GetObjectSymbol(object.TypeObject), effect))
	}

	// Устанавливаем режим ожидания выбора эликсира
	c.Game.WaitingForElixirChoice = true
}

func (c *Controller) UseScroll() {
	objects := c.Game.Player.Backpack.GetObjects(SCROL)
	if len(objects) == 0 {
		c.Game.AddMessage("No scrolls in backpack!")
		return
	}

	// Отображаем список свитков
	c.Game.AddMessage("Choose scroll (1-" + strconv.Itoa(len(objects)) + "):")
	for i, object := range objects {
		var effect string
		switch object.SubtypeObject {
		case DEXTERITY_SCROL:
			effect = fmt.Sprintf("Dexterity: %d", object.Dexterity)
		case STRENGTH_SCROL:
			effect = fmt.Sprintf("Strength: %d", object.Strength)
		case HEALTH_SCROL:
			effect = fmt.Sprintf("Max HP: %d", object.MaxHealth)
		}
		c.Game.AddMessage(fmt.Sprintf("%d. %s (%s)", i+1, GetObjectSymbol(object.TypeObject), effect))
	}

	// Устанавливаем режим ожидания выбора свитка
	c.Game.WaitingForScrollChoice = true
}

func (c *Controller) UseWeapon() {
	objects := c.Game.Player.Backpack.GetObjects(WEAPON)
	if len(objects) == 0 {
		c.Game.AddMessage("No weapons in backpack!")
		return
	}

	// Отображаем список оружия
	c.Game.AddMessage("Choose a weapon (0 to unequip, 1-" + strconv.Itoa(len(objects)) + "):")
	for i, object := range objects {
		c.Game.AddMessage(fmt.Sprintf("%d. %s (Damage: %s)", i+1, GetObjectSymbol(object.TypeObject), object.Damage))
	}

	// Устанавливаем режим ожидания выбора оружия
	c.Game.WaitingForWeaponChoice = true
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

// механизм временных эффектов и их истечения
func (c *Controller) DecreaseTemporaryEffects(player *Player) {
	if player.TemporaryEffects == nil {
		return
	}

	for effectType, duration := range player.TemporaryEffects {
		if duration <= 0 {
			delete(player.TemporaryEffects, effectType)
		} else {
			player.TemporaryEffects[effectType] = duration - 1
		}
	}

	// Если здоровье игрока упало до 0 или ниже, восстанавливаем до 1
	if player.HP <= 0 {
		player.HP = 1
	}
}

// метод для проверки вероятности попадания в зависимости от ловкости атакующего и цели
func (c *Controller) IsHit(attackerDexterity, targetDexterity int) bool {
	// Вероятность попадания зависит от ловкости
	hitChance := 70 + attackerDexterity - targetDexterity
	if hitChance > 90 {
		hitChance = 90
	} else if hitChance < 10 {
		hitChance = 10
	}
	return GeneratorNum(0, 100) < hitChance
}

// метод для расчета урона в зависимости от силы и оружи
func (c *Controller) CalculateDamage(attackerStrength int, weaponDamage string) int {
	baseDamage := attackerStrength / 2

	// Разбор урона оружия (например, "1d6")
	if weaponDamage != "1d1" {
		parts := strings.Split(weaponDamage, "d")
		if len(parts) == 2 {
			dice, _ := strconv.Atoi(parts[0])
			sides, _ := strconv.Atoi(parts[1])
			weaponDamageValue := 0
			for i := 0; i < dice; i++ {
				weaponDamageValue += GeneratorNum(1, sides)
			}
			baseDamage += weaponDamageValue
		}
	}

	return baseDamage
}

func (c *Controller) ApplyTemporaryEffect(player *Player, effectType string, value int, duration int) {
	if player.TemporaryEffects == nil {
		player.TemporaryEffects = make(map[string]int)
	}
	player.TemporaryEffects[effectType] = duration

	switch effectType {
	case "dexterity":
		player.Dexterity += value
	case "strength":
		player.Strength += value
	}
}

// PlayerAttack обрабатывает атаку игрока на врага с учетом всех особенностей
func (c *Controller) PlayerAttack(enemy *Enemy) {
	// Проверка попадания
	hitChance := 70 + c.Game.Player.Dexterity - enemy.DexterityEnemy
	if hitChance > 90 {
		hitChance = 90
	} else if hitChance < 10 {
		hitChance = 10
	}

	if GeneratorNum(0, 100) > hitChance {
		c.Game.AddMessage("You missed!")
		return
	}

	// Рассчитываем урон
	damage := c.Game.Player.Strength
	if c.Game.Player.CurrenWeapon != "1d1" {
		parts := strings.Split(c.Game.Player.CurrenWeapon, "d")
		if len(parts) == 2 {
			dice, _ := strconv.Atoi(parts[0])
			sides, _ := strconv.Atoi(parts[1])
			weaponDamageValue := 0
			for i := 0; i < dice; i++ {
				weaponDamageValue += GeneratorNum(1, sides)
			}
			damage += weaponDamageValue
		}
	}

	enemy.HealthEnemy -= damage
	c.Game.Player.CountHits++ // Увеличиваем количество нанесенных попаданий
	c.Game.AddMessage(fmt.Sprintf("You hit %s for %d damage!", enemy.TypeEnemy, damage))

	// Проверяем, умер ли враг
	if enemy.HealthEnemy <= 0 {
		c.Game.AddMessage(fmt.Sprintf("You defeated %s!", enemy.TypeEnemy))
		c.Game.Player.Treasure += enemy.Treasure
		c.Game.Player.CountEnemy++

		// Удаляем врага с уровня
		for i, e := range c.Game.CurrentLevel.Enemies {
			if e.PosXEnemy == enemy.PosXEnemy && e.PosYEnemy == enemy.PosYEnemy {
				c.Game.CurrentLevel.Enemies = append(c.Game.CurrentLevel.Enemies[:i], c.Game.CurrentLevel.Enemies[i+1:]...)
				break
			}
		}
		c.UpdateTiles()
	}
}

// функция для отображения таблицы лидеров
func (c *Controller) ShowLeaderboard() {
	c.Game.Renderer.ShowLeaderboard(c.Game)
}

func (c *Controller) HandleWeaponChoice(ch rune) {
	objects := c.Game.Player.Backpack.GetObjects(WEAPON)
	if len(objects) == 0 {
		c.Game.WaitingForWeaponChoice = false
		return
	}

	if ch >= '0' && ch <= '9' {
		choice, err := strconv.Atoi(string(ch))
		if err != nil {
			c.Game.AddMessage("Invalid choice!")
			c.Game.WaitingForWeaponChoice = false
			return
		}

		if choice == 0 {
			c.Game.Player.CurrenWeapon = "1d1"
			c.Game.AddMessage("Unequipped weapon")
			c.Game.WaitingForWeaponChoice = false
			return
		}

		if choice < 1 || choice > len(objects) {
			c.Game.AddMessage("Invalid choice!")
			c.Game.WaitingForWeaponChoice = false
			return
		}

		selectedObject, _ := c.Game.Player.Backpack.UseObject(WEAPON, choice-1)
		if selectedObject != nil {
			// Если у игрока уже было оружие, бросаем его на пол
			if c.Game.Player.CurrenWeapon != "1d1" {
				neighbor := []Tile{
					c.Game.CurrentLevel.Tiles[c.Game.Player.PosX+1][c.Game.Player.PosY],
					c.Game.CurrentLevel.Tiles[c.Game.Player.PosX][c.Game.Player.PosY+1],
					c.Game.CurrentLevel.Tiles[c.Game.Player.PosX-1][c.Game.Player.PosY],
					c.Game.CurrentLevel.Tiles[c.Game.Player.PosX][c.Game.Player.PosY-1],
				}
				for _, tileNeighbor := range neighbor {
					if !tileNeighbor.Blocked {
						NewPosX, NewPosY := tileNeighbor.PosX, tileNeighbor.PosY
						oldWeapon := &Object{
							TypeObject: WEAPON,
							Damage:     c.Game.Player.CurrenWeapon,
							PosX:       NewPosX,
							PosY:       NewPosY,
						}
						c.Game.CurrentLevel.Objects = append(c.Game.CurrentLevel.Objects, oldWeapon)
						c.Game.AddMessage(fmt.Sprintf("Dropped old weapon %s", oldWeapon.Damage))
						break
					}
				}
			}
			c.Game.Player.CurrenWeapon = selectedObject.Damage
			c.Game.AddMessage(fmt.Sprintf("Equipped %s (Damage: %s)", GetObjectSymbol(selectedObject.TypeObject), selectedObject.Damage))
			c.Game.WaitingForWeaponChoice = false
			return
		}
	} else {
		c.Game.AddMessage("Invalid choice!")
	}
	c.Game.WaitingForWeaponChoice = false
}

func (c *Controller) HandleFoodChoice(ch rune) {
	objects := c.Game.Player.Backpack.GetObjects(FOOD)
	if len(objects) == 0 {
		c.Game.WaitingForFoodChoice = false
		return
	}

	if ch >= '1' && ch <= '9' {
		choice, err := strconv.Atoi(string(ch))
		if err != nil {
			c.Game.AddMessage("Invalid choice!")
			c.Game.WaitingForFoodChoice = false
			return
		}

		if choice < 1 || choice > len(objects) {
			c.Game.AddMessage("Invalid choice!")
			c.Game.WaitingForFoodChoice = false
			return
		}

		selectedObject, _ := c.Game.Player.Backpack.UseObject(FOOD, choice-1)
		if selectedObject != nil {
			c.Game.Player.HP += selectedObject.Health
			if c.Game.Player.HP > c.Game.Player.MaxHP {
				c.Game.Player.HP = c.Game.Player.MaxHP
			}
			c.Game.Player.CountFood++
			c.Game.AddMessage(fmt.Sprintf("Ate food. HP: %d/%d", c.Game.Player.HP, c.Game.Player.MaxHP))
			c.Game.WaitingForFoodChoice = false
			return
		}
	} else {
		c.Game.AddMessage("Invalid choice!")
	}
	c.Game.WaitingForFoodChoice = false
}

func (c *Controller) HandleElixirChoice(ch rune) {
	objects := c.Game.Player.Backpack.GetObjects(ELEXIR)
	if len(objects) == 0 {
		c.Game.WaitingForElixirChoice = false
		return
	}

	if ch >= '1' && ch <= '9' {
		choice, err := strconv.Atoi(string(ch))
		if err != nil {
			c.Game.AddMessage("Invalid choice!")
			c.Game.WaitingForElixirChoice = false
			return
		}

		if choice < 1 || choice > len(objects) {
			c.Game.AddMessage("Invalid choice!")
			c.Game.WaitingForElixirChoice = false
			return
		}

		selectedObject, _ := c.Game.Player.Backpack.UseObject(ELEXIR, choice-1)
		if selectedObject != nil {
			switch selectedObject.SubtypeObject {
			case DEXTERITY_ELIXIR:
				c.Game.Player.Dexterity += selectedObject.Dexterity
				c.Game.AddMessage(fmt.Sprintf("Drank Dexterity Elixir. Dexterity: %d", c.Game.Player.Dexterity))
			case STRENGTH_ELIXIR:
				c.Game.Player.Strength += selectedObject.Strength
				c.Game.AddMessage(fmt.Sprintf("Drank Strength Elixir. Strength: %d", c.Game.Player.Strength))
			}
			c.Game.Player.CountElixir++
			c.Game.WaitingForElixirChoice = false
			return
		}
	} else {
		c.Game.AddMessage("Invalid choice!")
	}
	c.Game.WaitingForElixirChoice = false
}

func (c *Controller) HandleScrollChoice(ch rune) {
	objects := c.Game.Player.Backpack.GetObjects(SCROL)
	if len(objects) == 0 {
		c.Game.WaitingForScrollChoice = false
		return
	}

	if ch >= '1' && ch <= '9' {
		choice, err := strconv.Atoi(string(ch))
		if err != nil {
			c.Game.AddMessage("Invalid choice!")
			c.Game.WaitingForScrollChoice = false
			return
		}

		if choice < 1 || choice > len(objects) {
			c.Game.AddMessage("Invalid choice!")
			c.Game.WaitingForScrollChoice = false
			return
		}

		selectedObject, _ := c.Game.Player.Backpack.UseObject(SCROL, choice-1)
		if selectedObject != nil {
			switch selectedObject.SubtypeObject {
			case DEXTERITY_SCROL:
				c.Game.Player.Dexterity += selectedObject.Dexterity
				c.Game.AddMessage(fmt.Sprintf("Read Dexterity Scroll. Dexterity: %d", c.Game.Player.Dexterity))
			case STRENGTH_SCROL:
				c.Game.Player.Strength += selectedObject.Strength
				c.Game.AddMessage(fmt.Sprintf("Read Strength Scroll. Strength: %d", c.Game.Player.Strength))
			case HEALTH_SCROL:
				c.Game.Player.MaxHP += selectedObject.MaxHealth
				c.Game.Player.HP += selectedObject.MaxHealth
				c.Game.AddMessage(fmt.Sprintf("Read Health Scroll. Max HP: %d, HP: %d", c.Game.Player.MaxHP, c.Game.Player.HP))
			}
			c.Game.Player.CountScrollsRead++
			c.Game.WaitingForScrollChoice = false
			return
		}
	} else {
		c.Game.AddMessage("Invalid choice!")
	}
	c.Game.WaitingForScrollChoice = false
}
