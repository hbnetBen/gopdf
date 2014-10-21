package fontmaker

import (
	//"encoding/binary"
	//"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"os"
)

type TTFParser struct {
	tables           map[string]uint64
	unitsPerEm       uint64
	xMin             uint64
	yMin             uint64
	xMax             uint64
	yMax             uint64
	numberOfHMetrics uint64
	numGlyphs        uint64
	widths           []uint64
}

func (me *TTFParser) Parse(fontpath string) error {
	fmt.Printf("start parse\n")
	fd, err := os.Open(fontpath)
	if err != nil {
		return err
	}
	defer fd.Close()
	version, err := me.Read(fd, 4)
	if err != nil {
		return err
	}

	if !me.CompareBytes(version, []byte{0x00, 0x01, 0x00, 0x00}) {
		return errors.New("Unrecognized file (font) format")
	}

	i := uint64(0)
	numTables, err := me.ReadUShort(fd)
	if err != nil {
		return err
	}
	me.Skip(fd, 3*2) //searchRange, entrySelector, rangeShift
	me.tables = make(map[string]uint64)
	for i < numTables {

		tag, err := me.Read(fd, 4)
		if err != nil {
			return err
		}

		err = me.Skip(fd, 4)
		if err != nil {
			return err
		}

		offset, err := me.ReadULong(fd)
		if err != nil {
			return err
		}

		err = me.Skip(fd, 4)
		if err != nil {
			return err
		}
		//fmt.Printf("%s\n", me.BytesToString(tag))
		me.tables[me.BytesToString(tag)] = offset
		i++
	}

	//fmt.Printf("%+v\n", me.tables)

	err = me.ParseHead(fd)
	if err != nil {
		return err
	}

	err = me.ParseHhea(fd)
	if err != nil {
		return err
	}

	err = me.ParseMaxp(fd)
	if err != nil {
		return err
	}
	err = me.ParseHmtx(fd)
	if err != nil {
		return err
	}
	//fmt.Printf("%#v\n", me.widths)
	return nil
}

/*
กำลังทำ
func (me *TTFParser) ParseCmap(fd *os.File) error {

	return nil
}*/

func (me *TTFParser) ParseHmtx(fd *os.File) error {

	me.Seek(fd, "hmtx")
	i := uint64(0)
	for i < me.numberOfHMetrics {
		advanceWidth, err := me.ReadUShort(fd)
		if err != nil {
			return err
		}
		err = me.Skip(fd, 2)
		if err != nil {
			return err
		}
		me.widths = append(me.widths, advanceWidth)
		i++
	}
	if me.numberOfHMetrics < me.numGlyphs {
		var err error
		lastWidth := me.widths[me.numberOfHMetrics-1]
		me.widths, err = me.ArrayPadUint(me.widths, me.numGlyphs, lastWidth)
		if err != nil {
			return err
		}
	}
	return nil
}

func (me *TTFParser) ArrayPadUint(arr []uint64, size uint64, val uint64) ([]uint64, error) {
	var result []uint64
	i := uint64(0)
	for i < size {
		if int(i) < len(arr) {
			result = append(result, arr[i])
		} else {
			result = append(result, val)
		}
		i++
	}

	return result, nil
}

func (me *TTFParser) ParseHead(fd *os.File) error {

	//fmt.Printf("\nParseHead\n")
	err := me.Seek(fd, "head")
	if err != nil {
		return err
	}

	err = me.Skip(fd, 3*4) // version, fontRevision, checkSumAdjustment
	if err != nil {
		return err
	}
	magicNumber, err := me.ReadULong(fd)
	if err != nil {
		return err
	}

	//fmt.Printf("\nmagicNumber = %d\n", magicNumber)
	if magicNumber != 0x5F0F3CF5 {
		return errors.New("Incorrect magic number")
	}

	err = me.Skip(fd, 2)
	if err != nil {
		return err
	}

	me.unitsPerEm, err = me.ReadUShort(fd)
	if err != nil {
		return err
	}

	err = me.Skip(fd, 2*8) // created, modified
	if err != nil {
		return err
	}

	me.xMin, err = me.ReadUShort(fd)
	if err != nil {
		return err
	}

	me.yMin, err = me.ReadUShort(fd)
	if err != nil {
		return err
	}

	me.xMax, err = me.ReadUShort(fd)
	if err != nil {
		return err
	}

	me.yMax, err = me.ReadUShort(fd)
	if err != nil {
		return err
	}

	return nil
}

func (me *TTFParser) ParseHhea(fd *os.File) error {

	err := me.Seek(fd, "hhea")
	if err != nil {
		return err
	}

	err = me.Skip(fd, 4+15*2)
	if err != nil {
		return err
	}

	me.numberOfHMetrics, err = me.ReadUShort(fd)
	if err != nil {
		return err
	}
	return nil
}

func (me *TTFParser) ParseMaxp(fd *os.File) error {
	err := me.Seek(fd, "maxp")
	if err != nil {
		return err
	}
	err = me.Skip(fd, 4)
	if err != nil {
		return err
	}
	me.numGlyphs, err = me.ReadUShort(fd)
	if err != nil {
		return err
	}
	return nil
}

func (me *TTFParser) Seek(fd *os.File, tag string) error {
	val, ok := me.tables[tag]
	if !ok {
		return errors.New("me.tables not contain key=" + tag)
	}
	_, err := fd.Seek(int64(val), 0)
	if err != nil {
		return err
	}
	return nil
}

func (me *TTFParser) BytesToString(b []byte) string {
	return string(b)
}

func (me *TTFParser) ReadUShort(fd *os.File) (uint64, error) {
	buff, err := me.Read(fd, 2)
	if err != nil {
		return 0, err
	}
	num := big.NewInt(0)
	num.SetBytes(buff)
	return num.Uint64(), nil
}

func (me *TTFParser) ReadULong(fd *os.File) (uint64, error) {
	buff, err := me.Read(fd, 4)
	if err != nil {
		return 0, err
	}
	num := big.NewInt(0)
	num.SetBytes(buff)
	return num.Uint64(), nil
}

func (me *TTFParser) Skip(fd *os.File, length int64) error {
	_, err := fd.Seek(int64(length), 1)
	if err != nil {
		return err
	}
	return nil
}

func (me *TTFParser) Read(fd *os.File, length int) ([]byte, error) {
	buff := make([]byte, length)
	readlength, err := fd.Read(buff)
	if err != nil {
		return nil, err
	}
	if readlength != length {
		return nil, errors.New("file out of length")
	}
	//fmt.Printf("%d,%s\n", readlength, string(buff))
	return buff, nil
}

func (me *TTFParser) CompareBytes(a []byte, b []byte) bool {

	if a == nil && b == nil {
		return true
	} else if a == nil && b != nil {
		return false
	} else if a != nil && b == nil {
		return false
	}

	if len(a) != len(b) {
		return false
	}

	i := 0
	length := len(a)
	for i < length {
		if a[i] != b[i] {
			return false
		}
		i++
	}
	return true
}