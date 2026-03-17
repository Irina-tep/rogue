package main

import (
	"log"
	"math"

	"github.com/rthornton128/goncurses"
)

// уровень с картой, комнатами, объектами
type Level struct {
	Tiles    [][]Tile
	Rooms    []Room
	Tunnels  []Tunnel // все тоннели
	Enemies  []Enemy  // все врагами
	Objects  []*Object
	Number   int      // номер уровня
	Explored [][]bool // Матрица исследованных областей
}

// конструктор Level
func NewLevel() Level {
	l := Level{}
	l.Explored = make([][]bool, ScreenWidth)
	for i := range l.Explored {
		l.Explored[i] = make([]bool, ScreenHeight)
	}
	l.CreateRooms()
	l.CreateTunnels()
	l.CreateObjects()
	l.CreateEnemies()
	l.createTiles()
	return l
}

// клетки игрового поля
type Tile struct {
	PosX            int
	PosY            int
	Blocked         bool // можно ли переместиться на клетку или нет, на стену переместиться нельзя, поэтому в этом случае будет true, внутри комнаты перемещаться можно, будет falsw
	Symbol          byte
	BlockedForEnemy bool // враги могут ходить только по своей комнате
	ColorAttr       goncurses.Char
}

// Это определяет набор констант для типов игровых клеток, которые у нас есть, и упрощает дальнейшее расширение, когда мы захотим добавить двери или лестницы.
const (
	TileFloor        string = "floor"
	TileWallVertical string = "wallVertical"
	TileWallHorizont string = "wallHorizont"
	TileOutside      string = "outside"
	TilePlayer       string = "player"
	TileTunnel       string = "tunnel"
	TilePortal       string = "Portal"
	TileZombie       string = "zombie"    // Зомби
	TileVampire      string = "vampire"   // Вампир
	TileGhost        string = "ghost"     // Призрак
	TileOgre         string = "ogre"      // Огр
	TileSnakeMage    string = "snakeMage" // Змееволг
	TileFood         string = "food"      // Пища в определенной степени восстанавливает здоровье .
	// TileTreasure string = "treasure"  //Сокровища — имеют ценность, накапливаются со временем и влияют на итоговый счет. Сокровища можно получить только победив врагов
	TileElixir string = "elixir" // Эликсиры — временно увеличивают один из параметров персонажа: ловкость, силу или максимальное здоровье.
	TileScroll string = "scroll" // Свитки — навсегда увеличивают один из параметров: ловкость, силу или максимальное здоровье.
	TileWeapon string = "weapon" // Оружие обладает показателем силы .
)

// создаем свою ошибку, если найденный символ не найден, возможно потом не будем использовать
type TileNotFoundError struct{}

func (error *TileNotFoundError) Error() string {
	return "Tile not found"
}

func NewTile(x int, y int, tileType string) (Tile, error) {
	blocked := true
	blockedForEnemy := true
	var symbol byte
	var color goncurses.Char
	switch tileType {
	case TileFloor:
		symbol = '.'
		blocked = false
		blockedForEnemy = false
	case TileWallVertical:
		symbol = '|'
	case TileWallHorizont:
		symbol = '-'
	case TileOutside:
		symbol = ' '
	case TilePlayer:
		symbol = '@'
	case TileTunnel:
		symbol = '#'
		blocked = false
	case TilePortal:
		symbol = '%'
		blocked = false
	case TileZombie:
		symbol = 'Z'
		color = goncurses.ColorPair(1)
	case TileVampire:
		symbol = 'V'
		color = goncurses.ColorPair(2)
	case TileGhost:
		symbol = 'G'
		color = goncurses.ColorPair(3)
	case TileOgre:
		symbol = 'O'
		color = goncurses.ColorPair(4)
	case TileSnakeMage:
		symbol = 'S'
		color = goncurses.ColorPair(3)
	case TileElixir:
		symbol = '!'
		blocked = false
	case TileScroll:
		symbol = '?'
		blocked = false
	case TileWeapon:
		symbol = '/'
		blocked = false
	case TileFood:
		symbol = '*'
		blocked = false
	default:
		err := &TileNotFoundError{}
		return Tile{}, err
	}
	tile := Tile{
		PosX:            x,
		PosY:            y,
		Blocked:         blocked,
		Symbol:          symbol,
		BlockedForEnemy: blockedForEnemy,
		ColorAttr:       color,
	}
	return tile, nil
}

// создаем игровую клетку
func (level *Level) createTiles() {
	// Создаем слайс строк
	tiles := make([][]Tile, ScreenWidth)
	// Для каждой строки создаем слайс столбцов
	for i := range tiles {
		tiles[i] = make([]Tile, ScreenHeight)
	}
	// нужно заполнить весь уровень пустотой
	for x := 0; x < ScreenWidth; x++ {
		for y := 0; y < ScreenHeight; y++ {
			outside, err := NewTile(x, y, TileOutside)
			if err != nil {
				log.Fatal(err)
			}
			tiles[x][y] = outside
		}
	}

	// мы проходим по всем комнатам уровня и формируем их интерьер: стены, пол
	for _, room := range level.Rooms {
		x1, x2, y1, y2 := room.Interior()
		for x := x1; x <= x2; x++ {
			for y := y1; y <= y2; y++ {
				floor, err := NewTile(x, y, TileFloor)
				if err != nil {
					log.Fatal(err)
				}
				tiles[x][y] = floor
			}
		}

		for y := room.Y1; y <= room.Y2; y++ {
			wall1, err1 := NewTile(room.X1, y, TileWallVertical)
			if err1 != nil {
				log.Fatal(err1)
			}
			tiles[room.X1][y] = wall1
			wall2, err2 := NewTile(room.X2, y, TileWallVertical)
			if err2 != nil {
				log.Fatal(err2)
			}
			tiles[room.X2][y] = wall2
		}
		for x := room.X1; x <= room.X2; x++ {
			wall1, err1 := NewTile(x, room.Y1, TileWallHorizont)
			if err1 != nil {
				log.Fatal(err1)
			}
			tiles[x][room.Y1] = wall1
			wall2, err2 := NewTile(x, room.Y2, TileWallHorizont)
			if err2 != nil {
				log.Fatal(err2)
			}
			tiles[x][room.Y2] = wall2
		}

	}
	// строим клетки с коридорами
	for _, path := range level.Tunnels {
		for _, coordPath := range path.Path {
			tunnels, err := NewTile(coordPath[0], coordPath[1], TileTunnel)
			if err != nil {
				log.Fatal(err)
			}
			tiles[coordPath[0]][coordPath[1]] = tunnels
		}
	}
	// строим клетки с предметами
	for _, object := range level.Objects {
		tileType := GetObjectSymbol(object.TypeObject)
		tile, err := NewTile(object.PosX, object.PosY, tileType)
		if err != nil {
			log.Fatal(err)
		}
		tiles[object.PosX][object.PosY] = tile
	}

	// строим игрока
	startX, startY := level.Rooms[0].RandomPos()
	player, err := NewTile(startX, startY, TilePlayer)
	if err != nil {
		log.Fatal(err)
	}
	tiles[startX][startY] = player

	// строим клетки с врагами
	for _, enemy := range level.Enemies {
		tile, err := NewTile(enemy.PosXEnemy, enemy.PosYEnemy, enemy.TypeEnemy)
		if err != nil {
			log.Fatal(err)
		}
		tiles[enemy.PosXEnemy][enemy.PosYEnemy] = tile
	}
	level.Tiles = tiles
	// строим клетку с порталом - переход на следующий уровень
	if level.Number < CountLevels {
		coordXPort, coordYPort := level.CreatePortal()
		portal, err := NewTile(coordXPort, coordYPort, TilePortal)
		if err != nil {
			log.Fatal(err)
		}
		level.Tiles[coordXPort][coordYPort] = portal
	}
}

// генерация портала , который дает нам перейти на новый уровень
func (level *Level) CreatePortal() (int, int) {
	// выбираем рандомную комнату, кроме стартовой, она сейчас у на 0
	indexRoom := GeneratorNum(1, 7)
	for {
		coordXPort, coordYPort := level.Rooms[indexRoom].RandomPos()
		// возвращаем только те координаты, которые ранее не заняты
		if level.Tiles[coordXPort][coordYPort].Symbol == '.' {
			return coordXPort, coordYPort
		}
	}
	// возвращает координаты этого портала
}

// функция достает координаты игрока из тайлов
func (level *Level) GetPos() (int, int) {
	for x := 0; x < ScreenWidth; x++ {
		for y := 0; y < ScreenHeight; y++ {
			if level.Tiles[x][y].Symbol == '@' {
				return x, y
			}
		}
	}
	return 0, 0
}

// CreateObjects — размещение предметов в комнатах
func (level *Level) CreateObjects() {
	maxObjects := 5 - level.Number/3
	if maxObjects < 1 {
		maxObjects = 1
	}

	for _, room := range level.Rooms {
		numObjects := GeneratorNum(0, maxObjects)
		for i := 0; i < numObjects; i++ {
			posX, posY := room.RandomPos()
			objectType := GeneratorNum(FOOD, WEAPON)
			object := NewObject(objectType)
			object.PosX, object.PosY = posX, posY
			level.Objects = append(level.Objects, object)
		}
	}
}

//Туман-туманыч

func (l *Level) CalculateVisibility(playerX, playerY int, radius int) {
	// Определяем видимость в пределах радиуса
	for x := 0; x < ScreenWidth; x++ {
		for y := 0; y < ScreenHeight; y++ {
			distance := (x-playerX)*(x-playerX) + (y-playerY)*(y-playerY)
			if distance <= radius*radius && bresenham(playerX, playerY, x, y) {
				l.Explored[x][y] = true
			}
		}
	}

	// Проверяем, находится ли игрок рядом со входом в комнату
	for _, room := range l.Rooms {
		if (playerX == room.X1 || playerX == room.X2) && (math.Abs(float64(playerY-room.Y1)) <= 1 || math.Abs(float64(playerY-room.Y2)) <= 1) ||
			(playerY == room.Y1 || playerY == room.Y2) && (math.Abs(float64(playerX-room.X1)) <= 1 || math.Abs(float64(playerX-room.X2)) <= 1) {
			// Рассеиваем туман в пределах видимости комнаты
			for x := room.X1; x <= room.X2; x++ {
				for y := room.Y1; y <= room.Y2; y++ {
					if bresenham(playerX, playerY, x, y) {
						l.Explored[x][y] = true
					}
				}
			}
		}
	}
}

func bresenham(x0, y0, x1, y1 int) bool {
	dx := math.Abs(float64(x1 - x0))
	dy := -math.Abs(float64(y1 - y0))
	sx := 1
	if x0 >= x1 {
		sx = -1
	}
	sy := 1
	if y0 >= y1 {
		sy = -1
	}
	err := dx + dy

	for {
		if x0 == x1 && y0 == y1 {
			return true
		}
		e2 := 2 * err
		if e2 >= dy {
			if x0 == x1 {
				return false
			}
			err += dy
			x0 += sx
		}
		if e2 <= dx {
			if y0 == y1 {
				return false
			}
			err += dx
			y0 += sy
		}
	}
}
