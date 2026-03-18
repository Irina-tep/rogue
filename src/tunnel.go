package main

import (
	"math/rand/v2"
)

// количество вершин (0..8)
const (
	amountRoom = 9
)

type Tunnel struct {
	Path [][2]int
}

// находим связи со всеми комнатами, возаращаем слайс 2-х связных комнат
func CreateGraph() [][2]int {
	var Rebra [][2]int
	var VerticesInTree []int
	var AvailableVertices []int
	start := rand.IntN(amountRoom)
	VerticesInTree = append(VerticesInTree, start)
	AvailableVertices = append(AvailableVertices, start)
	for len(VerticesInTree) < amountRoom {
		indexCurrentVertice := rand.IntN(len(AvailableVertices))
		currentVertice := AvailableVertices[indexCurrentVertice]
		allNeighborVertice := orthogonalNeighbor(currentVertice)
		// Оставляем только тех соседей, которых еще нет в дереве
		var UnconnectedVertices []int
		for _, neighbor := range allNeighborVertice {

			found := false
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
			newNeighbor := UnconnectedVertices[rand.IntN(len(UnconnectedVertices))]
			Rebra = append(Rebra, [2]int{currentVertice, newNeighbor})
			VerticesInTree = append(VerticesInTree, newNeighbor)
			AvailableVertices = append(AvailableVertices, newNeighbor)
		} else {
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
	horizont := false
	if room1-room2 == 1 { //если первая комната расположена правее второй
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
	tunnel.Path = make([][2]int, 0)
	//делаем так чтобы направление пути всегда было слева направо и сверху вниз
	if (!horizont && coordYDoor1 > coordYDoor2) || (horizont && coordXDoor1 > coordXDoor2) {
		coordXDoor1, coordXDoor2 = coordXDoor2, coordXDoor1
		coordYDoor1, coordYDoor2 = coordYDoor2, coordYDoor1
	}
	if horizont {

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
