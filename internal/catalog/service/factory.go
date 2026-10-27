package service

type StoreFactory struct {
	path string
}

func NewStoreFactory(path string) StoreFactory {
	return StoreFactory{path: path}
}

func (factory StoreFactory) Create() (*Store, error) {
	return OpenStore(factory.path)
}
