package repository

func New(store map[string]string) Repository {
	return Repository{
		store: store,
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
