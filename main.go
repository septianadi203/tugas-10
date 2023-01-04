package main

import (
	"context"
	"fmt"
	"log"
	"personal-web/connection"
	"strconv"
	"text/template"
	"time"

	// "log"
	"net/http"
	// "strconv"
	// "text/template"
	// "time"

	"github.com/gorilla/mux"
)

var Data = map[string]interface{}{
	"Title":   "Personal web",
	"IsLogin": false,
}

type dataProject struct {
	Id           int
	ProjectName  string
	StartDate    time.Time
	EndDate      time.Time
	Description  string
	Technologies []string
	Duration     string
}

var Projects = []dataProject{}

func main() {
	route := mux.NewRouter()

	connection.DatabaseConnect()

	route.PathPrefix("/public/").Handler(http.StripPrefix("/public/", http.FileServer(http.Dir("./public"))))

	route.HandleFunc("/home", index).Methods("GET")
	route.HandleFunc("/project", projectForm).Methods("GET")
	route.HandleFunc("/project/{id}", projectDetail).Methods("GET")
	route.HandleFunc("/project", projectAdd).Methods("POST")
	route.HandleFunc("/contact", contactMe).Methods("GET")
	route.HandleFunc("/delete/{id}", deleteProject).Methods("GET")
	route.HandleFunc("/edit/{id}", editProject).Methods("GET")
	route.HandleFunc("/editInput/{id}", editProjectInput).Methods("POST")


	port := 5000
	fmt.Println("Server is running on port", port)
	http.ListenAndServe("localhost:5000", route)
}

func index(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	// var tmpl, err = template.ParseFiles("views/index.html")

	var tmpl, err = template.ParseFiles("views/index.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message :" + err.Error()))
		return
	}

	rows, _ := connection.Conn.Query(context.Background(), "SELECT * FROM tb_project")

	var result []dataProject
	for rows.Next() {
		var each = dataProject{}

		var err = rows.Scan(&each.Id, &each.ProjectName, &each.StartDate, &each.EndDate, &each.Description, &each.Technologies)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		each.Duration = selisihDate(each.StartDate, each.EndDate)
		result = append(result, each)
	}

	respData := map[string]interface{}{
		"Data":     Data,
		"Projects": result,
	}

	fmt.Println(result)
	w.WriteHeader(http.StatusOK)
	tmpl.Execute(w, respData)
}

func selisihDate(start time.Time, end time.Time) string {

	distance := end.Sub(start)

	// Menghitung durasi
	//pengkondisian
	var duration string
	year := int(distance.Hours() / (12 * 30 * 24))
	if year != 0 {
		duration = strconv.Itoa(year) + " tahun"
	} else {
		month := int(distance.Hours() / (30 * 24))
		if month != 0 {
			duration = strconv.Itoa(month) + " bulan"
		} else {
			week := int(distance.Hours() / (7 * 24))
			if week != 0 {
				duration = strconv.Itoa(week) + " minggu"
			} else {
				day := int(distance.Hours() / (24))
				if day != 0 {
					duration = strconv.Itoa(day) + " hari"
				}
			}
		}
	}
	return duration
}

func projectForm(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	var tmpl, err = template.ParseFiles("views/my-project.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message :" + err.Error()))
		return
	}

	respData := map[string]interface{}{
		"Data":     Data,
		"Projects": Projects,
	}

	w.WriteHeader(http.StatusOK)
	tmpl.Execute(w, respData)
}

func projectDetail(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	var tmpl, err = template.ParseFiles("views/project-detail.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message :" + err.Error()))
		return
	}

	id, _ := strconv.Atoi(mux.Vars(r)["id"])

	ProjectDetail := dataProject{}

	err = connection.Conn.QueryRow(context.Background(), "SELECT id, project_name, start_date, end_date, description, technologies FROM tb_project WHERE id=$1", id).Scan(
		&ProjectDetail.Id, &ProjectDetail.ProjectName, &ProjectDetail.StartDate, &ProjectDetail.EndDate, &ProjectDetail.Description, &ProjectDetail.Technologies,
	)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message : " + err.Error()))
		return
	}

	ProjectDetail.Duration = selisihDate(ProjectDetail.StartDate, ProjectDetail.EndDate)

	respDataDetail := map[string]interface{}{
		"Data": Data,
		
		"ProjectDetail": ProjectDetail,
	}

	w.WriteHeader(http.StatusOK)
	tmpl.Execute(w, respDataDetail)
}

func projectAdd(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Fatal(err)
	}

	projectName := r.PostForm.Get("name")
	startDate := r.PostForm.Get("start-date")
	endDate := r.PostForm.Get("end-date")
	description := r.PostForm.Get("message")
	techStack := r.Form["project-tech"]

	// Menghitung durasi
	// Parsing string ke time
	//dari string starDate ke time startDateTime

	// Start Date
	startDateTime, _ := time.Parse("2006-01-02", startDate)

	// End Date
	endDateTime, _ := time.Parse("2006-01-02", endDate)

	_, err = connection.Conn.Exec(context.Background(), "INSERT INTO tb_project( project_name, start_date, end_date, description, technologies ) VALUES ($1,$2,$3,$4,$5)", projectName, startDateTime, endDateTime, description, techStack)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        w.Write([]byte("message : " + err.Error()))
        return
    }


	http.Redirect(w, r, "/home", http.StatusMovedPermanently)
}

func contactMe(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	var tmpl, err = template.ParseFiles("views/contact-form.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message :" + err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	tmpl.Execute(w, Data)
}

func deleteProject(w http.ResponseWriter, r *http.Request) {
	// ngambil id pake ini
	id, _ := strconv.Atoi(mux.Vars(r)["id"])

	_, err := connection.Conn.Exec(context.Background(), "DELETE FROM tb_project WHERE id=$1", id)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        w.Write([]byte("message : " + err.Error()))
        return
    }

	// Projects = append(Projects[:id], Projects[id+1:]...)

	http.Redirect(w, r, "/home", http.StatusFound)
}

func editProject(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	var tmpl, err = template.ParseFiles("views/edit.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message :" + err.Error()))
		return
	}

	id, _ := strconv.Atoi(mux.Vars(r)["id"])

	ProjectDetail := dataProject{}

	err = connection.Conn.QueryRow(context.Background(), "SELECT * FROM tb_project WHERE id=$1", id).Scan(
		&ProjectDetail.Id, &ProjectDetail.ProjectName, &ProjectDetail.StartDate, &ProjectDetail.EndDate, &ProjectDetail.Description, &ProjectDetail.Technologies,
	)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message : " + err.Error()))
		return
	}


	respData := map[string]interface{}{
		"Data":          Data,
		"ProjectDetail": ProjectDetail,
	}

	w.WriteHeader(http.StatusOK)
	tmpl.Execute(w, respData)
}

func editProjectInput(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Fatal(err)
	}

	projectName := r.PostForm.Get("name")
	startDate := r.PostForm.Get("start-date")
	endDate := r.PostForm.Get("end-date")
	description := r.PostForm.Get("message")
	techStack := r.Form["project-tech"]

	// Menghitung durasi
	// Parsing string ke time
	//dari string starDate ke time startDateTime

	id, _ := strconv.Atoi(mux.Vars(r)["id"])

	// Start Date
	startDateTime, _ := time.Parse("2006-01-02", startDate)

	// End Date
	endDateTime, _ := time.Parse("2006-01-02", endDate)

	_, err = connection.Conn.Exec(context.Background(), "UPDATE tb_project SET project_name = $1, start_date = $2, end_date = $3, description = $4, technologies = $5 WHERE id = $6", projectName, startDateTime, endDateTime, description, techStack, id)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        w.Write([]byte("message : " + err.Error()))
        return
    }

	http.Redirect(w, r, "/home", http.StatusMovedPermanently)
}
