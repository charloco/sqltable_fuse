package filemodel

import (
	"bytes"
	"compress/gzip"
	"io/ioutil"

	"gorm.io/gorm"
)

type FileModel struct {
	gorm.Model
	Type   string
	Script []byte `gorm:"type:mediumblob"`
}

// 压缩函数
func compress(data []byte) ([]byte, error) {
	var buffer bytes.Buffer
	writer := gzip.NewWriter(&buffer)
	_, err := writer.Write(data)
	writer.Close()
	return buffer.Bytes(), err
}

// 解压函数
func decompress(data []byte) ([]byte, error) {
	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	return ioutil.ReadAll(reader)
}

// 保存
func SaveModel(db *gorm.DB, model FileModel) error {
	compressedScript, err := compress(model.Script)
	if err != nil {
		return err
	}
	model.Script = compressedScript
	result := db.Save(&model)
	return result.Error
}

// 创建
func CreateModel(db *gorm.DB, model FileModel) error {
	compressedScript, err := compress(model.Script)
	if err != nil {
		return err
	}
	model.Script = compressedScript
	result := db.Create(&model)
	return result.Error
}

// 读取
func GetModelByID(db *gorm.DB, id uint) (FileModel, error) {
	var model FileModel
	result := db.First(&model, id)
	if result.Error != nil {
		return model, result.Error
	}
	decompressedScript, err := decompress(model.Script)
	if err != nil {
		return model, err
	}
	model.Script = decompressedScript
	return model, nil
}
func GetAllModels(db *gorm.DB) ([]FileModel, error) {
	var models []FileModel
	result := db.Find(&models)
	if result.Error != nil {
		return models, result.Error
	}
	for index, model := range models {
		decompressedScript, err := decompress(model.Script)
		if err != nil {
			return models, err
		}
		model.Script = decompressedScript
		models[index] = model
	}
	return models, nil
}
