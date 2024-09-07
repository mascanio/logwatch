package appendabletable

type ColBuilder interface {
	GetCols() []Column
	Resize(cols []Column, width int)
}
