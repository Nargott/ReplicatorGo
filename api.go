package main

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"net/http"
)

type API struct {
	r    *mux.Router
	conf *Config
}

func newAPI(conf *Config) *API {
	api := &API{
		r:    mux.NewRouter(),
		conf: conf,
	}

	return api
}

func (api *API) ConfigureRoutes() error {
	api.r.HandleFunc("/", api.HomeHandler).Methods("GET")
	api.r.HandleFunc("/health", api.HealthHandler).Methods("GET")
	api.r.HandleFunc("/groups", api.GroupsHandler).Methods("GET")
	//api.r.HandleFunc("/groups_html", ArticlesHandler).Methods("GET")

	return http.ListenAndServe(":8181", api.r)
}

func (api *API) HomeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte("Signal Bot\n"))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		Rlog.Error("HomeHandler Write Error: %v", err)
	}
}

func (api *API) HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	healthResponse := make(map[string]string)
	healthResponse["status"] = "UP"

	resp, _ := json.Marshal(healthResponse)

	_, err := w.Write(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		Rlog.Error("HealthHandler Write Error: %v", err)
	}
}

func (api *API) GroupsHandler(w http.ResponseWriter, r *http.Request) {
	groups, err := GetGroupsList(api.conf)
	if err != nil {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusInternalServerError)
		Rlog.Errorf("GroupsHandler GetGroupsList Error: %v", err)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	groupsResponse := make([]map[string]string, len(groups))
	for i, group := range groups {
		groupsResponse[i] = make(map[string]string)
		groupsResponse[i]["name"] = group.Name
		groupsResponse[i]["id"] = group.Id
		groupsResponse[i]["internal_id"] = group.InternalId
	}

	resp, err := json.Marshal(groupsResponse)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		Rlog.Errorf("GroupsHandler json.Marshal Error: %v", err)
	}

	_, err = w.Write(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		Rlog.Errorf("GroupsHandler Write Error: %v", err)
	}
}
