package api

// -- API v1 --

func (api *API) setupRoutesV1() {
	api.engine.GET("/v1/ping", api.srv.HandlePing)
}
