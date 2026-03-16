package main

// Здесь отрисовка
import (
	"fmt"
	"sort"
	"time"

	"github.com/rthornton128/goncurses"
)

const (
	ScreenWidth  int = 80 // 64
	ScreenHeight int = 50 // 43
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

	r.Stdsrc = stdscr
	// Создаем окно для игровой области (80x50)
	r.GameWindow = stdscr.Sub(ScreenHeight+2, ScreenWidth+2, 5, 0) // координаты левого верхнего угла 0 0
	// Окно для сообщений (5 строк внизу)
	r.MessageWindow = stdscr.Sub(5, ScreenWidth+2, ScreenHeight+7, 0)
	// Окно для статуса (5 строк сверху)
	r.StatusWindow = stdscr.Sub(5, ScreenWidth+2, 0, 0)
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
		// Отрисовываем уровень
		for x := 0; x < len(g.CurrentLevel.Tiles); x++ {
			for y := 0; y < len(g.CurrentLevel.Tiles[x]); y++ {
				tile := g.CurrentLevel.Tiles[x][y]
				r.GameWindow.MovePrint(y+1, x+1, string(tile.Symbol))
			}
		}
		// Отрисовываем статус
		r.StatusWindow.MovePrintf(1, 2, "HP: %d/%d", g.Player.HP, g.Player.MaxHP)
		r.StatusWindow.MovePrintf(1, 20, "Treasure: %d", g.Player.Treasure)
		r.StatusWindow.MovePrintf(1, 40, "Dexterity: %d", g.Player.Dexterity)
		r.StatusWindow.MovePrintf(2, 2, "Strength: %d", g.Player.Strength)
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
