package collection

// Kind identifies the collection type being converted.
type Kind string

const (
	KindBooks Kind = "books"
	KindMusic Kind = "music"
	KindVideo Kind = "video"
)

type ErrUnknownKind Kind

func (e ErrUnknownKind) Error() string {
	return "unknown kind " + string(e)
}
