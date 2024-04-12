package groups

type Sub string

const (
	WithDirectory Sub = "with_directory"
)

func (s Sub) String() string {
	return string(s)
}
