package registry

const (
	DWORD = 4
	SZ    = 1
)

type Registry interface {
	SetDWORD(key, name string, value uint32) error
	SetString(key, name, value string) error
	Delete(key, name string) error
	KeyExists(key string) (bool, error)
	ValueExists(key, name string) (bool, error)
}
