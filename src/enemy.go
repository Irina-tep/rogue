package main

type Enemy struct {
	PosXEnemy      int
	PosYEnemy      int
	TypeEnemy      string // Тип врага: Zombie, Vampire, Ghost, Ogre, SnakeMage
	HealthEnemy    int    // Здоровье
	DexterityEnemy int    // Ловкость
	StrengthEnemy  int    // Сила
	HostilityEnemy int    // Враждебность (радиус преследования)
	CurrentRoom    *Room  // Текущая комната
	Mode           int    // Режим: Roaming, Chasing, Fight
	IsVisible      bool   // Видимость (для призраков)
	IsSleeping     bool   // Состояние сна (для змееволгов)
	Cooldown       int    // Задержка перед атакой (для огров)
}

// типов врагов
const (
	Zombie    string = "zombie"    // Зомби
	Vampire   string = "vampire"   // Вампир
	Ghost     string = "ghost"     // Призрак
	Ogre      string = "ogre"      // Огр
	SnakeMage string = "snakeMage" // Змееволг
)

const ( // шанс появления врага, потом будем менять вероятность появления врага в комнате в зависимости от уровня, так как По мере того, как игрок переходит на каждый новый уровень:
	// количество и сложность врагов увеличиваются
	ChanceZombie    int = 30
	ChanceVampire   int = 25
	ChanceGhost     int = 20
	ChanceOgre      int = 15
	ChanceSnakeMage int = 10
)

// режим врага
const (
	Roaming     = 1 // режим врага 1 - свободное движение врага, 2 - преследование игрока, 3 -режим боя
	Chasing     = 2
	Fight   int = 3
)

// враждебность
const (
	LoWHostility    int = 4 // низкая враждебность
	MiddleHostility int = 6 // средняя враждеюность
	HighHostility   int = 8 // высокая враждебность
)

// характеристики врагов будут меняться, пока поставлены просто так
func NewEnemy(coordXEnemy, coordYEnemy int, typeEnemy string, currentRoom *Room) Enemy {
	enemy := Enemy{
		PosXEnemy:   coordXEnemy,
		PosYEnemy:   coordYEnemy,
		TypeEnemy:   typeEnemy,
		CurrentRoom: currentRoom,
		Mode:        Roaming,
		IsVisible:   true,
		IsSleeping:  false,
		Cooldown:    0,
	}

	switch typeEnemy {
	case Zombie:
		enemy.HealthEnemy = 30
		enemy.DexterityEnemy = 5
		enemy.StrengthEnemy = 10
		enemy.HostilityEnemy = 5
	case Vampire:
		enemy.HealthEnemy = 25
		enemy.DexterityEnemy = 15
		enemy.StrengthEnemy = 10
		enemy.HostilityEnemy = 7
	case Ghost:
		enemy.HealthEnemy = 10
		enemy.DexterityEnemy = 20
		enemy.StrengthEnemy = 5
		enemy.HostilityEnemy = 3
	case Ogre:
		enemy.HealthEnemy = 40
		enemy.DexterityEnemy = 5
		enemy.StrengthEnemy = 20
		enemy.HostilityEnemy = 5
	case SnakeMage:
		enemy.HealthEnemy = 20
		enemy.DexterityEnemy = 25
		enemy.StrengthEnemy = 10
		enemy.HostilityEnemy = 8
	}

	return enemy
}

type EnemyType struct {
	Name   string
	Chance int
}

// генерация врагов в комнатах с учетом шанса появления и сложности по уровням
func (level *Level) CreateEnemies() {
	enemyChances := map[string]int{
		Zombie:    30 + level.Number*2, // Увеличиваем шанс появления зомби с уровнем
		Vampire:   25 + level.Number*3,
		Ghost:     20 + level.Number*1,
		Ogre:      15 + level.Number*2,
		SnakeMage: 10 + level.Number*2,
	}

	for i := 1; i < amountRoom; i++ {
		for enemyType, chance := range enemyChances {
			if GeneratorNum(0, 100) < chance {
				x, y := level.Rooms[i].RandomPos()
				currentRoom := &level.Rooms[i]
				enemy := NewEnemy(x, y, enemyType, currentRoom)
				level.Enemies = append(level.Enemies, enemy)
			}
		}
	}
}

// func (en *Enemy) EnemyMove() (int, int) {
// 	newX, newY := en.PosXEnemy, en.PosYEnemy
// 	switch en.TypeEnemy {
// 	case Zombie, Vampire:
// 		x1, x2, y1, y2 := en.CurrentRoom.Interior()
// 		// генерируется случайное направление движения на 1 клетку
// 		// 1 вправо, 2 вниз, 3 влево, 4 вверх
// 		direction := GeneratorNum(1, 4)
// 		if direction == 1 && en.PosXEnemy < x2 {
// 			newX = en.PosXEnemy + 1
// 		} else if direction == 2 && en.PosYEnemy < y2 {
// 			newY = en.PosYEnemy + 1
// 		} else if direction == 3 && en.PosXEnemy > x1 {
// 			newX = en.PosXEnemy - 1
// 		} else if direction == 4 && en.PosYEnemy > y1 {
// 			newY = en.PosYEnemy - 1
// 		}
// 	case Ghost:
// 		newX, newY = en.CurrentRoom.RandomPos()

// 	case Ogre:
// 		x1, x2, y1, y2 := en.CurrentRoom.Interior()
// 		// генерируется случайное направление движения на 1 клетку
// 		direction := GeneratorNum(1, 4)
// 		if direction == 1 && en.PosXEnemy < x2-1 {
// 			newX = en.PosXEnemy + 2
// 		} else if direction == 2 && en.PosYEnemy < y2-1 {
// 			newY = en.PosYEnemy + 2
// 		} else if direction == 3 && en.PosXEnemy > x1+1 {
// 			newX = en.PosXEnemy - 2
// 		} else if direction == 4 && en.PosYEnemy > y1+1 {
// 			newY = en.PosYEnemy - 2
// 		}

// 	case SnakeMage:
// 		// Добавьте логику для SnakeMage
// 		x1, x2, y1, y2 := en.CurrentRoom.Interior()
// 		direction := GeneratorNum(1, 4)
// 		if direction == 1 && en.PosXEnemy < x2 && en.PosYEnemy > y1 {
// 			newX = en.PosXEnemy + 1
// 			newY = en.PosYEnemy - 1
// 		} else if direction == 2 && en.PosXEnemy < x2 && en.PosYEnemy < y2 {
// 			newX = en.PosXEnemy + 1
// 			newY = en.PosYEnemy + 1
// 		} else if direction == 3 && en.PosXEnemy > x1 && en.PosYEnemy > y1 {
// 			newX = en.PosXEnemy - 1
// 			newY = en.PosYEnemy - 1
// 		} else if direction == 4 && en.PosXEnemy > x1 && en.PosYEnemy < y2 {
// 			newX = en.PosXEnemy - 1
// 			newY = en.PosYEnemy + 1
// 		}
// 	}
// 	return newX, newY
// }

func (en *Enemy) EnemyMove(level *Level) (int, int) {
	newX, newY := en.PosXEnemy, en.PosYEnemy

	switch en.TypeEnemy {
	case Zombie, Vampire:
		x1, x2, y1, y2 := en.CurrentRoom.Interior()
		direction := GeneratorNum(1, 4)
		if direction == 1 && newX < x2 && !en.isBlocked(level, newX+1, newY) {
			newX = en.PosXEnemy + 1
		} else if direction == 2 && newY < y2 && !en.isBlocked(level, newX, newY+1) {
			newY = en.PosYEnemy + 1
		} else if direction == 3 && newX > x1 && !en.isBlocked(level, newX-1, newY) {
			newX = en.PosXEnemy - 1
		} else if direction == 4 && newY > y1 && !en.isBlocked(level, newX, newY-1) {
			newY = en.PosYEnemy - 1
		}

	case Ghost:
		if GeneratorNum(0, 1) == 0 {
			en.IsVisible = !en.IsVisible
		}
		if en.IsVisible {
			newX, newY = en.CurrentRoom.RandomPos()
		}

	case Ogre:
		if en.Cooldown > 0 {
			en.Cooldown--
			return en.PosXEnemy, en.PosYEnemy
		}
		x1, x2, y1, y2 := en.CurrentRoom.Interior()
		direction := GeneratorNum(1, 4)
		if direction == 1 && newX < x2-1 && !en.isBlocked(level, newX+2, newY) {
			newX = en.PosXEnemy + 2
		} else if direction == 2 && newY < y2-1 && !en.isBlocked(level, newX, newY+2) {
			newY = en.PosYEnemy + 2
		} else if direction == 3 && newX > x1+1 && !en.isBlocked(level, newX-2, newY) {
			newX = en.PosXEnemy - 2
		} else if direction == 4 && newY > y1+1 && !en.isBlocked(level, newX, newY-2) {
			newY = en.PosYEnemy - 2
		}

	case SnakeMage:
		x1, x2, y1, y2 := en.CurrentRoom.Interior()
		direction := GeneratorNum(1, 4)
		if direction == 1 && newX < x2 && newY > y1 && !en.isBlocked(level, newX+1, newY-1) {
			newX = en.PosXEnemy + 1
			newY = en.PosYEnemy - 1
		} else if direction == 2 && newX < x2 && newY < y2 && !en.isBlocked(level, newX+1, newY+1) {
			newX = en.PosXEnemy + 1
			newY = en.PosYEnemy + 1
		} else if direction == 3 && newX > x1 && newY > y1 && !en.isBlocked(level, newX-1, newY-1) {
			newX = en.PosXEnemy - 1
			newY = en.PosYEnemy - 1
		} else if direction == 4 && newX > x1 && newY < y2 && !en.isBlocked(level, newX-1, newY+1) {
			newX = en.PosXEnemy - 1
			newY = en.PosYEnemy + 1
		}
	}

	return newX, newY
}

func (en *Enemy) ChaseTarget(level *Level, targetX, targetY int) (int, int) {
	newX, newY := en.PosXEnemy, en.PosYEnemy

	switch en.TypeEnemy {
	case Zombie, Vampire:
		if targetX > en.PosXEnemy && !en.isBlocked(level, en.PosXEnemy+1, en.PosYEnemy) {
			newX = en.PosXEnemy + 1
		} else if targetX < en.PosXEnemy && !en.isBlocked(level, en.PosXEnemy-1, en.PosYEnemy) {
			newX = en.PosXEnemy - 1
		}

		if targetY > en.PosYEnemy && !en.isBlocked(level, en.PosXEnemy, en.PosYEnemy+1) {
			newY = en.PosYEnemy + 1
		} else if targetY < en.PosYEnemy && !en.isBlocked(level, en.PosXEnemy, en.PosYEnemy-1) {
			newY = en.PosYEnemy - 1
		}

	case Ghost:
		if GeneratorNum(0, 1) == 0 {
			newX, newY = targetX, targetY
		}

	case Ogre:
		if en.Cooldown > 0 {
			en.Cooldown--
			return en.PosXEnemy, en.PosYEnemy
		}
		if targetX > en.PosXEnemy && !en.isBlocked(level, en.PosXEnemy+2, en.PosYEnemy) {
			newX = en.PosXEnemy + 2
		} else if targetX < en.PosXEnemy && !en.isBlocked(level, en.PosXEnemy-2, en.PosYEnemy) {
			newX = en.PosXEnemy - 2
		}

		if targetY > en.PosYEnemy && !en.isBlocked(level, en.PosXEnemy, en.PosYEnemy+2) {
			newY = en.PosYEnemy + 2
		} else if targetY < en.PosYEnemy && !en.isBlocked(level, en.PosXEnemy, en.PosYEnemy-2) {
			newY = en.PosYEnemy - 2
		}

	case SnakeMage:
		if targetX > en.PosXEnemy && targetY > en.PosYEnemy && !en.isBlocked(level, en.PosXEnemy+1, en.PosYEnemy+1) {
			newX = en.PosXEnemy + 1
			newY = en.PosYEnemy + 1
		} else if targetX > en.PosXEnemy && targetY < en.PosYEnemy && !en.isBlocked(level, en.PosXEnemy+1, en.PosYEnemy-1) {
			newX = en.PosXEnemy + 1
			newY = en.PosYEnemy - 1
		} else if targetX < en.PosXEnemy && targetY > en.PosYEnemy && !en.isBlocked(level, en.PosXEnemy-1, en.PosYEnemy+1) {
			newX = en.PosXEnemy - 1
			newY = en.PosYEnemy + 1
		} else if targetX < en.PosXEnemy && targetY < en.PosYEnemy && !en.isBlocked(level, en.PosXEnemy-1, en.PosYEnemy-1) {
			newX = en.PosXEnemy - 1
			newY = en.PosYEnemy - 1
		}
	}

	return newX, newY
}

// Метод для проверки, заблокирована ли клетка
func (en *Enemy) isBlocked(level *Level, x, y int) bool {
	if x < 0 || x >= ScreenWidth || y < 0 || y >= ScreenHeight {
		return true
	}
	return en.CurrentRoom.IsBlocked(level, x, y)
}

func (en *Enemy) Attack(player *Player) {
	switch en.TypeEnemy {
	case Vampire:
		if GeneratorNum(0, 1) == 0 {
			player.MaxHP -= 1
			if player.MaxHP < 1 {
				player.MaxHP = 1
			}
		}
	case SnakeMage:
		if GeneratorNum(0, 1) == 0 {
			player.IsSleeping = true
		}
	case Ogre:
		en.Cooldown = 1
	}

	// Рассчитываем урон
	damage := en.StrengthEnemy + GeneratorNum(0, 5)
	player.HP -= damage
	if player.HP < 0 {
		player.HP = 0
	}
}

// func (en *Enemy) ChaseTarget(TargetX int, TargetY int) (int, int) {
// 	newX, newY := en.PosXEnemy, en.PosYEnemy
// 	lenX := en.PosXEnemy - TargetX
// 	lenY := en.PosYEnemy - TargetY
// 	singX := 1 // сохраняем знак
// 	singY := 1 // сохраняем знак
// 	// x1, x2, y1, y2 := en.CurrentRoom.Interior()  //не знаю нужно ли проверять на выход из комнаты
// 	if lenX < 0 {
// 		lenX *= (-1)
// 		singX = -1
// 	} else if lenY < 0 {
// 		lenY *= (-1)
// 		singY = -1
// 	}

// 	if lenX < lenY { // двигаемся в вертик направлении

// 		if singY == -1 { // двигаемся вниз
// 			switch en.TypeEnemy {
// 			case Zombie, Vampire:
// 				newY = en.PosYEnemy + 1
// 			case Ghost:
// 				newX, newY = TargetX, TargetY-1
// 			case Ogre:
// 				newY = en.PosYEnemy + 2
// 			case SnakeMage:
// 				if singX == -1 { // если игрок справа от врага
// 					newX = en.PosXEnemy + 1
// 					newY = en.PosYEnemy + 1
// 				} else {
// 					newX = en.PosXEnemy + 1
// 					newY = en.PosYEnemy - 1
// 				}
// 			}
// 		} else { // двигаемся вверх
// 			switch en.TypeEnemy {
// 			case Zombie, Vampire:
// 				newY = en.PosYEnemy - 1
// 			case Ghost:
// 				newX, newY = TargetX, TargetY+1
// 			case Ogre:
// 				newY = en.PosYEnemy - 2
// 			case SnakeMage:
// 				if singX == -1 { // если игрок справа от врага
// 					newX = en.PosXEnemy - 1
// 					newY = en.PosYEnemy + 1
// 				} else {
// 					newX = en.PosXEnemy - 1
// 					newY = en.PosYEnemy - 1
// 				}
// 			}
// 		}
// 	} else { // двигаемся в горизонтальном направлении
// 		if singX == -1 { // двигаемся вправо
// 			switch en.TypeEnemy {
// 			case Zombie, Vampire:
// 				newX = en.PosXEnemy + 1
// 			case Ghost:
// 				newX, newY = TargetX-1, TargetY
// 			case Ogre:
// 				newX = en.PosXEnemy + 2
// 			case SnakeMage:
// 				if singY == -1 { // если игрок снизу от врага
// 					newX = en.PosXEnemy + 1
// 					newY = en.PosYEnemy + 1
// 				} else {
// 					newX = en.PosXEnemy - 1
// 					newY = en.PosYEnemy + 1
// 				}
// 			}
// 		} else { // двигаемся влево
// 			switch en.TypeEnemy {
// 			case Zombie, Vampire:
// 				newX = en.PosXEnemy - 1
// 			case Ghost:
// 				newX, newY = TargetX+1, TargetY
// 			case Ogre:
// 				newX = en.PosXEnemy - 2
// 			case SnakeMage:
// 				if singY == -1 { // если игрок снизу от врага
// 					newX = en.PosXEnemy + 1
// 					newY = en.PosYEnemy - 1
// 				} else {
// 					newX = en.PosXEnemy - 1
// 					newY = en.PosYEnemy - 1
// 				}
// 			}
// 		}
// 	}
// 	return newX, newY
// }

// // функция объекта — это способность к перемещению.
// func (player *Player) Move(dx int, dy int) {
// 	player.PosX += dx
// 	player.PosY += dy
// }

// При убийстве врага вызвать метод AddTreasure, чтобы получить сокровище

// if enemy.HealthEnemy <= 0 {
//     controller.AddTreasure(enemy.Treasure)
//     // Удаляем врага с уровня
//     // ...
// }
