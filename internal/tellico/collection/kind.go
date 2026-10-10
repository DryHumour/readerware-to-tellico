package collection

// Kind identifies the collection type being converted.
type Kind string

const (
	KindBooks Kind = "books"
	KindMusic Kind = "music"
	KindVideo Kind = "video"
)

type UnknownKindError Kind

func (e UnknownKindError) Error() string {
	return "unknown kind " + string(e)
}
