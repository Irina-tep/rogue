package main

// создадим структуру, которую будем использовать для создания наших комнат
type Room struct {
	X1 int
	Y1 int
	X2 int
	Y2 int
}

// Конструктор принимает координаты x и y верхнего угла и вычисляет нижний правый угол на основе параметров ширины и высоты.
func NewRoom(x int, y int, width int, height int) Room {
	return Room{
		X1: x,
		Y1: y,
		X2: x + width,
		Y2: y + height,
	}
}

// Чтобы гарантировать, что наши комнаты будут окружены хотя бы одним слоем настенной плитки, нам следует создать функцию, которая возвращает внутреннее пространство комнаты
// Возвращает координаты плитки внутренней части левый верхний, правый нижний
func (r *Room) Interior() (int, int, int, int) {
	return r.X1 + 1, r.X2 - 1, r.Y1 + 1, r.Y2 - 1
}

// создаются случайные координаты, в которых будет появляться персонаж в начале уровня внутри комнаты
func (r *Room) RandomPos() (int, int) {
	// Инициализация при запуске программы
	// rand.Seed(time.Now().UnixNano())
	x1, x2, y1, y2 := r.Interior()
	randX := GeneratorNum(x1, x2)
	randY := GeneratorNum(y1, y2)
	return randX, randY
}

const (
	GridSizeRoom = 3 // делим поле 3х3
	MinSpacing   = 3 // минимальное расстояние между комнатами
	MinsizeRoom  = 4 // минимальная длина и ширина 4 клетки
)

func (level *Level) CreateRooms() {
	// рандомная генерация комнат
	// Разбиваем поле на сетку с учетом зазоров
	// 3x3 сетки с промежутками 3 клетки
	// Рассчитываем доступное пространство для комнат
	availableWidth := ScreenWidth - (GridSizeRoom-1)*MinSpacing
	availableHeight := ScreenHeight - (GridSizeRoom-1)*MinSpacing

	// Разбиваем на 3 колонки и 3 ряда
	widthSection := availableWidth / GridSizeRoom
	heightSection := availableHeight / GridSizeRoom
	// Для каждой секции генерируем комнату
	for indexY := 0; indexY < GridSizeRoom; indexY++ {
		for indexX := 0; indexX < GridSizeRoom; indexX++ {
			// Вычисляем границы секции
			sectionXstart := indexX * (widthSection + MinSpacing)
			sectionYstart := indexY * (heightSection + MinSpacing)
			sectionXend := sectionXstart + widthSection
			sectionYend := sectionYstart + heightSection
			// Генерируем комнату внутри секции со случайным высотой и ширриной
			width := GeneratorNum(MinsizeRoom, widthSection)
			height := GeneratorNum(MinsizeRoom, heightSection)
			// Случайная позиция внутри секции
			max_x := sectionXend - width
			max_y := sectionYend - height
			x := GeneratorNum(sectionXstart, max_x)
			y := GeneratorNum(sectionYstart, max_y)
			room := NewRoom(x, y, width, height) // левый верний угол, ширина, высота
			level.Rooms = append(level.Rooms, room)
		}
	}
}

// Метод для проверки, заблокирована ли клетка
func (r *Room) IsBlocked(level *Level, x, y int) bool {
	if x < r.X1 || x > r.X2 || y < r.Y1 || y > r.Y2 {
		return true // Клетка вне комнаты
	}
	return level.Tiles[x][y].Blocked
}
