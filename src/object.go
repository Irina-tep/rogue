package main

import (
	"fmt"
)

// Типы предметов
const (
	FOOD = iota
	ELEXIR
	SCROL
	WEAPON

	// Подтипы еды
	RATION
	FRUIT

	// Подтипы эликсиров
	DEXTERITY_ELIXIR
	STRENGTH_ELIXIR

	// Подтипы свитков
	DEXTERITY_SCROL
	STRENGTH_SCROL
	HEALTH_SCROL

	// Подтипы оружия
	DAGGER
	SWORD
)

type Object struct {
	TypeObject    int
	SubtypeObject int
	Health        int
	MaxHealth     int
	Dexterity     int
	Strength      int
	ValueObject   int
	Quantity      int
	IsCursed      bool
	IsIdentified  bool
	Damage        string
	PosX, PosY    int
}

// NewObject — создание нового предмета с заданным типом
func NewObject(objectType int) *Object {
	object := &Object{TypeObject: objectType}

	switch objectType {
	case FOOD:
		object.SubtypeObject = GeneratorNum(RATION, FRUIT)
		object.Health = 5
	case ELEXIR:
		subtype := GeneratorNum(DEXTERITY_ELIXIR, STRENGTH_ELIXIR)
		object.SubtypeObject = subtype
		if subtype == DEXTERITY_ELIXIR {
			object.Dexterity = 2
		} else if subtype == STRENGTH_ELIXIR {
			object.Strength = 2
		}
	case SCROL:
		subtype := GeneratorNum(DEXTERITY_SCROL, HEALTH_SCROL)
		object.SubtypeObject = subtype
		if subtype == DEXTERITY_SCROL {
			object.Dexterity = 2
		} else if subtype == STRENGTH_SCROL {
			object.Strength = 2
		} else if subtype == HEALTH_SCROL {
			object.MaxHealth = 5
		}
	case WEAPON:
		subtype := GeneratorNum(DAGGER, SWORD)
		object.SubtypeObject = subtype
		if subtype == DAGGER {
			object.Damage = "1d4"
		} else if subtype == SWORD {
			object.Damage = "1d6"
		}
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
