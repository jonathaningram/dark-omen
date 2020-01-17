package arm

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"path"
	"strings"

	"github.com/jonathaningram/dark-omen/internal/cstringutil"
)

var (
	// format is the army format ID used in all .ARM and save files.
	format = [4]byte{0x9e, 0x02, 0x00, 0x00}
)

const (
	saveHeaderSize = 504
	headerSize     = 192
)

type Army struct {
	Regiments                  []*Regiment
	SmallBannerPath            string
	smallBannerPathRaw         string
	SmallBannerDisabledPath    string
	smallBannerDisabledPathRaw string
	LargeBannerPath            string
	largeBannerPathRaw         string
	GoldFromTreasures          uint16
	GoldInCoffers              uint16
	MagicItems                 []byte
}

// Decoder reads and decodes army information from an input stream.
type Decoder struct {
	r io.ReaderAt
}

// NewDecoder returns a new decoder that reads from r.
func NewDecoder(r io.ReaderAt) *Decoder {
	return &Decoder{r: r}
}

// Decode reads the encoded army information from its input and returns a new
// Army containing decoded information including regiments, their status and
// some army data and statistics.
func (d *Decoder) Decode() (*Army, error) {
	// Check if this is a .ARM file or a save file.
	buf := make([]byte, 4)
	n, err := d.r.ReadAt(buf, 0)
	if n != 4 {
		return nil, fmt.Errorf("only read %d bytes, expected %d", n, 4)
	}
	if err != nil {
		return nil, err
	}
	var header *header
	var startPos int64
	if !bytes.Equal(buf, format[:]) {
		_, err := d.readSaveHeader()
		if err != nil {
			return nil, err
		}
		startPos = saveHeaderSize
	}
	header, err = d.readHeader(startPos)
	if err != nil {
		return nil, err
	}

	regiments, err := d.readRegiments(header, startPos)
	if err != nil {
		return nil, err
	}

	army := &Army{
		Regiments:                  regiments,
		SmallBannerPath:            normalizeBooksPath(header.smallBannerPath),
		smallBannerPathRaw:         header.smallBannerPath,
		SmallBannerDisabledPath:    normalizeBooksPath(header.smallBannerDisabledPath),
		smallBannerDisabledPathRaw: header.smallBannerDisabledPath,
		LargeBannerPath:            normalizeBooksPath(header.largeBannerPath),
		largeBannerPathRaw:         header.largeBannerPath,
		GoldFromTreasures:          header.goldFromTreasures,
		GoldInCoffers:              header.goldInCoffers,
		MagicItems:                 header.magicItems,
	}

	return army, nil
}

type header struct {
	format                  uint16
	regimentCount           uint16
	regimentBlockSize       uint16
	race                    uint8
	unknown1                []byte // purpose of bytes at index 13, 14, 15 is unknown
	defaultName             string
	armyName                string
	smallBannerPath         string
	smallBannerDisabledPath string
	largeBannerPath         string
	goldFromTreasures       uint16
	goldInCoffers           uint16
	magicItems              []byte
	unknown2                []byte // purpose of bytes at index 190 and 191 is unknown
}

func (d *Decoder) readHeader(startPos int64) (*header, error) {
	buf := make([]byte, headerSize)
	n, err := d.r.ReadAt(buf, startPos)
	if n != headerSize {
		return nil, fmt.Errorf("only read %d bytes, expected %d", n, headerSize)
	}
	if err != nil {
		return nil, err
	}

	return &header{
		format:                  binary.LittleEndian.Uint16(buf[0:4]),
		regimentCount:           binary.LittleEndian.Uint16(buf[4:8]),
		regimentBlockSize:       binary.LittleEndian.Uint16(buf[8:12]),
		race:                    uint8(buf[12]),
		unknown1:                buf[13:16],
		defaultName:             cstringutil.ToGo(buf[16:18]),
		armyName:                cstringutil.ToGo(buf[18:50]),
		smallBannerPath:         cstringutil.ToGo(buf[50:82]),
		smallBannerDisabledPath: cstringutil.ToGo(buf[82:114]),
		largeBannerPath:         cstringutil.ToGo(buf[114:146]),
		goldFromTreasures:       binary.LittleEndian.Uint16(buf[146:148]),
		goldInCoffers:           binary.LittleEndian.Uint16(buf[148:150]),
		magicItems:              buf[150:190],
		unknown2:                buf[190:192],
	}, nil
}

func (d *Decoder) readRegiments(header *header, startPos int64) ([]*Regiment, error) {
	regiments := make([]*Regiment, header.regimentCount)

	for i := uint16(0); i < header.regimentCount; i++ {
		buf := make([]byte, header.regimentBlockSize)
		_, err := d.r.ReadAt(buf, startPos+int64(headerSize+i*header.regimentBlockSize))
		if err != nil {
			if err == io.EOF {
				return nil, fmt.Errorf("army does not contain enough regiments, expected to find %d, but got EOF while reading regiment at index %d: %w", header.regimentCount, i, io.ErrUnexpectedEOF)
			}
			return nil, err
		}

		magicItems := [3]uint16{
			binary.LittleEndian.Uint16(buf[162:164]),
			binary.LittleEndian.Uint16(buf[164:166]),
			binary.LittleEndian.Uint16(buf[166:168]),
		}

		regiments[i] = &Regiment{
			status:             buf[0:2],
			unknown1:           buf[2:4],
			id:                 binary.LittleEndian.Uint16(buf[4:6]),
			unknown2:           buf[6:8],
			WizardType:         buf[8],
			MaxArmour:          buf[9],
			Cost:               binary.LittleEndian.Uint16(buf[10:12]),
			BannerIndex:        binary.LittleEndian.Uint16(buf[12:14]),
			unknown3:           buf[14:16],
			regimentAttributes: buf[16:20],
			SpriteIndex:        binary.LittleEndian.Uint16(buf[20:22]),
			Name:               cstringutil.ToGo(buf[22:54]),
			nameID:             binary.LittleEndian.Uint16(buf[54:56]),
			alignment:          buf[56],
			MaxTroops:          buf[57],
			AliveTroops:        buf[58],
			ranks:              buf[59],
			unknown4:           buf[60:64],
			troopAttributes: &troopAttributes{
				Movement:       buf[64],
				WeaponSkill:    buf[65],
				BallisticSkill: buf[66],
				Strength:       buf[67],
				Toughness:      buf[68],
				Wounds:         buf[69],
				Initiative:     buf[70],
				Attacks:        buf[71],
				Leadership:     buf[72],
			},
			mount:         buf[73],
			armour:        buf[74],
			weapon:        buf[75],
			typ:           buf[76],
			pointValue:    buf[77],
			missileWeapon: buf[78],
			unknown5:      buf[79],
			unknown6:      buf[80:84],
			Leader: &leader{
				SpriteIndex: binary.LittleEndian.Uint16(buf[84:86]),
				Name:        cstringutil.ToGo(buf[86:118]),
				attributes: &troopAttributes{
					Movement:       buf[127],
					WeaponSkill:    buf[128],
					BallisticSkill: buf[129],
					Strength:       buf[130],
					Toughness:      buf[131],
					Wounds:         buf[132],
					Initiative:     buf[133],
					Attacks:        buf[134],
					Leadership:     buf[135],
				},
				mount:         buf[136],
				armour:        buf[137],
				weapon:        buf[138],
				unitType:      buf[139],
				pointValue:    buf[140],
				missileWeapon: buf[141],
				headID:        binary.LittleEndian.Uint16(buf[146:148]),
				x:             buf[148:152],
				y:             buf[152:156],
			},
			unknown7:             buf[142:146],
			Experience:           binary.LittleEndian.Uint16(buf[156:158]),
			duplicateID:          buf[158],
			MinArmour:            buf[159],
			MagicBook:            binary.LittleEndian.Uint16(buf[160:162]),
			MagicItems:           magicItems,
			purchasedArmour:      buf[180],
			maxPurchasableArmour: buf[181],
			repurchasedTroops:    buf[182],
			maxPurchasableTroops: buf[183],
			bookProfile:          buf[184:188],
		}
	}

	return regiments, nil
}

func normalizeBooksPath(p string) string {
	return path.Join(strings.Split(strings.ReplaceAll(p, "[BOOKS]", "BOOKS"), `\`)...)
}
