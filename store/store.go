package store

type Store interface {
	GetByID(id string) (string, error)
	Set(id, value string) error
}
