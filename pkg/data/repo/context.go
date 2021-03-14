package repo

// GormRepository is mostly used for DB injection and customization
type GormRepository interface {
	WithDB(*GormApi)
}