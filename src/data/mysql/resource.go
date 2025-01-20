package mysqldb

import (
	"database/sql"

	"github.com/pchchv/aas/src/models"
)

func (d *MySQLDB) CreateResource(tx *sql.Tx, resource *models.Resource) error {
	return d.CommonDB.CreateResource(tx, resource)
}

func (d *MySQLDB) UpdateResource(tx *sql.Tx, resource *models.Resource) error {
	return d.CommonDB.UpdateResource(tx, resource)
}

func (d *MySQLDB) GetResourceById(tx *sql.Tx, resourceId int64) (*models.Resource, error) {
	return d.CommonDB.GetResourceById(tx, resourceId)
}

func (d *MySQLDB) GetResourceByResourceIdentifier(tx *sql.Tx, resourceIdentifier string) (*models.Resource, error) {
	return d.CommonDB.GetResourceByResourceIdentifier(tx, resourceIdentifier)
}

func (d *MySQLDB) GetResourcesByIds(tx *sql.Tx, resourceIds []int64) ([]models.Resource, error) {
	return d.CommonDB.GetResourcesByIds(tx, resourceIds)
}

func (d *MySQLDB) GetAllResources(tx *sql.Tx) ([]models.Resource, error) {
	return d.CommonDB.GetAllResources(tx)
}

func (d *MySQLDB) DeleteResource(tx *sql.Tx, resourceId int64) error {
	return d.CommonDB.DeleteResource(tx, resourceId)
}
