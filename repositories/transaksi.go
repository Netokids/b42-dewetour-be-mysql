package repositories

import (
	"Backend/models"

	"gorm.io/gorm"
)

type TransaksiRepository interface {
	FindTransaksi() ([]models.Transaction, error)
	GetTransaksi(ID int) (models.Transaction, error)
	AddTransaksi(transaksi models.Transaction) (models.Transaction, error)
	UpdateTransaksi(status string, ID int) (models.Transaction, error)
	DeleteTransaksi(transaksi models.Transaction) (models.Transaction, error)
}

func RepositoryTransaksi(db *gorm.DB) *repository {
	return &repository{db}
}

func (r *repository) FindTransaksi() ([]models.Transaction, error) {
	var Transaksi []models.Transaction
	err := r.db.Preload("Trip").Preload("Trip.Country").Preload("User").Find(&Transaksi).Error

	return Transaksi, err
}

func (r *repository) GetTransaksi(ID int) (models.Transaction, error) {
	var Transaksi models.Transaction
	err := r.db.Preload("Trip").Preload("Trip.Country").Preload("User").First(&Transaksi, ID).Error

	return Transaksi, err
}

func (r *repository) AddTransaksi(transaksi models.Transaction) (models.Transaction, error) {
	err := r.db.Create(&transaksi).Error
	return transaksi, err
}

func (r *repository) UpdateTransaksi(status string, ID int) (models.Transaction, error) {
	// err := r.db.Exec("UPDATE transaksis SET counter_qty=?, total=?, status=?, image=?, trip_id=? WHERE transaksis.id=?", transaksi.CounterQTY, transaksi.Total, transaksi.Status, transaksi.Image, transaksi.Trip_id, transaksi.ID).Error
	var transaction models.Transaction
	r.db.Preload("Trip").First(&transaction, ID)

	// If is different & Status is "success" decrement product quantity
	if status != transaction.Status && status == "success" {
		var Trip models.Trip
		r.db.First(&Trip, transaction.Trip.ID)
		Trip.Kuota = Trip.Kuota - transaction.CounterQTY
		r.db.Save(&Trip)
	}
	transaction.Status = status

	err := r.db.Save(&transaction).Error

	return transaction, err
}

func (r *repository) DeleteTransaksi(transaksi models.Transaction) (models.Transaction, error) {
	err := r.db.Delete(&transaksi).Error

	return transaksi, err
}
