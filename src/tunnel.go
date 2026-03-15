package main

import (
	"math/rand/v2"
)

// количество вершин (0..8)
const (
	amountRoom = 9
)

type Tunnel struct {
	Path [][2]int //Путь
}

// находим связи со всеми комнатами, возаращаем слайс 2-х связных комнат
func CreateGraph() [][2]int {
	var Rebra [][2]int          // Список пар как срез из [2]int
	var VerticesInTree []int    // множество вершин, уже добавленных в дерево
	var AvailableVertices []int // вершины, из которых можно растить дерево
	// Шаг 1: Инициализация
	start := rand.IntN(amountRoom)                       //случайное_число(0, 8)
	VerticesInTree = append(VerticesInTree, start)       // добавить старт в в_дереве
	AvailableVertices = append(AvailableVertices, start) // добавить старт в доступные_вершины
	// Шаг 2: Основной цикл - строим дерево
	for len(VerticesInTree) < amountRoom {
		// Выбираем случайную вершину из тех, у кого есть неподключенные соседи
		// (на практике просто выбираем случайную из доступных_вершины)
		indexCurrentVertice := rand.IntN(len(AvailableVertices)) //сохраняем индекс текущей вершины
		currentVertice := AvailableVertices[indexCurrentVertice] //текущая = выбрать_случайный_элемент(доступные_вершины)
		// Получаем всех ортогональных соседей текущей вершины
		allNeighborVertice := orthogonalNeighbor(currentVertice)
		// Оставляем только тех соседей, которых еще нет в дереве
		var UnconnectedVertices []int // неподключенные_соседи
		for _, neighbor := range allNeighborVertice {
			// Проверяем, есть ли сосед в дереве c помощью slices.Contains
			// found := slices.Contains(VerticesInTree, neighbor)
			// if !found {
			// 	UnconnectedVertices = append(UnconnectedVertices, neighbor)
			// }
			found := false
			// Проверяем, есть ли сосед в дереве через цикл
			for _, v := range VerticesInTree {
				if neighbor == v {
					found = true
					break
				}
			}
			if !found {
				UnconnectedVertices = append(UnconnectedVertices, neighbor)
			}
		}
		// Если у текущей вершины есть неподключенные соседи
		if len(UnconnectedVertices) > 0 {
			// Выбираем случайного соседа для подключения
			newNeighbor := UnconnectedVertices[rand.IntN(len(UnconnectedVertices))] //новый_сосед = выбрать_случайный_элемент(неподключенные_соседи)
			// Добавляем ребро
			Rebra = append(Rebra, [2]int{currentVertice, newNeighbor}) // добавить (текущая, новый_сосед) в ребра
			// Добавляем новую вершину в дерево
			VerticesInTree = append(VerticesInTree, newNeighbor)       //добавить новый_сосед в в_дереве
			AvailableVertices = append(AvailableVertices, newNeighbor) //добавить новый_сосед в доступные_вершины
		} else {
			// Если у текущей вершины нет неподключенных соседей, удаляем её из доступных вершин (она "исчерпала себя")
			AvailableVertices = append(AvailableVertices[:indexCurrentVertice], AvailableVertices[indexCurrentVertice+1:]...)
		}
	}
	return Rebra
}

func orthogonalNeighbor(vertice int) []int {
	row := vertice / 3
	col := vertice % 3
	orthoNeighbor := make([]int, 0)
	if row > 0 { // вверний сосед
		orthoNeighbor = append(orthoNeighbor, vertice-3)
	}
	if row < 2 { // нижний сосед
		orthoNeighbor = append(orthoNeighbor, vertice+3)
	}
	if col > 0 { //левый сосед
		orthoNeighbor = append(orthoNeighbor, vertice-1)
	}
	if col < 2 { //правый сосед
		orthoNeighbor = append(orthoNeighbor, vertice+1)
	}
	return orthoNeighbor
}

// находим двери для этого нужны расположение всех комнат в левеле и 2 вершины графа(комнаты) возвращаем координаты дверей
func FoundDoors(level *Level, room1 int, room2 int) (int, int, int, int, bool) {
	var coordXDoor1, coordYDoor1, coordXDoor2, coordYDoor2 int
	horizont := false     //направление соединение комнат (комната1 расположены относительно комнаты2 слева/справа/сверху/внизу)
	if room1-room2 == 1 { //если первая комната расположена правее второй
		//генерируем координату двери x=x1, y от х1+1 до х2-1, так исключаем углы
		coordXDoor1 = level.Rooms[room1].X1
		coordYDoor1 = GeneratorNum(level.Rooms[room1].Y1+1, level.Rooms[room1].Y2-1)
		coordXDoor2 = level.Rooms[room2].X2
		coordYDoor2 = GeneratorNum(level.Rooms[room2].Y1+1, level.Rooms[room2].Y2-1)
		horizont = true
	} else if room1-room2 == -1 { //если первая комната расположена левее второй
		coordXDoor1 = level.Rooms[room1].X2
		coordYDoor1 = GeneratorNum(level.Rooms[room1].Y1+1, level.Rooms[room1].Y2-1)
		coordXDoor2 = level.Rooms[room2].X1
		coordYDoor2 = GeneratorNum(level.Rooms[room2].Y1+1, level.Rooms[room2].Y2-1)
		horizont = true
	} else if room1-room2 == 3 { //если первая комната расположена ниже второй
		coordXDoor1 = GeneratorNum(level.Rooms[room1].X1+1, level.Rooms[room1].X2-1)
		coordYDoor1 = level.Rooms[room1].Y1
		coordXDoor2 = GeneratorNum(level.Rooms[room2].X1+1, level.Rooms[room2].X2-1)
		coordYDoor2 = level.Rooms[room2].Y2
		horizont = false
	} else if room1-room2 == -3 { //если первая комната расположена выше второй
		coordXDoor1 = GeneratorNum(level.Rooms[room1].X1+1, level.Rooms[room1].X2-1)
		coordYDoor1 = level.Rooms[room1].Y2
		coordXDoor2 = GeneratorNum(level.Rooms[room2].X1+1, level.Rooms[room2].X2-1)
		coordYDoor2 = level.Rooms[room2].Y1
		horizont = false
	}
	return coordXDoor1, coordYDoor1, coordXDoor2, coordYDoor2, horizont
}

func (tunnel *Tunnel) CreateTunnel(coordXDoor1 int, coordYDoor1 int, coordXDoor2 int, coordYDoor2 int, horizont bool) {
	// Очищаем путь перед созданием
	tunnel.Path = make([][2]int, 0)
	//делаем так чтобы направление пути всегда было слева направо и сверху вниз
	if (!horizont && coordYDoor1 > coordYDoor2) || (horizont && coordXDoor1 > coordXDoor2) { //если идет справа налево, то меняем местами координаты
		coordXDoor1, coordXDoor2 = coordXDoor2, coordXDoor1
		coordYDoor1, coordYDoor2 = coordYDoor2, coordYDoor1
	}
	if horizont {
		// Проверяем, что есть место для поворота, если нет, то будет просто горизонтальная линия
		if coordXDoor2-coordXDoor1 < 3 {
			for x := coordXDoor1; x <= coordXDoor2; x++ {
				tunnel.Path = append(tunnel.Path, [2]int{x, coordYDoor1})
			}
		} else {
			//генерируем Х поворота
			coordTurnX := GeneratorNum(coordXDoor1+1, coordXDoor2-1)
			//заполняем горизонтальный путь
			for x := coordXDoor1; x <= coordTurnX; x++ {
				tunnel.Path = append(tunnel.Path, [2]int{x, coordYDoor1})
			}
			//заполняем вертикальный путь
			if coordYDoor1 < coordYDoor2 {
				for y := coordYDoor1 + 1; y <= coordYDoor2; y++ {
					tunnel.Path = append(tunnel.Path, [2]int{coordTurnX, y})
				}
			} else {
				for y := coordYDoor1 - 1; y >= coordYDoor2; y-- {
					tunnel.Path = append(tunnel.Path, [2]int{coordTurnX, y})
				}
			}
			// Финальная горизонтальная часть
			for x := coordTurnX + 1; x <= coordXDoor2; x++ {
				tunnel.Path = append(tunnel.Path, [2]int{x, coordYDoor2})
			}
		}

	} else {
		// Проверяем, что есть место для поворотa, если нет, то будет просто вертикальная линия
		if coordYDoor2-coordYDoor1 < 3 {
			for y := coordYDoor1; y <= coordYDoor2; y++ {
				tunnel.Path = append(tunnel.Path, [2]int{coordXDoor1, y})
			}
		} else {
			coordTurnY := GeneratorNum(coordYDoor1+1, coordYDoor2-1)
			//заполняем вертикальный путь
			for y := coordYDoor1; y <= coordTurnY; y++ {
				tunnel.Path = append(tunnel.Path, [2]int{coordXDoor1, y})
			}
			//заполняем горизонтальный путь
			if coordXDoor1 < coordXDoor2 {
				for x := coordXDoor1 + 1; x <= coordXDoor2; x++ {
					tunnel.Path = append(tunnel.Path, [2]int{x, coordTurnY})
				}
			} else {
				for x := coordXDoor1 - 1; x >= coordXDoor2; x-- {
					tunnel.Path = append(tunnel.Path, [2]int{x, coordTurnY})
				}
			}
			// Финальная вертикальная часть
			for y := coordTurnY + 1; y <= coordYDoor2; y++ {
				tunnel.Path = append(tunnel.Path, [2]int{coordXDoor2, y})
			}
		}
	}
}

func (level *Level) CreateTunnels() {
	//создаем связи через граф
	connection := CreateGraph()
	// Для каждой связи создаем туннель
	for _, tunnel := range connection {
		room1 := tunnel[0]
		room2 := tunnel[1]
		// Находим двери для пары комнат
		coordXDoor1, coordYDoor1, coordXDoor2, coordYDoor2, horizont := FoundDoors(level, room1, room2)
		// Создаем туннель между дверями
		var t Tunnel
		t.CreateTunnel(coordXDoor1, coordYDoor1, coordXDoor2, coordYDoor2, horizont)
		level.Tunnels = append(level.Tunnels, t)
	}
}
