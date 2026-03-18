package main

import (
	"fmt"
)

type Enemy struct {
	PosXEnemy      int
	PosYEnemy      int
	TypeEnemy      string
	HealthEnemy    int
	DexterityEnemy int
	StrengthEnemy  int
	HostilityEnemy int
	CurrentRoom    *Room
	Mode           int
	IsVisible      bool
	IsSleeping     bool
	Cooldown       int
	Treasure       int
}

// типов врагов
const (
	Zombie    string = "zombie"
	Vampire   string = "vampire"
	Ghost     string = "ghost"
	Ogre      string = "ogre"
	SnakeMage string = "snakeMage"
)

const ( // шанс появления врага
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
	LoWHostility    int = 4
	MiddleHostility int = 6
	HighHostility   int = 8
)

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
		enemy.Treasure = 10
	case Vampire:
		enemy.HealthEnemy = 25
		enemy.DexterityEnemy = 15
		enemy.StrengthEnemy = 10
		enemy.HostilityEnemy = 7
		enemy.Treasure = 20
	case Ghost:
		enemy.HealthEnemy = 10
		enemy.DexterityEnemy = 20
		enemy.StrengthEnemy = 5
		enemy.HostilityEnemy = 10
		enemy.Treasure = 15
	case Ogre:
		enemy.HealthEnemy = 40
		enemy.DexterityEnemy = 5
		enemy.StrengthEnemy = 15
		enemy.HostilityEnemy = 5
		enemy.Treasure = 25
	case SnakeMage:
		enemy.HealthEnemy = 20
		enemy.DexterityEnemy = 25
		enemy.StrengthEnemy = 5
		enemy.HostilityEnemy = 8
		enemy.Treasure = 30

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
		Zombie:    30 + level.Number*2,
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

func (en *Enemy) EnemyMove() (int, int) {
	newX, newY := en.PosXEnemy, en.PosYEnemy
	switch en.TypeEnemy {
	case Zombie, Vampire:
		x1, x2, y1, y2 := en.CurrentRoom.Interior()
		direction := GeneratorNum(1, 4)
		if direction == 1 && en.PosXEnemy < x2 {
			newX = en.PosXEnemy + 1
		} else if direction == 2 && en.PosYEnemy < y2 {
			newY = en.PosYEnemy + 1
		} else if direction == 3 && en.PosXEnemy > x1 {
			newX = en.PosXEnemy - 1
		} else if direction == 4 && en.PosYEnemy > y1 {
			newY = en.PosYEnemy - 1
		}
	case Ghost:
		newX, newY = en.CurrentRoom.RandomPos()
	case Ogre:
		if en.Cooldown > 0 {
			en.Cooldown--
			return en.PosXEnemy, en.PosYEnemy
		}
		x1, x2, y1, y2 := en.CurrentRoom.Interior()
		direction := GeneratorNum(1, 4)
		if direction == 1 && en.PosXEnemy < x2-1 {
			newX = en.PosXEnemy + 2
		} else if direction == 2 && en.PosYEnemy < y2-1 {
			newY = en.PosYEnemy + 2
		} else if direction == 3 && en.PosXEnemy > x1+1 {
			newX = en.PosXEnemy - 2
		} else if direction == 4 && en.PosYEnemy > y1+1 {
			newY = en.PosYEnemy - 2
		}
	case SnakeMage:
		x1, x2, y1, y2 := en.CurrentRoom.Interior()
		direction := GeneratorNum(1, 4)
		if direction == 1 && en.PosXEnemy < x2 && en.PosYEnemy > y1 {
			newX = en.PosXEnemy + 1
			newY = en.PosYEnemy - 1
		} else if direction == 2 && en.PosXEnemy < x2 && en.PosYEnemy < y2 {
			newX = en.PosXEnemy + 1
			newY = en.PosYEnemy + 1
		} else if direction == 3 && en.PosXEnemy > x1 && en.PosYEnemy > y1 {
			newX = en.PosXEnemy - 1
			newY = en.PosYEnemy - 1
		} else if direction == 4 && en.PosXEnemy > x1 && en.PosYEnemy < y2 {
			newX = en.PosXEnemy - 1
			newY = en.PosYEnemy + 1
		}
	}
	return newX, newY
}

func (en *Enemy) ChaseTarget(TargetX int, TargetY int) (int, int) {
	newX, newY := en.PosXEnemy, en.PosYEnemy
	lenX := en.PosXEnemy - TargetX
	lenY := en.PosYEnemy - TargetY
	singX := 1
	singY := 1
	if lenX < 0 {
		lenX *= (-1)
		singX = -1
	} else if lenY < 0 {
		lenY *= (-1)
		singY = -1
	}

	if lenX < lenY {
		if singY == -1 { // двигаемся вниз
			switch en.TypeEnemy {
			case Zombie, Vampire:
				newY = en.PosYEnemy + 1
			case Ghost:
				newX, newY = TargetX, TargetY-1
			case Ogre:
				newY = en.PosYEnemy + 2
			case SnakeMage:
				if singX == -1 { // если игрок справа от врага
					newX = en.PosXEnemy + 1
					newY = en.PosYEnemy + 1
				} else {
					newX = en.PosXEnemy + 1
					newY = en.PosYEnemy - 1
				}
			}
		} else { // двигаемся вверх
			switch en.TypeEnemy {
			case Zombie, Vampire:
				newY = en.PosYEnemy - 1
			case Ghost:
				newX, newY = TargetX, TargetY+1
			case Ogre:
				newY = en.PosYEnemy - 2
			case SnakeMage:
				if singX == -1 { // если игрок справа от врага
					newX = en.PosXEnemy - 1
					newY = en.PosYEnemy + 1
				} else {
					newX = en.PosXEnemy - 1
					newY = en.PosYEnemy - 1
				}
			}
		}
	} else { // двигаемся в горизонтальном направлении
		if singX == -1 { // двигаемся вправо
			switch en.TypeEnemy {
			case Zombie, Vampire:
				newX = en.PosXEnemy + 1
			case Ghost:
				newX, newY = TargetX-1, TargetY
			case Ogre:
				newX = en.PosXEnemy + 2
			case SnakeMage:
				if singY == -1 { // если игрок снизу от врага
					newX = en.PosXEnemy + 1
					newY = en.PosYEnemy + 1
				} else {
					newX = en.PosXEnemy - 1
					newY = en.PosYEnemy + 1
				}
			}
		} else { // двигаемся влево
			switch en.TypeEnemy {
			case Zombie, Vampire:
				newX = en.PosXEnemy - 1
			case Ghost:
				newX, newY = TargetX+1, TargetY
			case Ogre:
				newX = en.PosXEnemy - 2
			case SnakeMage:
				if singY == -1 { // если игрок снизу от врага
					newX = en.PosXEnemy + 1
					newY = en.PosYEnemy - 1
				} else {
					newX = en.PosXEnemy - 1
					newY = en.PosYEnemy - 1
				}
			}
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

// Attack атакует игрока, учитывая тип врага и его уникальные способности
func (en *Enemy) Attack(player *Player, controller *Controller) bool {
	hitChance := 70 + en.DexterityEnemy - player.Dexterity
	if hitChance > 90 {
		hitChance = 90
	} else if hitChance < 10 {
		hitChance = 10
	}

	if GeneratorNum(0, 100) > hitChance {
		controller.Game.AddMessage(fmt.Sprintf("%s missed!", en.TypeEnemy))
		return false
	}

	// Уникальные особенности для каждого типа врага при атаке
	switch en.TypeEnemy {
	case Zombie:
		if GeneratorNum(0, 100) < 10 { // 10% шанс
			player.MaxHP -= 1
			if player.MaxHP < 1 {
				player.MaxHP = 1
			}
			controller.Game.AddMessage("Zombie infected you, reducing your max HP!")
		}
	case Vampire:
		// Первая атака вампира всегда промахивается
		if en.Mode != Fight {
			controller.Game.AddMessage("Vampire's first attack missed!")
			en.Mode = Fight
			return false
		}
		// Уменьшаем максимальное здоровье игрока
		player.MaxHP -= 1
		if player.MaxHP < 1 {
			player.MaxHP = 1
		}
		controller.Game.AddMessage("Vampire drained your max HP!")
	case Ghost:
		// Призраки могут стать невидимыми после атаки
		if GeneratorNum(0, 100) < 50 { // 50% шанс
			en.IsVisible = false
			controller.Game.AddMessage("Ghost turned invisible!")
		}
	case Ogre:
		// Огры отдыхают один ход после атаки
		en.Cooldown = 1
		controller.Game.AddMessage("Ogre is resting after attack!")
	case SnakeMage:
		// Змееволги могут усыпить игрока
		if GeneratorNum(0, 100) < 30 { // 30% шанс
			player.IsSleeping = true
			controller.Game.AddMessage("Snake Mage put you to sleep!")
		}
	}

	player.TotalHitsTaken++

	// Рассчитываем урон
	damage := en.StrengthEnemy + GeneratorNum(0, 5)
	player.HP -= damage
	if player.HP < 0 {
		player.HP = 0
	}

	controller.Game.Player.CountHits++
	controller.Game.AddMessage(fmt.Sprintf("%s hit you for %d damage!", en.TypeEnemy, damage))

	// Проверяем, умер ли игрок
	if player.HP <= 0 {
		controller.Game.StateGame = YouLose
		controller.Game.AddMessage("You died!")
	}

	return true
}
