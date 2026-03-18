package main

import (
	"fmt"
	"log"
	"sort"

	"github.com/rthornton128/goncurses"
)

const (
	ScreenWidth  int = 80 //61
	ScreenHeight int = 50 //40
)

type Renderer struct {
	Stdsrc        *goncurses.Window
	GameWindow    *goncurses.Window
	MessageWindow *goncurses.Window
	StatusWindow  *goncurses.Window
}

func (r *Renderer) Init() error {
	stdscr, err := goncurses.Init()
	if err != nil {
		return fmt.Errorf("failed to initialize goncurses: %v", err)
	}

	goncurses.CBreak(true)
	goncurses.Echo(false)
	goncurses.Cursor(0)
	goncurses.StartColor()

	goncurses.InitPair(1, int16(goncurses.C_GREEN), int16(goncurses.C_BLACK))
	goncurses.InitPair(2, int16(goncurses.C_RED), int16(goncurses.C_BLACK))
	goncurses.InitPair(3, int16(goncurses.C_WHITE), int16(goncurses.C_BLACK))
	goncurses.InitPair(4, int16(goncurses.C_YELLOW), int16(goncurses.C_BLACK))

	r.Stdsrc = stdscr

	r.GameWindow = stdscr.Sub(ScreenHeight+2, ScreenWidth+2, 5, 0) // координаты левого верхнего угла 0 0
	r.MessageWindow = stdscr.Sub(ScreenHeight+7, 35, 0, ScreenWidth+2)
	r.StatusWindow = stdscr.Sub(5, ScreenWidth+2, 0, 0)
	return nil
}

// Очистка рендерера
func (r *Renderer) Cleanup() {
	goncurses.End()
}

// Основная функция отрисовки
func (r *Renderer) Render(g *Game) {
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
	} else if g.StateGame == YouLose {
		message := "You Lose!"
		startX := ScreenWidth/2 - len(message)
		startY := ScreenHeight / 2
		r.GameWindow.MovePrint(startY, startX, message)
		r.GameWindow.Box(goncurses.ACS_VLINE, goncurses.ACS_HLINE)
		r.GameWindow.Refresh()
		r.MessageWindow.Refresh()
	} else if g.StateGame == MainMenu {
		title := "ROGUE-LIKE GAME"
		startX := ScreenWidth/2 - len(title)/2
		r.GameWindow.MovePrint(ScreenHeight/2-3, startX, title)
		r.GameWindow.MovePrint(ScreenHeight/2-1, ScreenWidth/2-10, "1. Continue last save")
		r.GameWindow.MovePrint(ScreenHeight/2, ScreenWidth/2-10, "2. Start new game")
		r.GameWindow.MovePrint(ScreenHeight/2+1, ScreenWidth/2-10, "3. Leaderboard")
		r.GameWindow.MovePrint(ScreenHeight/2+2, ScreenWidth/2-10, "Q. Quit")
		if g.SaveManager != nil {
			saves, err := g.SaveManager.ListSaves()
			if err == nil && len(saves) > 0 {
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
				if x == g.Player.PosX && y == g.Player.PosY {
					r.GameWindow.MovePrint(y+1, x+1, string(tile.Symbol))
				} else if g.CurrentLevel.Visible[x][y] {
					r.GameWindow.AttrOn(tile.ColorAttr | goncurses.A_BOLD)
					r.GameWindow.MovePrint(y+1, x+1, string(tile.Symbol))
					r.GameWindow.AttrOff(tile.ColorAttr | goncurses.A_BOLD)
				} else if g.CurrentLevel.Explored[x][y] {
					if tile.Symbol == '|' || tile.Symbol == '-' || tile.Symbol == '#' {
						r.GameWindow.MovePrint(y+1, x+1, string(tile.Symbol))
					} else {
						r.GameWindow.MovePrint(y+1, x+1, " ")
					}
				} else {
					r.GameWindow.MovePrint(y+1, x+1, " ")
				}
			}
		}
		r.StatusWindow.MovePrintf(1, 2, "HP: %d/%d", g.Player.HP, g.Player.MaxHP)
		r.StatusWindow.MovePrintf(1, 20, "Treasure: %d", g.Player.Treasure)
		r.StatusWindow.MovePrintf(1, 40, "Dexterity: %d", g.Player.Dexterity)
		r.StatusWindow.MovePrintf(2, 2, "Strength: %d", g.Player.Strength)
		r.StatusWindow.MovePrintf(3, 2, "Level: %d", g.CurrentLevelIndex+1)
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

	r.GameWindow.Refresh()
	r.StatusWindow.Refresh()
	r.MessageWindow.Refresh()
	r.Stdsrc.Refresh()
}

// EnterPlayerName запрашивает у игрока ввод имени
func (r *Renderer) EnterPlayerName() string {
	r.GameWindow.Clear()
	r.GameWindow.MovePrint(ScreenHeight/2-1, ScreenWidth/2-10, "Enter your name:")

	goncurses.Cursor(1)
	name := ""

	r.GameWindow.MovePrint(ScreenHeight/2+1, ScreenWidth/2-10, "________________")
	r.GameWindow.MovePrint(ScreenHeight/2+1, ScreenWidth/2-10, "")
	r.GameWindow.Refresh()

	for {
		ch := r.GameWindow.GetChar()

		if ch == '\n' || ch == '\r' || ch == goncurses.KEY_ENTER {
			break
		} else if ch == 127 || ch == 8 {
			if len(name) > 0 {
				name = name[:len(name)-1]
			}
		} else if ch >= 32 && ch <= 126 {
			name += string(ch)
		}

		r.GameWindow.MovePrint(ScreenHeight/2+1, ScreenWidth/2-10, "                  ")
		r.GameWindow.MovePrint(ScreenHeight/2+1, ScreenWidth/2-10, name)
		r.GameWindow.Refresh()
	}

	goncurses.Cursor(0)
	return name
}

// Отрисовываем таблицу лидеров полностью со статистикой
func (r *Renderer) ShowLeaderboard(g *Game) {
	oldGameWindow := r.GameWindow
	oldMessageWindow := r.MessageWindow
	oldStatusWindow := r.StatusWindow

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

	leaderboardWindow.MovePrint(3, 3, "Rank")
	leaderboardWindow.MovePrint(3, 10, "Name")
	leaderboardWindow.MovePrint(3, 22, "Level")
	leaderboardWindow.MovePrint(3, 30, "Treasure")
	leaderboardWindow.MovePrint(3, 42, "Enemies")
	leaderboardWindow.MovePrint(3, 50, "Food")
	leaderboardWindow.MovePrint(3, 56, "Elixirs")
	leaderboardWindow.MovePrint(3, 64, "Scrolls")
	leaderboardWindow.MovePrint(3, 72, "Hits")

	leaderboard := g.SaveManager.GetLeaderboard()

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

	leaderboardWindow.GetChar()

	r.GameWindow = oldGameWindow
	r.MessageWindow = oldMessageWindow
	r.StatusWindow = oldStatusWindow

	r.Stdsrc.Clear()
	r.Stdsrc.Refresh()
}
