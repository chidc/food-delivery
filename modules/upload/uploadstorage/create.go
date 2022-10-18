package uploadstorage

import (
	"context"
	"demo/common"
)

func (store *sqlStore) CreateImage(context context.Context, data *common.Image) error {
	db := store.db

	if err := db.Table(data.TableName()).
		Create(data).Error; err != nil {
		return common.ErrDB(err)
	}

	return nil
}
