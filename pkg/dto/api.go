package dto

/*
 * API JSON Representations
 */

type ErrorJSON struct {
	Error string `json:"err"`
}

func MakeErrorJSON(err error) ErrorJSON {
	return ErrorJSON{
		Error: err.Error(),
	}
}

type PingJSON struct {
	Msg string `json:"msg"`
}

func MakePingJSON() PingJSON {
	return PingJSON{
		Msg: "pong",
	}
}
