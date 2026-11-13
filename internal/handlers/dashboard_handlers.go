package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/Miroku77/goprxveil/internal/browser"
	"github.com/Miroku77/goprxveil/internal/storage"
)

var db *sql.DB

func InitDB(database *sql.DB) {
	db = database
}

// GET /api/groups - returns a list of user's groups: [{"id": 1, "name": "Laptops"}].
//
// POST /api/groups - accepts {"name": "Laptops"}, saves the group, and returns {"id": 1, "name": "Laptops"}.
func ManipulateGroupHandler(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value("userId").(int)
	if r.Method == http.MethodGet {
		groups, err := storage.FindAllFromGroup(db, userId)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err = json.NewEncoder(w).Encode(groups); err != nil {
			log.Println("Error responce JSON:", err)
		}
	} else if r.Method == http.MethodPost {
		defer r.Body.Close()
		var group = struct {
			Name string `json:"name"`
		}{}
		if err := json.NewDecoder(r.Body).Decode(&group); err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		newIdGroup, err := storage.SaveGroupForUser(db, group.Name, userId)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if err = json.NewEncoder(w).Encode(map[string]any{
			"id":   newIdGroup,
			"name": group.Name,
		}); err != nil {
			log.Println("Error responce JSON:", err)
		}
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}

// DELETE /api/groups/{id}?user_id={2} - delete group by id for user
func DeleteByIdGroupHandler(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value("userId").(int)
	idGroup, err := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/api/groups/"))
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err = storage.DeleteFromGroupById(db, idGroup, userId); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// GET /api/urls?group_id={1} - accepts param id url and return list urls in group {"id":2, "url":"https://..."}
//
// POST /api/urls - accepts {"group_id":2, "url":"https://..."}, save url and return this id
func ManipulateURLHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		group_id, err := strconv.Atoi(r.URL.Query().Get("group_id"))
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		urls, err := storage.FindAllUrlsFromGroup(db, group_id)
		if err != nil {
			log.Println(err)
			return
		}
		type UrlResponce struct {
			Id     int    `json:"id"`
			Url    string `json:"url"`
			Name   string `json:"name"`
			Active bool   `json:"active"`
		}
		var responce = []UrlResponce{}
		for _, u := range urls {
			responce = append(responce, UrlResponce{
				Id:     u.Id,
				Url:    u.Url,
				Name:   u.Name,
				Active: u.Active,
			})
		}
		w.Header().Set("Content-Type", "application/json")
		if err = json.NewEncoder(w).Encode(responce); err != nil {
			log.Println("Error responce JSON:", err)
		}
	} else if r.Method == http.MethodPost {
		defer r.Body.Close()
		var urls = struct {
			GroupId int    `json:"group_id"`
			Url     string `json:"url"`
		}{}
		err := json.NewDecoder(r.Body).Decode(&urls)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		newUrlId, err := storage.SaveUrlForGroup(db, urls.Url, urls.GroupId)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		u, err := url.Parse(urls.Url)
		if err != nil {
			log.Println(err)
			return
		}
		domen := u.Host
		// sckrap
		go browser.FirstAddProductToDB(db, domen, newUrlId, urls.Url)

		w.Header().Set("Content-Type", "application/json")
		if err = json.NewEncoder(w).Encode(map[string]int{"id": newUrlId}); err != nil {
			log.Println("Error responce JSON:", err)
			return
		}
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}

// DELETE /api/urls/{id}?group_id={2} - delete url by id for group. If it's last url from group - group dont remove
func DeleteURLHandler(w http.ResponseWriter, r *http.Request) {
	groupId, err := strconv.Atoi(r.URL.Query().Get("group_id"))
	urlId, err := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/api/urls/"))
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err = storage.DeleteFromUrlById(db, urlId, groupId); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// PATCH /api/urls/{id}/active  	swap active
func UpdateActiveUrlHandler(w http.ResponseWriter, r *http.Request) {
	urlId, err := strconv.Atoi(strings.Split(strings.TrimPrefix(r.URL.Path, "/api/urls/"), "/")[0])
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err = storage.UpdateActiveUrl(db, urlId); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// serves the main control panel page (dashboard.html)
func DashboardHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "templates/dashboard/dashboard.html")
}
