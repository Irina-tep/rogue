package main

// Здесь отрисовка
import (
	"fmt"
	"log"
	"sort"

	"github.com/rthornton128/goncurses"
)

const (
	ScreenWidth  int = 61
	ScreenHeight int = 40
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
	r.MessageWindow = stdscr.Sub(ScreenHeight+7, 35, 0, ScreenWidth+2)
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
		// time.Sleep(10 * time.Second)
	} else if g.StateGame == YouLose {
		message := "You Lose!"
		startX := ScreenWidth/2 - len(message)
		startY := ScreenHeight / 2
		r.GameWindow.MovePrint(startY, startX, message)
		r.GameWindow.Box(goncurses.ACS_VLINE, goncurses.ACS_HLINE)
		r.GameWindow.Refresh()
		r.MessageWindow.Refresh()
		// Ждем 10 секунд (блокирует игру)
		// time.Sleep(10 * time.Second)
	} else if g.StateGame == MainMenu {
		// Отрисовываем главное меню
		title := "ROGUE-LIKE GAME"
		startX := ScreenWidth/2 - len(title)/2
		r.GameWindow.MovePrint(ScreenHeight/2-3, startX, title)
		r.GameWindow.MovePrint(ScreenHeight/2-1, ScreenWidth/2-10, "1. Continue last save")
		r.GameWindow.MovePrint(ScreenHeight/2, ScreenWidth/2-10, "2. Start new game")
		r.GameWindow.MovePrint(ScreenHeight/2+1, ScreenWidth/2-10, "3. Leaderboard")
		r.GameWindow.MovePrint(ScreenHeight/2+2, ScreenWidth/2-10, "Q. Quit")
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
				r.GameWindow.MovePrint(ScreenHeight/2+4, infoX, info)
			}
		}
	} else {
		// Отрисовываем уровень
		for x := 0; x < len(g.CurrentLevel.Tiles); x++ {
			for y := 0; y < len(g.CurrentLevel.Tiles[x]); y++ {
				tile := g.CurrentLevel.Tiles[x][y]
				// Если это позиция игрока, всегда отображаем игрока
				if x == g.Player.PosX && y == g.Player.PosY {
					r.GameWindow.MovePrint(y+1, x+1, string(tile.Symbol))
				} else if g.CurrentLevel.Visible[x][y] {
					// Клетка видима в данный момент
					r.GameWindow.AttrOn(tile.ColorAttr | goncurses.A_BOLD)
					r.GameWindow.MovePrint(y+1, x+1, string(tile.Symbol))
					r.GameWindow.AttrOff(tile.ColorAttr | goncurses.A_BOLD)
				} else if g.CurrentLevel.Explored[x][y] {
					// Клетка исследована, но не видима: отображаем только стены и тоннели
					if tile.Symbol == '|' || tile.Symbol == '-' || tile.Symbol == '#' {
						r.GameWindow.MovePrint(y+1, x+1, string(tile.Symbol))
					} else {
						r.GameWindow.MovePrint(y+1, x+1, " ")
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
		start := len(g.Messages) - 10
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

// Отрисовываем таблицу лидеров только со статистикой сокровищ

// func (r *Renderer) ShowLeaderboard(g *Game) {
// 	r.GameWindow.Clear()
// 	r.MessageWindow.Clear()
// 	r.StatusWindow.Clear()

// 	title := "LEADERBOARD"
// 	startX := ScreenWidth/2 - len(title)/2
// 	r.GameWindow.MovePrint(ScreenHeight/2-5, startX, title)

// 	// Получаем таблицу лидеров
// 	leaderboard := g.SaveManager.GetLeaderboard()

// 	// Отрисовываем заголовки
// 	r.GameWindow.MovePrint(ScreenHeight/2-3, 2, "Rank")
// 	r.GameWindow.MovePrint(ScreenHeight/2-3, 10, "Player")
// 	r.GameWindow.MovePrint(ScreenHeight/2-3, 25, "Level")
// 	r.GameWindow.MovePrint(ScreenHeight/2-3, 35, "Treasure")

// 	// Отрисовываем записи
// 	for i, stat := range leaderboard {
// 		if i >= 10 { // Ограничиваемся топ-10
// 			break
// 		}
// 		y := ScreenHeight/2 - 2 + i
// 		r.GameWindow.MovePrint(y, 2, fmt.Sprintf("%d", i+1))
// 		r.GameWindow.MovePrint(y, 10, stat.PlayerName)
// 		r.GameWindow.MovePrint(y, 25, fmt.Sprintf("%d", stat.ReachedLevel))
// 		r.GameWindow.MovePrint(y, 35, fmt.Sprintf("%d", stat.TotalTreasure))
// 	}

// 	r.GameWindow.MovePrint(ScreenHeight-2, 2, "Press any key to return to menu...")

// 	r.GameWindow.Box(goncurses.ACS_VLINE, goncurses.ACS_HLINE)
// 	r.GameWindow.Refresh()
// 	r.MessageWindow.Refresh()
// 	r.StatusWindow.Refresh()
// }

// EnterPlayerName запрашивает у игрока ввод имени
func (r *Renderer) EnterPlayerName() string {
	r.GameWindow.Clear()
	r.GameWindow.MovePrint(ScreenHeight/2-1, ScreenWidth/2-10, "Enter your name:")
	// goncurses.Echo(true)
	goncurses.Cursor(1)
	name := ""
	// Показываем начальное положение курсора
	r.GameWindow.MovePrint(ScreenHeight/2+1, ScreenWidth/2-10, "________________")
	r.GameWindow.MovePrint(ScreenHeight/2+1, ScreenWidth/2-10, "")
	r.GameWindow.Refresh()

	for {
		ch := r.GameWindow.GetChar()
		// Проверяем различные коды клавиши Enter
		if ch == '\n' || ch == '\r' || ch == goncurses.KEY_ENTER {
			break
		} else if ch == 127 || ch == 8 { // Backspace
			if len(name) > 0 {
				name = name[:len(name)-1]
			}
		} else if ch >= 32 && ch <= 126 { // Только печатные ASCII символы
			name += string(ch)
		}
		// Очищаем строку и выводим заново
		r.GameWindow.MovePrint(ScreenHeight/2+1, ScreenWidth/2-10, "                  ")
		r.GameWindow.MovePrint(ScreenHeight/2+1, ScreenWidth/2-10, name)
		r.GameWindow.Refresh()
	}
	// goncurses.Echo(false)
	goncurses.Cursor(0)
	return name
}

// Отрисовываем таблицу лидеров полностью со статистикой
func (r *Renderer) ShowLeaderboard(g *Game) {
	// Сохраняем текущие окна
	oldGameWindow := r.GameWindow
	oldMessageWindow := r.MessageWindow
	oldStatusWindow := r.StatusWindow

	// Создаем новое окно для таблицы лидеров
	leaderboardWindow, err := goncurses.NewWindow(30, 80, 2, 2)
	if err != nil {
		log.Fatalf("Failed to create leaderboard window: %v", err)
	}
	defer leaderboardWindow.Delete()

	leaderboardWindow.Clear()
	leaderboardWindow.Box(goncurses.ACS_VLINE, goncurses.ACS_HLINE)

	title := "LEADERBOARD"
	startX := 40 - len(title)/2
	leaderboardWindow.MovePrint(1, startX, title)

	// Отрисовываем заголовки с равномерным распределением
	leaderboardWindow.MovePrint(3, 3, "Rank")
	leaderboardWindow.MovePrint(3, 10, "Name")
	leaderboardWindow.MovePrint(3, 22, "Level")
	leaderboardWindow.MovePrint(3, 30, "Treasure")
	leaderboardWindow.MovePrint(3, 42, "Enemies")
	leaderboardWindow.MovePrint(3, 50, "Food")
	leaderboardWindow.MovePrint(3, 56, "Elixirs")
	leaderboardWindow.MovePrint(3, 64, "Scrolls")
	leaderboardWindow.MovePrint(3, 72, "Hits")

	// Получаем таблицу лидеров
	leaderboard := g.SaveManager.GetLeaderboard()

	// Отрисовываем записи
	for i, stat := range leaderboard {
		if i >= 20 { // Ограничиваемся топ-20
			break
		}
		y := 5 + i
		leaderboardWindow.MovePrint(y, 3, fmt.Sprintf("%2d", i+1))
		leaderboardWindow.MovePrint(y, 10, fmt.Sprintf("%-10s", stat.PlayerName))
		leaderboardWindow.MovePrint(y, 22, fmt.Sprintf("%3d", stat.ReachedLevel))
		leaderboardWindow.MovePrint(y, 30, fmt.Sprintf("%6d", stat.TotalTreasure))
		leaderboardWindow.MovePrint(y, 42, fmt.Sprintf("%6d", stat.TotalEnemies))
		leaderboardWindow.MovePrint(y, 50, fmt.Sprintf("%3d", stat.TotalFood))
		leaderboardWindow.MovePrint(y, 56, fmt.Sprintf("%3d", stat.TotalElixirs))
		leaderboardWindow.MovePrint(y, 64, fmt.Sprintf("%3d", stat.TotalScrolls))
		leaderboardWindow.MovePrint(y, 72, fmt.Sprintf("%3d/%d", stat.TotalHits, stat.TotalHitsTaken))
	}

	leaderboardWindow.MovePrint(27, 3, "Press any key to return to menu...")
	leaderboardWindow.Refresh()

	// Ждем нажатия клавиши
	leaderboardWindow.GetChar()

	// Удаляем окно таблицы лидеров
	leaderboardWindow.Delete()

	// Восстанавливаем старые окна
	r.GameWindow = oldGameWindow
	r.MessageWindow = oldMessageWindow
	r.StatusWindow = oldStatusWindow

	// Очищаем и обновляем основной экран
	r.Stdsrc.Clear()
	r.Stdsrc.Refresh()

	// Перерисовываем основное окно
	r.Render(g)
}
