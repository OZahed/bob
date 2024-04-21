package reflectutils

import (
	"fmt"
	"reflect"
)

type Error struct {
	Msg     string
	DstType string
	SrcType string
}

func (c Error) Error() string {
	return fmt.Sprintf("could not copy src of %s into dst of %s, error: %s", c.SrcType, c.DstType, c.Msg)
}

func NewError(src, dst reflect.Value, msg string) Error {
	return Error{
		Msg:     msg,
		DstType: dst.String(),
		SrcType: src.String(),
	}
}
