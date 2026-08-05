package finding

type BaselineState string

const (
	BaselineNew      BaselineState = "new"
	BaselineExisting BaselineState = "existing"
)

func (s BaselineState) Valid() bool {
	return s == BaselineNew || s == BaselineExisting
}
