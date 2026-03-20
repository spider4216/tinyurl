package repository

func New() Repository {
	return Repository{
		store: make(map[string]string),
	}
}

type Repository struct {
	store map[string]string
}

func (r Repository) Insert(k string, v string) {
	r.store[k] = v
}

func (r Repository) Get(k string) string {
	v, ok := r.store[k]

	if !ok {
		return ""
	}

	return v
}
