package main

const (
	t_none        = TypeRepr("")
	t_string      = TypeRepr("string")
	t_int         = TypeRepr("int")
	t_int32       = TypeRepr("int32")
	t_int64       = TypeRepr("int64")
	t_float64     = TypeRepr("float64")
	t_float32     = TypeRepr("float32")
	t_bool        = TypeRepr("bool")
	t_stringSlice = TypeRepr("[]string")
)

const (
	z_string = `""`
	z_int    = "0"
	z_bool   = "false"
	z_nil    = "nil"
)
