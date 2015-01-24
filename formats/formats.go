package formats

type TOSFormat interface {
	Parse() error
	Decompress(string) error
}
