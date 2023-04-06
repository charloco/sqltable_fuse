package filemodel

import "testing"

func init() {
	OpenDB("sqlite://test.db")
	GetDB().Migrator().DropTable("file_models")
	GetDB().AutoMigrate(&FileModel{})
}
func TestAddModel(t *testing.T) {
	model := FileModel{
		Type:   "firewall",
		Script: []byte("acl"),
	}
	CreateModel(GetDB(), model)
	model = FileModel{
		Type:   "waf",
		Script: []byte("acl"),
	}
	CreateModel(GetDB(), model)
	models, err := GetAllModels(GetDB())
	if err != nil {
		t.Fatal("create model failed")
	}
	if len(models) < 2 {
		t.Fatal("create model failed")
	}
	t.Log(models)
}
