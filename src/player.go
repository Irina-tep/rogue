package main

type Player struct {
	Name              string
	PosX              int
	PosY              int
	HP                int
	MaxHP             int
	Dexterity         int
	Strength          int
	CurrenWeapon      string
	Treasure          int
	Backpack          *Backpack
	TemporaryEffects  map[string]int
	IsSleeping        bool
	CurrentLevelIndex int
	CountEnemy        int
	CountFood         int
	CountElixir       int
	CountScrollsRead  int
	CountHits         int
	TotalHitsTaken    int
	CountTile         int
}

func NewPlayer(x int, y int, currentLevelIndex int) Player {
	player := Player{
		Name:              "Player",
		PosX:              x,
		PosY:              y,
		HP:                20,
		MaxHP:             20,
		Dexterity:         20,
		Strength:          10,
		CurrenWeapon:      "3d1",
		Treasure:          0,
		Backpack:          NewBackpack(),
		TemporaryEffects:  make(map[string]int),
		IsSleeping:        false,
		CurrentLevelIndex: currentLevelIndex + 1,
	}
	return player
}

type Backpack struct {
	Objects map[int][]*Object // int — тип предмета (FOOD, ELEXIR, SCROL, WEAPON), []*Object — список предметов этого типа
}

func NewBackpack() *Backpack {
	backpack := &Backpack{
		Objects: make(map[int][]*Object),
	}

	backpack.Objects[FOOD] = []*Object{}
	backpack.Objects[ELEXIR] = []*Object{}
	backpack.Objects[SCROL] = []*Object{}
	backpack.Objects[WEAPON] = []*Object{}
	return backpack
}

// Метод для добавления предмета в рюкзак
func (bp *Backpack) AddObject(object *Object) bool {
	if len(bp.Objects[object.TypeObject]) >= 9 {
		return false
	}
	bp.Objects[object.TypeObject] = append(bp.Objects[object.TypeObject], object)
	return true
}

// Метод для использования предмета (удаление из рюкзака)
func (bp *Backpack) UseObject(objectType int, index int) (*Object, bool) {
	if objects, exists := bp.Objects[objectType]; exists && index < len(objects) {
		object := objects[index]
		bp.Objects[objectType] = append(objects[:index], objects[index+1:]...)
		return object, true
	}
	return nil, false
}

// Метод для получения списка предметов определенного типа
func (bp *Backpack) GetObjects(objectType int) []*Object {
	return bp.Objects[objectType]
}
