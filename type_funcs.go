package main

func ZeroValueByType(typ string) string {
	switch typ {
	case t_string:
		return z_string
	case t_int, t_int64, t_int32:
		return z_int
	case t_float64, t_float32:
		return z_int
	case t_bool:
		return z_bool
	}
	return ""
}
