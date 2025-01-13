package models

type ModelMapper interface {
	ToModel() interface{}
}
