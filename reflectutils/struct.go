package reflectutils

import (
	r "reflect"
)

// TypeCast is a simple implementation for casting one type into another similar data type
func TypeCast(src, dst interface{}) error {
	if src == nil || dst == nil {
		return NewError(r.ValueOf(src), r.ValueOf(dst), "source or destination is nil")
	}

	dstType := r.TypeOf(dst)
	dstValue := r.ValueOf(dst)
	srcValue := r.ValueOf(src)
	srcType := r.TypeOf(src)

	if dstType.Kind() != r.Pointer {
		return NewError(srcValue, dstValue, "destination is not a pointer and can not be assigned")
	}

	if srcType.Kind() == r.Pointer {
		srcType = srcType.Elem()
		srcValue = srcValue.Elem()
	}

	if dstValue.CanSet() {
		return NewError(srcValue, dstValue, "destination value can not be set")
	}

	if !srcValue.IsValid() {
		return NewError(srcValue, dstValue, "source does not represent a value")
	}

	if dstType.Kind() == r.Pointer {
		dstValue = dstValue.Elem()
		dstType = dstType.Elem()
	}

	if srcType.Kind() != dstType.Kind() {
		return NewError(srcValue, dstValue, "source and destination are not of same reflect kind")
	}

	if srcValue.Type() == dstValue.Type() {
		dstValue.Set(srcValue)

		return nil
	}

	switch {
	case srcType.Kind() == r.Slice:
		return CopySlice(srcValue, dstValue)
	case srcType.Kind() == r.Map:
		return CopyMap(srcValue, dstValue)
	case srcType.Kind() == r.Interface:
		return CopyInterface(srcValue, dstValue)
	case srcType.Kind() == r.Func:
		return NewError(srcValue, dstValue, "casting function types is not supported")
	// we should check type equal before checking struct
	case dstType == srcType:
		return copyValue(srcValue, dstValue)
	case srcType.Kind() == r.Struct:
		return CopyStruct(srcValue, dstValue)
	}

	return nil
}

func CopyMap(srcValue, dstValue r.Value) error {
	srcKType := srcValue.Type().Key()
	srcVType := srcValue.Type().Elem()

	dstKType := srcValue.Type().Key()
	dstVType := srcValue.Type().Elem()

	if !(srcKType == dstKType && srcVType == dstVType) {
		return NewError(srcValue, dstValue, "key value types on maps are not compatible")
	}

	if dstValue.IsZero() {
		dstValue.Set(r.MakeMap(dstValue.Type()))
	}

	for _, k := range srcValue.MapKeys() {
		dstValue.SetMapIndex(k, srcValue.MapIndex(k))
	}

	return nil

}

func CopyInterface(srcValue, dstValue r.Value) error {
	if !srcValue.Type().AssignableTo(dstValue.Type()) {
		return NewError(srcValue, dstValue, "can not cast interfaces")
	}

	dstValue.Set(r.ValueOf(srcValue.Interface()))

	return nil
}

func CopySlice(srcSlice, dstSlice r.Value) error {
	srcElm := srcSlice.Type().Elem()
	dstElm := dstSlice.Type().Elem()

	if srcSlice.Len() > dstSlice.Len() {
		dstSlice.Grow(srcSlice.Len() - dstSlice.Len())
	}

	dstSlice.SetLen(srcSlice.Len())

	for i := 0; i < srcSlice.Len(); i++ {
		switch {
		case srcElm.Kind() == r.Interface:
			if err := CopyInterface(srcSlice.Index(i), dstSlice.Index(i)); err != nil {
				return err
			}
		case srcElm.Kind() == r.Map:
			if err := CopyMap(srcSlice.Index(i), dstSlice.Index(i)); err != nil {
				return err
			}
		case srcElm.Kind() == r.Slice:
			if err := CopySlice(srcSlice.Index(i), dstSlice.Index(i)); err != nil {
				return err
			}
		// For built-in or std lib types or common types
		case dstElm == srcElm:
			if err := copyValue(srcSlice, dstSlice); err != nil {
				return err
			}
		case srcElm.Kind() == r.Struct:
			if err := CopyStruct(srcSlice.Index(i), dstSlice.Index(i)); err != nil {
				return err
			}
		}
	}

	return nil
}

func CopyStruct(srcValue, dstValue r.Value) error {
	var srcStructField r.StructField
	var srcFieldValue r.Value

	dstType := dstValue.Type()

	for i := 0; i < srcValue.NumField(); i++ {
		srcFieldValue = srcValue.Field(i)
		srcStructField = srcValue.Type().Field(i)

		if !srcStructField.IsExported() {
			continue
		}

		dstStructField, ok := dstType.FieldByName(srcStructField.Name)
		if !ok {
			continue
		}

		dstFieldValue := dstValue.FieldByName(srcStructField.Name)

		if srcStructField.Type.Kind() != dstStructField.Type.Kind() {
			return NewError(srcFieldValue, dstFieldValue, "are not the same kind")
		}

		// For Almost any data type on standard Library, it would work
		// Unless there is a type alias or Enum or an struct

		switch {
		case srcStructField.Type.Kind() == r.Slice, srcStructField.Type.Kind() == r.Array:
			if err := CopySlice(srcFieldValue, dstFieldValue); err != nil {
				return err
			}
		case srcStructField.Type.Kind() == r.Map:
			if err := CopyMap(srcFieldValue, dstFieldValue); err != nil {
				return err
			}
		case srcStructField.Type.Kind() == r.Interface:
			if err := CopyInterface(srcFieldValue, dstFieldValue); err != nil {
				return err
			}
		case srcStructField.Type == dstStructField.Type:
			if err := copyValue(srcFieldValue, dstFieldValue); err != nil {
				return err
			}
		case srcStructField.Type.Kind() == r.Struct:
			if err := CopyStruct(srcFieldValue, dstFieldValue); err != nil {
				return err
			}
		case srcFieldValue.Type().Kind() == dstFieldValue.Type().Kind():
			if err := copyValue(srcFieldValue, dstFieldValue); err != nil {
				return err
			}
		default:
			return NewError(srcFieldValue, dstFieldValue, "types are incompatible")
		}
	}

	return nil
}

func copyValue(srcFieldValue, dstFieldValue r.Value) error {
	switch {
	// these switch types are required to handle type aliases like Enums
	case srcFieldValue.Type().Kind() == r.Pointer:
		// Preventing direct copy of pointer
		return copyValue(srcFieldValue.Elem(), dstFieldValue.Elem())
	case srcFieldValue.Type() == dstFieldValue.Type():
		dstFieldValue.Set(srcFieldValue)
	case srcFieldValue.Type().Kind() == dstFieldValue.Type().Kind():
		switch dstFieldValue.Type().Kind() {
		case r.Int, r.Int64, r.Int32, r.Int16, r.Int8:
			dstFieldValue.SetInt(srcFieldValue.Int())
		case r.Uint, r.Uint64, r.Uint32, r.Uint16, r.Uint8:
			dstFieldValue.SetUint(srcFieldValue.Uint())
		case r.Float32, r.Float64:
			dstFieldValue.SetFloat(srcFieldValue.Float())
		case r.String:
			dstFieldValue.SetString(srcFieldValue.String())
		}
	default:
		return NewError(srcFieldValue, dstFieldValue, "could not copy value")
	}

	return nil
}
