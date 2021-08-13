package hotline

import (
	"encoding/binary"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

const defaultCreator = "TTXT"
const defaultType = "TEXT"

var fileCreatorCodes = map[string]string{
	"sit": "SIT!",
	"pdf": "CARO",
}

var fileTypeCodes = map[string]string{
	"sit": "SIT!",
	"jpg": "JPEG",
	"pdf": "PDF ",
}

func fileTypeFromFilename(fn string) string {
	ext := strings.Split(fn, ".")
	code := fileTypeCodes[ext[len(ext)-1]]

	if code == "" {
		code = defaultType
	}

	return code
}

func fileCreatorFromFilename(fn string) string {
	ext := strings.Split(fn, ".")
	code := fileCreatorCodes[ext[len(ext)-1]]
	if code == "" {
		code = defaultCreator
	}

	return code
}

func getFileNameList(filePath string) (fields []Field, err error) {
	files, err := ioutil.ReadDir(filePath)
	if err != nil {
		return fields, nil
	}

	for _, file := range files {
		var fileType []byte
		var fnwi FileNameWithInfo
		fileCreator := make([]byte, 4)
		//fileSize := make([]byte, 4)
		if !file.IsDir() {
			fileType = []byte(fileTypeFromFilename(file.Name()))
			fileCreator = []byte(fileCreatorFromFilename(file.Name()))

			binary.BigEndian.PutUint32(fnwi.FileSize[:], uint32(file.Size()))
			copy(fnwi.Type[:], fileType[:])
			copy(fnwi.Creator[:], fileCreator[:])
		} else {
			fileType = []byte("fldr")

			dir, err := ioutil.ReadDir(filePath + "/" + file.Name())
			if err != nil {
				return fields, err
			}
			binary.BigEndian.PutUint32(fnwi.FileSize[:], uint32(len(dir)))
			copy(fnwi.Type[:], fileType[:])
			copy(fnwi.Creator[:], fileCreator[:])
		}

		nameSize := make([]byte, 2)
		binary.BigEndian.PutUint16(nameSize, uint16(len(file.Name())))
		copy(fnwi.NameSize[:], nameSize[:])

		fnwi.name = []byte(file.Name())

		b, err := fnwi.MarshalBinary()
		if err != nil {
			return nil, err
		}
		fields = append(fields, NewField(fieldFileNameWithInfo, b))
	}

	return fields, nil
}

func CalcTotalSize(filePath string) ([]byte, error) {
	var totalSize uint32
	err := filepath.Walk(filePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		totalSize += uint32(info.Size())

		return nil
	})
	if err != nil {
		return nil, err
	}

	bs := make([]byte, 4)
	binary.BigEndian.PutUint32(bs, totalSize)

	return bs, nil
}

func CalcItemCount(filePath string) ([]byte, error) {
	var itemcount uint16
	err := filepath.Walk(filePath, func(path string, info os.FileInfo, err error) error {
		itemcount += 1

		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	bs := make([]byte, 2)
	binary.BigEndian.PutUint16(bs, itemcount-1)

	return bs, nil
}

func EncodeFilePath(filePath string) []byte {
	pathSections := strings.Split(filePath, "/")
	pathItemCount := make([]byte, 2)
	binary.BigEndian.PutUint16(pathItemCount, uint16(len(pathSections)))

	bytes := pathItemCount

	for _, section := range pathSections {
		bytes = append(bytes, []byte{0, 0}...)

		pathStr := []byte(section)
		bytes = append(bytes, byte(len(pathStr)))
		bytes = append(bytes, pathStr...)
	}

	return bytes
}
