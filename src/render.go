package main

// Здесь отрисовка
import (
	"fmt"
	"sort"
	"time"

	"github.com/rthornton128/goncurses"
)

const (
	ScreenWidth  int = 61 // 64
	ScreenHeight int = 40 // 43
)

// система отрисовки
type Renderer struct {
	Stdsrc        *goncurses.Window // главная структура при использовании библиотеки goncurses
	GameWindow    *goncurses.Window // окно с отрисовкой самой игры
	MessageWindow *goncurses.Window // окно отвечает за сообщения
	StatusWindow  *goncurses.Window // окно со статусом
}

// Инициализация рендерера
func (r *Renderer) Init() error {
	stdscr, err := goncurses.Init()
	if err != nil {
		return fmt.Errorf("failed to initialize goncurses: %v", err)
	}

	// Настраиваем режим ввода
	goncurses.CBreak(true) // Отключаем буферизацию ввода
	goncurses.Echo(false)  // Отключаем эхо ввода
	goncurses.Cursor(0)    // Скрываем курсор
	goncurses.StartColor() // Включаем поддержку цветов

	// Инициализация цветовых пар
	goncurses.InitPair(1, int16(goncurses.C_GREEN), int16(goncurses.C_BLACK))  // Зомби
	goncurses.InitPair(2, int16(goncurses.C_RED), int16(goncurses.C_BLACK))    // Вампир
	goncurses.InitPair(3, int16(goncurses.C_WHITE), int16(goncurses.C_BLACK))  // Призрак
	goncurses.InitPair(4, int16(goncurses.C_YELLOW), int16(goncurses.C_BLACK)) // Огр

	r.Stdsrc = stdscr
	// Создаем окно для игровой области (80x50)
	r.GameWindow = stdscr.Sub(ScreenHeight+2, ScreenWidth+2, 5, 0) // координаты левого верхнего угла 0 0
	// Окно для сообщений (5 строк внизу)
	r.MessageWindow = stdscr.Sub(5, ScreenWidth+2, ScreenHeight+1, 0)
	// Окно для статуса (5 строк сверху)
	r.StatusWindow = stdscr.Sub(5, ScreenWidth+1, 0, 0)
	return nil
}

// Очистка рендерера
func (r *Renderer) Cleanup() {
	goncurses.End()
}

// Основная функция отрисовки
func (r *Renderer) Render(g *Game) {
	// Очищаем окна
	r.GameWindow.Clear()
	r.MessageWindow.Clear()
	r.StatusWindow.Clear()

	if g.StateGame == YouWin {
		message := "You Win!"
		startX := ScreenWidth/2 - len(message)
		startY := ScreenHeight / 2
		r.GameWindow.MovePrint(startY, startX, message)
		r.GameWindow.Box(goncurses.ACS_VLINE, goncurses.ACS_HLINE)
		r.GameWindow.Refresh()
		r.MessageWindow.Refresh()
		// Ждем 10 секунд (блокирует игру)
		time.Sleep(10 * time.Second)
	} else if g.StateGame == YouLose {
		message := "You Lose!"
		startX := ScreenWidth/2 - len(message)
		startY := ScreenHeight / 2
		r.GameWindow.MovePrint(startY, startX, message)
		r.GameWindow.Box(goncurses.ACS_VLINE, goncurses.ACS_HLINE)
		r.GameWindow.Refresh()
		r.MessageWindow.Refresh()
		// Ждем 10 секунд (блокирует игру)
		time.Sleep(10 * time.Second)
	} else if g.StateGame == MainMenu {
		// Отрисовываем главное меню
		title := "ROGUE-LIKE GAME"
		startX := ScreenWidth/2 - len(title)/2
		r.GameWindow.MovePrint(ScreenHeight/2-3, startX, title)

		r.GameWindow.MovePrint(ScreenHeight/2-1, ScreenWidth/2-10, "1. Continue last save")
		r.GameWindow.MovePrint(ScreenHeight/2, ScreenWidth/2-10, "2. Start new game")
		r.GameWindow.MovePrint(ScreenHeight/2+1, ScreenWidth/2-10, "Q. Quit")

		// Показываем информацию о последнем сохранении если оно есть
		if g.SaveManager != nil {
			saves, err := g.SaveManager.ListSaves()
			if err == nil && len(saves) > 0 {
				// Сортируем по времени (последние сохранения первыми)
				sort.Slice(saves, func(i, j int) bool {
					return saves[i].Timestamp.After(saves[j].Timestamp)
				})
				lastSave := saves[0]
				info := fmt.Sprintf("Last save: Level %d, HP: %d/%d",
					lastSave.Level, lastSave.PlayerHP, lastSave.PlayerMaxHP)
				infoX := ScreenWidth/2 - len(info)/2
				r.GameWindow.MovePrint(ScreenHeight/2+3, infoX, info)
			}
		}
	} else {
		// Определяем видимость
		g.CurrentLevel.CalculateVisibility(g.Player.PosX, g.Player.PosY, 8) // Радиус видимости
		// Отрисовываем уровень

		for x := 0; x < len(g.CurrentLevel.Tiles); x++ {
			for y := 0; y < len(g.CurrentLevel.Tiles[x]); y++ {
				tile := g.CurrentLevel.Tiles[x][y]
				// Если это позиция игрока, всегда отображаем игрока
				if x == g.Player.PosX && y == g.Player.PosY {
					//colorPair := getEnemyColor(string(tile.Symbol))
					r.GameWindow.MovePrint(y+1, x+1, string(tile.Symbol))
					//r.GameWindow.ColorOff(colorPair)
				} else if g.CurrentLevel.Explored[x][y] {
					// Проверяем, находится ли игрок в текущей комнате
					inPlayerRoom := false
					for _, room := range g.CurrentLevel.Rooms {
						if x >= room.X1 && x <= room.X2 && y >= room.Y1 && y <= room.Y2 &&
							g.Player.PosX >= room.X1 && g.Player.PosX <= room.X2 &&
							g.Player.PosY >= room.Y1 && g.Player.PosY <= room.Y2 {
							inPlayerRoom = true
							break
						}
					}

					if inPlayerRoom {
						r.GameWindow.AttrOn(tile.ColorAttr | goncurses.A_BOLD)
						r.GameWindow.MovePrint(y+1, x+1, string(tile.Symbol))
						r.GameWindow.AttrOff(tile.ColorAttr | goncurses.A_BOLD)
					} else {
						// Если комната ранее исследована, но игрок не в ней, отображаем только стены и тоннели
						if tile.Symbol == '|' || tile.Symbol == '-' || tile.Symbol == '#' {
							r.GameWindow.MovePrint(y+1, x+1, string(tile.Symbol))
						} else {
							r.GameWindow.MovePrint(y+1, x+1, " ")
						}
					}
				} else {
					// Если область не исследована, отображаем пустоту
					r.GameWindow.MovePrint(y+1, x+1, " ")
				}
			}
		}
		// Отрисовываем статус
		r.StatusWindow.MovePrintf(1, 2, "HP: %d/%d", g.Player.HP, g.Player.MaxHP)
		r.StatusWindow.MovePrintf(1, 20, "Treasure: %d", g.Player.Treasure)
		r.StatusWindow.MovePrintf(1, 40, "Dexterity: %d", g.Player.Dexterity)
		r.StatusWindow.MovePrintf(2, 2, "Strength: %d", g.Player.Strength)
		r.StatusWindow.MovePrintf(3, 2, "Level: %d", g.CurrentLevelIndex+1)
		// r.StatusWindow.MovePrintf(2, 2, "Current Weapon: %s", g.Player.CurrenWeapon)

		// Отрисовываем сообщения (последние 4 сообщения)
		msgY := 1
		start := len(g.Messages) - 4
		if start < 0 {
			start = 0
		}
		for i := start; i < len(g.Messages); i++ {
			r.MessageWindow.MovePrint(msgY, 2, g.Messages[i])
			msgY++
		}
	}
	r.GameWindow.Box(goncurses.ACS_VLINE, goncurses.ACS_HLINE)
	r.MessageWindow.Box(goncurses.ACS_VLINE, goncurses.ACS_HLINE)
	r.StatusWindow.Box(goncurses.ACS_VLINE, goncurses.ACS_HLINE)
	// Обновляем экран
	r.GameWindow.Refresh()
	r.StatusWindow.Refresh()
	r.MessageWindow.Refresh()
	r.Stdsrc.Refresh()
}
