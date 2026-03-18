package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"time"
)

type SaveData struct {
	Player       PlayerData `json:"player"`       // Данные игрока
	CurrentLevel LevelData  `json:"currentLevel"` // Текущий уровень
	SaveName     string     `json:"saveName"`     // Имя сохранения
	Timestamp    time.Time  `json:"timestamp"`    // Время сохранения
}

// копируем в новую структуру данных все состояние игры
type PlayerData struct {
	Name              string                `json:"name"`
	PosX              int                   `json:"posX"`
	PosY              int                   `json:"posY"`
	HP                int                   `json:"hp"`
	MaxHP             int                   `json:"maxHp"`
	Dexterity         int                   `json:"dexterity"`
	Strength          int                   `json:"strength"`
	CurrenWeapon      string                `json:"weapon"`
	Treasure          int                   `json:"treasure"`
	CurrentLevelIndex int                   `json:"level"`
	CountEnemy        int                   `json:"countEnemy"`
	CountFood         int                   `json:"countFood"`
	CountElixir       int                   `json:"countElixir"`
	CountScrollsRead  int                   `json:"countScrollsRead"`
	CountHits         int                   `json:"countHits"`
	CountTile         int                   `json:"countTile"`
	Backpack          map[int][]*ObjectData `json:"backpack"`         // Добавлено поле рюкзака
	TemporaryEffects  map[string]int        `json:"temporaryEffects"` // Добавлено поле временных эффектов
	IsSleeping        bool                  `json:"isSleeping"`       // Сохраняем состояние сна
}

type LevelData struct {
	Tiles    [][]TileData `json:"tiles"`       // Матрица тайлов
	Rooms    []RoomData   `json:"rooms"`       // Комнаты
	Tunnels  []TunnelData `json:"tunnels"`     // Туннели
	Enemies  []EnemyData  `json:"enemies"`     // Враги
	Number   int          `json:"numberLevel"` // Номер уровня
	Explored [][]bool     `json:"explored"`    // Исследованные области
	Visible  [][]bool     `json:"visible"`     // Видимые в данный момент области
}

type RoomData struct {
	X1 int `json:"x1"`
	Y1 int `json:"y1"`
	X2 int `json:"x2"`
	Y2 int `json:"y2"`
}

type TileData struct {
	X               int  `json:"x"`
	Y               int  `json:"y"`
	Blocked         bool `json:"blocked"`
	Symbol          byte `json:"symbol"`
	BlockedForEnemy bool `json:"blockedForEnemy"`
}

type TunnelData struct {
	Path [][2]int // Путь
}

type EnemyData struct {
	PosXEnemy      int    `json:"posXEnemy"`
	PosYEnemy      int    `json:"posYEnemy"`
	TypeEnemy      string `json:"typeEnemy"`
	HealthEnemy    int    `json:"healthEnemy"`
	DexterityEnemy int    `json:"dexterityEnemy"`
	StrengthEnemy  int    `json:"strengthEnemy"`
	HostilityEnemy int    `json:"hostilityEnemy"`
	CurrentRoom    int    `json:"currentRoom"` // индекс комнаты, а не указатель
	Mode           int    `json:"mode"`
	Treasure       int    `json:"treasure"` // Сохраняем количество сокровищ
}

type ObjectData struct {
	TypeObject    int    `json:"typeObject"`
	SubtypeObject int    `json:"subtypeObject"`
	Health        int    `json:"health"`
	MaxHealth     int    `json:"maxHealth"`
	Dexterity     int    `json:"dexterity"`
	Strength      int    `json:"strength"`
	ValueObject   int    `json:"valueObject"`
	Damage        string `json:"damage"`
}

// Менеджер сохранений
type SaveManager struct {
	SaveDir    string          // Директория для сохранений
	Current    *SaveData       // Текущее сохранение
	Statistics []StatisticData // Статистика всех прохождений
}

func NewSaveManager() *SaveManager {
	// Создаем директорию для сохранений
	saveDir := "./saves"
	if _, err := os.Stat(saveDir); os.IsNotExist(err) {
		os.Mkdir(saveDir, 0755)
	}

	sm := &SaveManager{
		SaveDir: saveDir,
	}

	// Загружаем статистику
	sm.LoadStatistics()
	return sm
}

// SaveGame - сохранение игры
func (sm *SaveManager) SaveGame(g *Game, saveName string) error {
	// Конвертируем игровые данные в структуру сохранения
	saveData := &SaveData{
		SaveName:  saveName,
		Timestamp: time.Now(),
		Player: PlayerData{
			Name:              g.Player.Name,
			PosX:              g.Player.PosX,
			PosY:              g.Player.PosY,
			HP:                g.Player.HP,
			MaxHP:             g.Player.MaxHP,
			Dexterity:         g.Player.Dexterity,
			Strength:          g.Player.Strength,
			CurrenWeapon:      g.Player.CurrenWeapon,
			Treasure:          g.Player.Treasure,
			CurrentLevelIndex: g.Player.CurrentLevelIndex,
			CountEnemy:        g.Player.CountEnemy,
			CountFood:         g.Player.CountFood,
			CountElixir:       g.Player.CountElixir,
			CountScrollsRead:  g.Player.CountScrollsRead,
			CountHits:         g.Player.CountHits,
			CountTile:         g.Player.CountTile,
			Backpack:          make(map[int][]*ObjectData),
		},
		CurrentLevel: LevelData{
			Tiles:    make([][]TileData, ScreenWidth),
			Rooms:    make([]RoomData, len(g.CurrentLevel.Rooms)),
			Tunnels:  make([]TunnelData, len(g.CurrentLevel.Tunnels)),
			Enemies:  make([]EnemyData, len(g.CurrentLevel.Enemies)),
			Number:   g.CurrentLevel.Number,
			Explored: make([][]bool, ScreenWidth),
			Visible:  make([][]bool, ScreenWidth),
		},
	}

	// СОХРАНЯЕМ ТАЙЛЫ - конвертируем в TileData
	for x := 0; x < ScreenWidth; x++ {
		saveData.CurrentLevel.Tiles[x] = make([]TileData, ScreenHeight)
		for y := 0; y < ScreenHeight; y++ {
			tile := g.CurrentLevel.Tiles[x][y]
			saveData.CurrentLevel.Tiles[x][y] = TileData{
				X:               tile.PosX,
				Y:               tile.PosY,
				Blocked:         tile.Blocked,
				Symbol:          tile.Symbol,
				BlockedForEnemy: tile.BlockedForEnemy,
			}
		}
	}
	// Сохраняем исследованные и видимые области
	for x := 0; x < ScreenWidth; x++ {
		saveData.CurrentLevel.Explored[x] = make([]bool, ScreenHeight)
		saveData.CurrentLevel.Visible[x] = make([]bool, ScreenHeight)
		for y := 0; y < ScreenHeight; y++ {
			saveData.CurrentLevel.Explored[x][y] = g.CurrentLevel.Explored[x][y]
			saveData.CurrentLevel.Visible[x][y] = g.CurrentLevel.Visible[x][y]
		}
	}
	// Сохраняем комнаты
	for i, room := range g.CurrentLevel.Rooms {
		saveData.CurrentLevel.Rooms[i] = RoomData{
			X1: room.X1,
			Y1: room.Y1,
			X2: room.X2,
			Y2: room.Y2,
		}
	}
	// Сохраняем туннели
	for i, tunnel := range g.CurrentLevel.Tunnels {
		saveData.CurrentLevel.Tunnels[i] = TunnelData{
			Path: tunnel.Path,
		}
	}
	// СОХРАНЯЕМ ВРАГОВ со всеми полями
	for i, enemy := range g.CurrentLevel.Enemies {
		roomIndex := -1
		for j, room := range g.CurrentLevel.Rooms {
			if enemy.CurrentRoom != nil &&
				enemy.CurrentRoom.X1 == room.X1 &&
				enemy.CurrentRoom.Y1 == room.Y1 &&
				enemy.CurrentRoom.X2 == room.X2 &&
				enemy.CurrentRoom.Y2 == room.Y2 {
				roomIndex = j
				break
			}
		}

		saveData.CurrentLevel.Enemies[i] = EnemyData{
			PosXEnemy:      enemy.PosXEnemy,
			PosYEnemy:      enemy.PosYEnemy,
			TypeEnemy:      enemy.TypeEnemy,
			HealthEnemy:    enemy.HealthEnemy,
			DexterityEnemy: enemy.DexterityEnemy,
			StrengthEnemy:  enemy.StrengthEnemy,
			HostilityEnemy: enemy.HostilityEnemy,
			CurrentRoom:    roomIndex,
			Mode:           enemy.Mode,
			Treasure:       enemy.Treasure,
		}
	}
	// Сохраняем рюкзак
	for objectType, objects := range g.Player.Backpack.Objects {
		for _, object := range objects {
			saveData.Player.Backpack[objectType] = append(saveData.Player.Backpack[objectType], &ObjectData{
				TypeObject:    object.TypeObject,
				SubtypeObject: object.SubtypeObject,
				Health:        object.Health,
				MaxHealth:     object.MaxHealth,
				Dexterity:     object.Dexterity,
				Strength:      object.Strength,
				ValueObject:   object.ValueObject,
				Damage:        object.Damage,
			})
		}
	}

	// Сохраняем временные эффекты
	saveData.Player.TemporaryEffects = g.Player.TemporaryEffects
	saveData.Player.IsSleeping = g.Player.IsSleeping // Сохраняем состояние сна
	// Сохраняем в JSON файл
	filename := filepath.Join(sm.SaveDir, saveName+".json")
	data, err := json.MarshalIndent(saveData, "", "  ")
	if err != nil {
		return fmt.Errorf("serialization error: %v", err)
	}

	err = ioutil.WriteFile(filename, data, 0644)
	if err != nil {
		return fmt.Errorf("file write error: %v", err)
	}

	sm.Current = saveData
	return nil
}

// дополнительные методы

// ListSaves - список всех сохранений
func (sm *SaveManager) ListSaves() ([]SaveInfo, error) {
	files, err := ioutil.ReadDir(sm.SaveDir)
	if err != nil {
		return nil, err
	}

	var saves []SaveInfo
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".json" && file.Name() != "statistics.json" {
			data, err := ioutil.ReadFile(filepath.Join(sm.SaveDir, file.Name()))
			if err != nil {
				continue
			}

			var saveData SaveData
			if json.Unmarshal(data, &saveData) == nil {
				saves = append(saves, SaveInfo{
					Name:        file.Name(),
					SaveName:    saveData.SaveName,
					Timestamp:   saveData.Timestamp,
					PlayerHP:    saveData.Player.HP,
					PlayerMaxHP: saveData.Player.MaxHP,
					Level:       saveData.Player.CurrentLevelIndex,
					EnemyCount:  len(saveData.CurrentLevel.Enemies),
				})
			}
		}
	}

	return saves, nil
}

// SaveInfo - информация для отображения в меню
type SaveInfo struct {
	Name        string    `json:"name"`
	SaveName    string    `json:"saveName"`
	Timestamp   time.Time `json:"timestamp"`
	PlayerHP    int       `json:"playerHp"`
	PlayerMaxHP int       `json:"playerMaxHp"`
	Level       int       `json:"level"`
	EnemyCount  int       `json:"enemyCount"`
}

// DeleteSave - удаление сохранения
func (sm *SaveManager) DeleteSave(filename string) error {
	filepath := filepath.Join(sm.SaveDir, filename)
	return os.Remove(filepath)
}

// GetSavePath - получить полный путь к файлу сохранения
func (sm *SaveManager) GetSavePath(filename string) string {
	return filepath.Join(sm.SaveDir, filename)
}

// LoadGame - загрузка игры из файла
func (sm *SaveManager) LoadGame(filename string) (*SaveData, error) {
	filepath := sm.GetSavePath(filename)
	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("error reading file while saving: %v", err)
	}

	var saveData SaveData
	if err := json.Unmarshal(data, &saveData); err != nil {
		return nil, fmt.Errorf("save deserialization error: %v", err)
	}

	sm.Current = &saveData
	return &saveData, nil
}

// GetLatestSave - получение последнего сохранения
func (sm *SaveManager) GetLatestSave() (*SaveData, error) {
	saves, err := sm.ListSaves()
	if err != nil {
		return nil, err
	}

	if len(saves) == 0 {
		return nil, fmt.Errorf("no saves found")
	}

	// Сортируем по времени (последние сохранения первыми)
	sort.Slice(saves, func(i, j int) bool {
		return saves[i].Timestamp.After(saves[j].Timestamp)
	})

	// Загружаем последнее сохранение
	return sm.LoadGame(saves[0].Name)
}

// ================== Статистика и таблица лидеров ==================

// StatisticData - данные статистики одного прохождения
type StatisticData struct {
	PlayerName     string    `json:"playerName"`
	Timestamp      time.Time `json:"timestamp"`
	ReachedLevel   int       `json:"reachedLevel"`   // достигнутый уровень
	TotalEnemies   int       `json:"totalEnemies"`   // всего побеждено врагов
	TotalTreasure  int       `json:"totalTreasure"`  // всего собрано сокровищ
	TotalFood      int       `json:"totalFood"`      // всего потреблено еды
	TotalElixirs   int       `json:"totalElixirs"`   // всего выпито эликсиров
	TotalScrolls   int       `json:"totalScrolls"`   // всего прочитано свитков
	TotalHits      int       `json:"totalHits"`      // всего нанесено попаданий
	TotalHitsTaken int       `json:"totalHitsTaken"` // всего получено попаданий
	TotalSteps     int       `json:"totalSteps"`     // всего пройдено клеток
	IsCompleted    bool      `json:"isCompleted"`    // завершено ли прохождение
}

// Leaderboard - таблица лидеров
var Leaderboard []StatisticData

// LoadStatistics - загрузка статистики из файла
func (sm *SaveManager) LoadStatistics() error {
	statsFile := filepath.Join(sm.SaveDir, "statistics.json")
	if _, err := os.Stat(statsFile); os.IsNotExist(err) {
		// Файл не существует, создаем пустую статистику
		sm.Statistics = []StatisticData{}
		return nil
	}
	data, err := ioutil.ReadFile(statsFile)
	if err != nil {
		return fmt.Errorf("error reading statistics file: %v", err)
	}
	var stats []StatisticData
	if err := json.Unmarshal(data, &stats); err != nil {
		return fmt.Errorf("statistics deserialization error: %v", err)
	}
	sm.Statistics = stats
	Leaderboard = stats
	return nil
}

// SaveStatistics - сохранение статистики в файл
func (sm *SaveManager) SaveStatistics() error {
	statsFile := filepath.Join(sm.SaveDir, "statistics.json")
	data, err := json.MarshalIndent(sm.Statistics, "", "  ")
	if err != nil {
		return fmt.Errorf("statistics serialization error: %v", err)
	}
	if err := ioutil.WriteFile(statsFile, data, 0644); err != nil {
		return fmt.Errorf("error writing statistics file: %v", err)
	}
	Leaderboard = sm.Statistics
	return nil
}

// AddStatistic - добавление новой статистики прохождения
func (sm *SaveManager) AddStatistic(playerName string, reachedLevel int, totalEnemies int,
	totalTreasure int, totalFood int, totalElixirs int, totalScrolls int, totalHits int, totalHitsTaken int, totalSteps int, isCompleted bool,
) {
	stat := StatisticData{
		PlayerName:     playerName,
		Timestamp:      time.Now(),
		ReachedLevel:   reachedLevel,
		TotalEnemies:   totalEnemies,
		TotalTreasure:  totalTreasure,
		TotalFood:      totalFood,
		TotalElixirs:   totalElixirs,
		TotalScrolls:   totalScrolls,
		TotalHits:      totalHits,
		TotalHitsTaken: totalHitsTaken,
		TotalSteps:     totalSteps,
		IsCompleted:    isCompleted,
	}

	sm.Statistics = append(sm.Statistics, stat)

	// Сортируем по количеству собранных сокровищ (по убыванию)
	sortStatistics(sm.Statistics)

	// Ограничиваем количество записей (например, топ-50)
	if len(sm.Statistics) > 50 {
		sm.Statistics = sm.Statistics[:50]
	}

	sm.SaveStatistics()
}

// sortStatistics - сортировка статистики
func sortStatistics(stats []StatisticData) {
	sort.Slice(stats, func(i, j int) bool {
		// Сначала сортируем по завершённости прохождения
		if stats[i].IsCompleted != stats[j].IsCompleted {
			return stats[i].IsCompleted
		}
		// Затем по количеству собранных сокровищ (по убыванию)
		if stats[i].TotalTreasure != stats[j].TotalTreasure {
			return stats[i].TotalTreasure > stats[j].TotalTreasure
		}
		// Затем по достигнутому уровню (по убыванию)
		if stats[i].ReachedLevel != stats[j].ReachedLevel {
			return stats[i].ReachedLevel > stats[j].ReachedLevel
		}
		// Затем по времени (чем меньше время, тем лучше)
		return stats[i].Timestamp.Before(stats[j].Timestamp)
	})
}

// GetLeaderboard - получение отсортированной таблицы лидеров
func (sm *SaveManager) GetLeaderboard() []StatisticData {
	return Leaderboard
}
