package main

// игрок с инвентарем и статистиками
type Player struct {
	PosX         int
	PosY         int
	HP           int
	MaxHP        int
	Dexterity    int //ловкость
	Strength     int
	CurrenWeapon string //текущее оружие
	Treasure     int    //сокровища  //нужно сохранять в статистику
	//для статистики
	CurrentLevelIndex int //достигнутый самый глубокий уровень
	CountEnemy        int // количество побежденных врагов
	CountFood         int // количество потребленной пищи
	CountElixir       int // количество выпитых эликсиров
	CountScrollsRead  int //количество прочитанных свитков
	CountHits         int // общее количество нанесенных и полученных попаданий
	CountTile         int // количество пройденных клеток
}

func NewPlayer(x int, y int, currentLevelIndex int) Player {
	player := Player{
		PosX:      x,
		PosY:      y,
		HP:        20,
		MaxHP:     20,
		Dexterity: 20,
		Strength:  10,
		// CurrenWeapon:
		Treasure:          0,
		CurrentLevelIndex: currentLevelIndex + 1,
	}
	return player
}

// // функция объекта — это способность к перемещению.
// func (player *Player) Move(dx int, dy int) {
// 	player.PosX += dx
// 	player.PosY += dy
// }
