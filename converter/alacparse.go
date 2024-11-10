package converter

import (
	"bytes"
	"encoding/binary"
	"fmt"
	mp4 "github.com/abema/go-mp4"
	"github.com/sunfish-shogi/bufseekio"
	"os"
	"slices"
	"strings"
)

type Tags struct {
	Title       string
	Album       string
	Artist      string
	Year        string
	AlbumArtist string
	TrackNumber string
	DiscNumber  string
}

var validBoxes = []uint32{
	677587310,  // (c)nam - title
	677587265,  // (c)ART - artists
	677587297,  // (c)alb - album
	677587300,  // (c)day - year
	1953655662, // trkn - track number
	1631670868, // aART - album artist
	1684632427, // disk - disc number
}

const (
	blockSize        = 128 * 1024
	blockHistorySize = 4
)

func ParseALACTags(ALACPath string) (*Tags, error) {
	var tags Tags

	ALACFile, err := os.OpenFile(ALACPath, os.O_RDONLY, 0644)
	if err != nil {
		fmt.Println(err.Error())
	}
	defer func(ALACFile *os.File) {
		_ = ALACFile.Close()
	}(ALACFile)

	readSeeker := bufseekio.NewReadSeeker(ALACFile, blockSize, blockHistorySize)

	_, err = mp4.ReadBoxStructure(readSeeker, func(h *mp4.ReadHandle) (interface{}, error) {
		if h.BoxInfo.Type == mp4.BoxTypeMoov() {
			_, err := h.Expand()
			if err != nil {
				return nil, err
			}
			return nil, nil
		}
		if h.BoxInfo.Type == mp4.BoxTypeUdta() {
			_, err := h.Expand()
			if err != nil {
				return nil, err
			}
			return nil, nil
		}
		if h.BoxInfo.Type == mp4.BoxTypeMeta() {
			_, err := h.Expand()
			if err != nil {
				return nil, err
			}
			return nil, nil
		}
		if h.BoxInfo.Type == mp4.BoxTypeIlst() {
			_, err := h.Expand()
			if err != nil {
				return nil, err
			}
			return nil, nil
		}

		tmp := []byte(h.BoxInfo.Type.String())
		var BoxNum uint32
		if err := binary.Read(bytes.NewReader(tmp), binary.BigEndian, &BoxNum); err != nil {
			return nil, err
		}

		if h.BoxInfo.Context.UnderIlst && slices.Contains(validBoxes, BoxNum) {
			out, err := h.Expand()
			if err != nil {
				return nil, err
			}
			str := fmt.Sprintf("%v", out[0])

			switch BoxNum {
			case 677587310:
				tags.Title = str

			case 677587265:
				tags.Artist = str

			case 677587297:
				tags.Album = str

			case 677587300:
				tags.Year = str

			case 1631670868:
				tags.AlbumArtist = str

			case 1953655662:
				tags.TrackNumber = str

			case 1684632427:
				tags.DiscNumber = str
			}
			return nil, nil
		}
		if h.BoxInfo.Type == mp4.BoxTypeData() && h.BoxInfo.Context.UnderIlst {
			box, _, err := h.ReadPayload()
			if err != nil {
				return nil, err
			}
			str, err := mp4.Stringify(box, h.BoxInfo.Context)
			if err != nil {
				return nil, err
			}
			dataType := strings.TrimPrefix(strings.Split(str, " ")[0], "DataType=")
			if dataType == "BINARY" {
				buf := bytes.NewBuffer(make([]byte, 0, h.BoxInfo.Size-h.BoxInfo.HeaderSize))
				if _, err := h.ReadData(buf); err != nil {
					return nil, err
				}
				for _, bufByte := range buf.Bytes()[:len(buf.Bytes())-2] {
					if bufByte != 0 {
						return bufByte, nil
					}
				}
				return nil, nil
			} else {
				str = strings.Split(str, "Data=")[1]
				str = strings.Trim(str, "\"")
				return str, nil
			}
		}
		return nil, nil
	})
	if err != nil {
		return nil, err
	}

	return &tags, nil
}
