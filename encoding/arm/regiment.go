package arm

import "math"

type Regiment struct {
	status []byte // TODO: 11 00 = unit with you. 10 00 = don't have unit...
	id     uint16

	// Name is the name of the regiment, e.g. "Grudgebringer Cavalry",
	// "Zombies #1", "Imperial Steam Tank".
	Name string

	nameID uint16

	// alignment is the regiment's alignment to good or evil.
	// 0x00 (decimal 0) is good. 0x40 (decimal 64) is neutral. 0x80 (decimal) is
	// evil.
	alignment uint8
	// typ is a bitfield for the regiment's type and race.
	// The lower 3 bits determine the race. The higher 5 bits determine the
	// regiment's type.
	typ uint8
	// BannerIndex is the index into the list of sprite file names found in
	// ENGREL.EXE for the regiment's banner. Use the engrel.ReadSpriteName
	// function with this index to find the name of the sprite file to use for
	// the regiment's banner (both the small and large banners).
	BannerIndex uint16
	// SpriteIndex is the index into the list of sprite file names found in
	// ENGREL.EXE for the regiment's troop sprite. Use the engrel.ReadSpriteName
	// function with this index to find the name of the sprite file to use for
	// the regiment's troop sprite.
	SpriteIndex uint16
	// MaxTroops is the maximum number of troops allowed in this regiment.
	MaxTroops uint8
	// AliveTroops is the number of troops currently alive in this regiment.
	AliveTroops uint8

	ranks              uint8
	regimentAttributes []byte
	troopAttributes    *troopAttributes
	mount              uint8
	armour             uint8
	weapon             uint8
	pointValue         uint8
	missileWeapon      uint8

	// Leader is the regiment's leader.
	Leader *leader
	// Experience is a number that represents the regiment's total experience.
	// It is a number between 0 and 6000. If experience is <1000 then the
	// regiment has a threat level of 1. If experience >=1000 and <3000 then the
	// regiment has a threat level of 2. If experience >= 3000 and <6000 then
	// the regiment has a threat level of 3. If experience >= 6000 then the
	// regiment has a threat level of 4.
	Experience uint16
	// MinArmour is the regiment's minimum or base level of armour.
	// This is displayed as the gold shields in the troop roster.
	MinArmour uint8
	// MaxArmour is the regiment's maximum level of armour.
	MaxArmour uint8
	// MagicBook is the magic book that is equipped to the regiment. A magic
	// book is one of the magic items.
	// This is an index into the list of magic items. Index 0 means no magic
	// book is equipped. Index 22 means the Bright Book is equipped. Index 23
	// means the Ice Book is equipped. Index 65535 means the regiment does not
	// have a magic book slot—only magic users can equip magic books.
	MagicBook uint16
	// MagicItems is a list of magic items that are equipped to the regiment.
	// Each magic item is an index into the list of magic items. Index 0 means
	// no magic item is equipped in that slot. Index 1 means the Grudgebringer
	// Sword is equipped in that slot. Index 65535 means the regiment can not
	// use that slot.
	// TODO: Does 65535 mean that? All units seem to have that value if there is
	// no item equipped.
	MagicItems [3]uint16

	Cost uint16

	WizardType uint8

	duplicateID          uint8
	purchasedArmour      uint8
	maxPurchasableArmour uint8
	repurchasedTroops    uint8
	maxPurchasableTroops uint8
	bookProfile          []byte

	unknown1 []byte // always 0x00 // TODO: Check this. Maybe not always in save files.
	unknown2 []byte // always 0x00 // TODO: Check this. Maybe not always in save files.
	unknown3 []byte // always 0x00 // TODO: Check this. Maybe not always in save files.
	unknown4 []byte
	unknown5 byte
	unknown6 []byte
	unknown7 []byte
}

type RegimentType int

const (
	RegimentTypeUnknown RegimentType = iota // 0b00000
	RegimentTypeInfantry
	RegimentTypeCavalry
	RegimentTypeArchers
	RegimentTypeArtillery
	RegimentTypeMagicUsers
	RegimentTypeMonsters
	RegimentTypeChariots
	RegimentTypeMisc // 0b01000
)

func (r *Regiment) Type() RegimentType {
	return RegimentType(r.typ >> 3)
}

func (r *Regiment) typeLabel() string {
	switch r.Type() {
	case RegimentTypeInfantry:
		return "Infantry"
	case RegimentTypeCavalry:
		return "Cavalry"
	case RegimentTypeArchers:
		return "Archers"
	case RegimentTypeArtillery:
		return "Artillery"
	case RegimentTypeMagicUsers:
		return "Magic users"
	case RegimentTypeMonsters:
		return "Monsters"
	case RegimentTypeChariots:
		return "Chariots"
	case RegimentTypeUnknown:
		fallthrough
	case RegimentTypeMisc:
		fallthrough
	default:
		return "Unknown"
	}
}

type RegimentRace int

const (
	RegimentRaceHuman RegimentRace = iota // 0b000
	RegimentRaceWoodElf
	RegimentRaceDwarf
	RegimentRaceNightGoblin
	RegimentRaceOrc
	RegimentRaceUndead
	RegimentRaceTownsfolk
	RegimentRaceOgre // 0b111 // TODO: The Imperial Steam Tank sits under this so maybe a different name.
)

func (r *Regiment) Race() RegimentRace {
	return RegimentRace((r.typ >> 0) & ((1 << 3) - 1))
}

func (r *Regiment) raceLabel() string {
	switch r.Race() {
	case RegimentRaceHuman:
		return "Human"
	case RegimentRaceWoodElf:
		return "Woof Elf"
	case RegimentRaceDwarf:
		return "Dwarf"
	case RegimentRaceNightGoblin:
		return "Night Goblin"
	case RegimentRaceOrc:
		return "Orc"
	case RegimentRaceUndead:
		return "Undead"
	case RegimentRaceTownsfolk:
		return "Townsfolk"
	case RegimentRaceOgre:
		return "Ogre"
	default:
		return "Unknown"
	}
}

type troopAttributes struct {
	Movement       uint8
	WeaponSkill    uint8
	BallisticSkill uint8
	Strength       uint8
	Toughness      uint8
	Wounds         uint8
	Initiative     uint8
	Attacks        uint8
	Leadership     uint8
}

type leader struct {
	// Name is the name of the leader.
	Name string
	// SpriteIndex is the index into the list of sprite file names found in
	// ENGREL.EXE for the leader's sprite. Use the engrel.ReadSpriteName
	// function with this index to find the name of the sprite file to use for
	// the leader's sprite.
	SpriteIndex uint16

	attributes    *troopAttributes
	mount         uint8
	armour        uint8
	weapon        uint8
	unitType      uint8
	pointValue    uint8
	missileWeapon uint8
	// headID is the leader's 3D head ID.
	headID uint16
	x      []byte
	y      []byte
}

var MagicItemSlotIndex uint16 = math.MaxUint16
