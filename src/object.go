package main

import (
	"fmt"
)

// Типы предметов
const (
	FOOD     = iota
	ELEXIR   = iota
	SCROL    = iota
	WEAPON   = iota
	TREASURE = iota //не участвует в генерации, так он не лежит в комнате, а накапливается при убийсте врага
)

// Подтипы предметов
const (
	RATION = iota
	FRUIT  = iota
)

// Object — структура для предметов
type Object struct {
	TypeObject    int    // Тип предмета (FOOD, POTION, WEAPON, etc.)
	SubtypeObject int    // Подтип (например, для еды: RATION, FRUIT)
	Health        int    // Здоровье (восстановление)
	MaxHealth     int    // Максимальное здоровье (увеличение)
	Dexterity     int    // Ловкость
	Strength      int    // Сила
	ValueObject   int    // Ценность (для сокровищ)
	Quantity      int    // Количество
	IsCursed      bool   // Проклят
	IsIdentified  bool   // Определён
	Damage        string // Урон (например, "1d6")
	PosX, PosY    int    // Позиция на карте
}

// NewObject — создание нового предмета с заданным типом
func NewObject(objectType int) *Object {
	object := &Object{
		TypeObject:   objectType,
		IsCursed:     false,
		IsIdentified: false,
		Damage:       "1d1",
		Quantity:     1,
	}

	// Инициализация в зависимости от типа предмета
	switch objectType {
	case FOOD:
		object.SubtypeObject = RATION
		if GeneratorNum(0, 1) == 0 {
			object.SubtypeObject = FRUIT
		}
	case ELEXIR:
		object.Health = GeneratorNum(1, 10)
	case WEAPON:
		object.Damage = generateWeaponDamage()
		object.Strength = GeneratorNum(1, 3)
	case SCROL:
		object.Dexterity = GeneratorNum(1, 3)
	}
	return object
}

// generateWeaponDamage — генерация случайного урона для оружия
func generateWeaponDamage() string {
	dice := GeneratorNum(1, 3)
	sides := GeneratorNum(2, 6)
	return fmt.Sprintf("%dd%d", dice, sides)
}

// GetObjectSymbol — возвращает символ для типа предмета
func GetObjectSymbol(objectType int) string {
	switch objectType {
	case FOOD:
		return "food"
	case ELEXIR:
		return "elixir"
	case SCROL:
		return "scroll"
	case WEAPON:
		return "weapon"
	default:
		return ""
	}
}
