package filemodel

import (
	"fmt"
	"log"
	"os"
	"regexp"

	"github.com/jacobsa/fuse/fuseops"
	"github.com/jacobsa/fuse/fuseutil"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const (
	rootNodeID fuseops.InodeID = 1
	baseID     fuseops.InodeID = 2
)

type InodeInfo struct {
	Attributes fuseops.InodeAttributes

	// File or directory?
	Dir bool

	// For directories, children.
	Children []fuseutil.Dirent
	RowID    uint
}

var db *gorm.DB

func GetDB() *gorm.DB {
	return db
}
func OpenDB(dsn string) error {
	//gorm.Open()
	var err error
	pat := regexp.MustCompile(`^([\d\w]+)://(.*)`)
	matched := pat.FindStringSubmatch(dsn)
	if len(matched) < 3 {
		return fmt.Errorf("DSN格式错误")
	}
	switch matched[1] {
	case "mysql":
		db, err = gorm.Open(mysql.Open(matched[2]))
		if err != nil {
			return err
		}
	case "sqlite":
		db, err = gorm.Open(sqlite.Open(matched[2]))
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("未知的数据库类型")
	}
	db.AutoMigrate(&FileModel{})
	return nil
}

func GetNodeInfos() (map[fuseops.InodeID]InodeInfo, error) {
	models, err := GetAllModels(db)
	if err != nil {
		return nil, err
	}
	m := map[fuseops.InodeID]InodeInfo{}
	entries := []fuseutil.Dirent{}
	for index, model := range models {
		node := InodeInfo{
			Attributes: fuseops.InodeAttributes{
				Nlink: 1,
				Mode:  0777,
				Size:  uint64(len(model.Script)),
			},
			RowID: model.ID,
		}

		name := fmt.Sprintf("%d_%s", model.ID, model.Type)
		nodeid := baseID + fuseops.InodeID(model.ID)
		log.Println("nodeid:", nodeid)
		entries = append(entries, fuseutil.Dirent{
			Offset: fuseops.DirOffset(index + 1),
			Inode:  nodeid,
			Name:   name,
			Type:   fuseutil.DT_File,
		})
		m[nodeid] = node
	}
	rootNode := InodeInfo{
		Attributes: fuseops.InodeAttributes{
			Nlink: 1,
			Mode:  0555 | os.ModeDir,
		},
		Dir:      true,
		Children: entries,
	}
	m[rootNodeID] = rootNode
	return m, nil
}
