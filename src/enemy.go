package main

type Enemy struct {
	PosXEnemy      int
	PosYEnemy      int
	TypeEnemy      string //тип
	HealthEnemy    int    //здоровье,
	DexterityEnemy int    //ловкость
	StrengthEnemy  int    //сила
	HostilityEnemy int    //враждебность Атрибут враждебности определяет расстояние, с которого противник начинает преследовать игрока
	CurrentRoom    *Room  //ссылка на текущую комнату, так как игрок не может выйти за пределы комнаты
	Mode           int    //режим врага 1 - свободное движение врага в соответствии с своей схемой движения, 2 - преследование игрока, 3 -режим боя
	Treasure       int    //сколько сокровищ свтоит кадый враг при уничтожении
}

// типов врагов
const (
	Zombie    string = "zombie"    //Зомби
	Vampire   string = "vampire"   //Вампир
	Ghost     string = "ghost"     //Призрак
	Ogre      string = "ogre"      //Огр
	SnakeMage string = "snakeMage" //Змееволг
)

const ( //шанс появления врага, потом будем менять вероятность появления врага в комнате в зависимости от уровня, так как По мере того, как игрок переходит на каждый новый уровень:
	// количество и сложность врагов увеличиваются
	ChanceZombie    int = 30
	ChanceVampire   int = 25
	ChanceGhost     int = 20
	ChanceOgre      int = 15
	ChanceSnakeMage int = 10
)

// режим врага
const (
	Roaming     = 1 //режим врага 1 - свободное движение врага, 2 - преследование игрока, 3 -режим боя
	Chasing     = 2
	Fight   int = 3
)

// враждебность
const (
	LoWHostility    int = 4 //низкая враждебность
	MiddleHostility int = 6 //средняя враждеюность
	HighHostility   int = 8 //высокая враждебность
)

// характеристики врагов будут меняться, пока поставлены просто так
func NewEnemy(coordXEnemy int, coordYEnemy int, typeEnemy string, currentRoom *Room) Enemy {
	enemy := Enemy{
		PosXEnemy:      coordXEnemy,
		PosYEnemy:      coordYEnemy,
		TypeEnemy:      typeEnemy,
		HealthEnemy:    5,
		DexterityEnemy: 5,
		StrengthEnemy:  5,
		HostilityEnemy: 0,
		CurrentRoom:    currentRoom,
		Mode:           Roaming,
		Treasure:       5, //пока для всех одинаковое вознаграждение за убиство
	}
	if enemy.TypeEnemy == Ghost {
		enemy.HostilityEnemy = LoWHostility
	} else if enemy.TypeEnemy == Zombie || enemy.TypeEnemy == Ogre {
		enemy.HostilityEnemy = MiddleHostility
	} else {
		enemy.HostilityEnemy = HighHostility
	}
	return enemy
}

type EnemyType struct {
	Name   string
	Chance int
}

// генерация врагов в комнатах. В дальнейшем сюда нужно будет учитывать шанс появдения и сложность по уровням
func (level *Level) CreateEnemies() {
	var enemyTypes = []EnemyType{
		{Zombie, ChanceZombie},
		{Vampire, ChanceVampire},
		{Ghost, ChanceGhost},
		{Ogre, ChanceOgre},
		{SnakeMage, ChanceSnakeMage},
	}
	for i := 1; i < amountRoom; i++ {
		for _, et := range enemyTypes {
			chanceEnemies := GeneratorNum(0, 100)
			if chanceEnemies < et.Chance {
				x, y := level.Rooms[i].RandomPos()
				currentRoom := &level.Rooms[i]
				enemy := NewEnemy(x, y, et.Name, currentRoom)
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
		//генерируется случайное направление движения на 1 клетку
		//1 вправо, 2 вниз, 3 влево, 4 вверх
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
		x1, x2, y1, y2 := en.CurrentRoom.Interior()
		//генерируется случайное направление движения на 1 клетку
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
		// Добавьте логику для SnakeMage
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
	singX := 1 //сохраняем знак
	singY := 1 //сохраняем знак
	// x1, x2, y1, y2 := en.CurrentRoom.Interior()  //не знаю нужно ли проверять на выход из комнаты
	if lenX < 0 {
		lenX *= (-1)
		singX = -1
	} else if lenY < 0 {
		lenY *= (-1)
		singY = -1
	}

	if lenX < lenY { //двигаемся в вертик направлении
		if singY == -1 { //двигаемся вниз
			switch en.TypeEnemy {
			case Zombie, Vampire:
				newY = en.PosYEnemy + 1
			case Ghost:
				newX, newY = TargetX, TargetY-1
			case Ogre:
				newY = en.PosYEnemy + 2
			case SnakeMage:
				if singX == -1 { //если игрок справа от врага
					newX = en.PosXEnemy + 1
					newY = en.PosYEnemy + 1
				} else {
					newX = en.PosXEnemy + 1
					newY = en.PosYEnemy - 1
				}
			}
		} else { //двигаемся вверх
			switch en.TypeEnemy {
			case Zombie, Vampire:
				newY = en.PosYEnemy - 1
			case Ghost:
				newX, newY = TargetX, TargetY+1
			case Ogre:
				newY = en.PosYEnemy - 2
			case SnakeMage:
				if singX == -1 { //если игрок справа от врага
					newX = en.PosXEnemy - 1
					newY = en.PosYEnemy + 1
				} else {
					newX = en.PosXEnemy - 1
					newY = en.PosYEnemy - 1
				}
			}
		}

	} else { //двигаемся в горизонтальном направлении
		if singX == -1 { //двигаемся вправо
			switch en.TypeEnemy {
			case Zombie, Vampire:
				newX = en.PosXEnemy + 1
			case Ghost:
				newX, newY = TargetX-1, TargetY
			case Ogre:
				newX = en.PosXEnemy + 2
			case SnakeMage:
				if singY == -1 { //если игрок снизу от врага
					newX = en.PosXEnemy + 1
					newY = en.PosYEnemy + 1
				} else {
					newX = en.PosXEnemy - 1
					newY = en.PosYEnemy + 1
				}
			}
		} else { //двигаемся влево
			switch en.TypeEnemy {
			case Zombie, Vampire:
				newX = en.PosXEnemy - 1
			case Ghost:
				newX, newY = TargetX+1, TargetY
			case Ogre:
				newX = en.PosXEnemy - 2
			case SnakeMage:
				if singY == -1 { //если игрок снизу от врага
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

// // функция объекта — это способность к перемещению.
// func (player *Player) Move(dx int, dy int) {
// 	player.PosX += dx
// 	player.PosY += dy
// }
