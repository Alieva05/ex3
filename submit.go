package main

import (
	"fmt"
	"net/http"
	"net/http/cgi"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

type FormData struct {
	Fio      string
	Phone    string
	Email    string
	Dob      string
	Gender   string
	Bio      string
	Langs    []string
	Contract bool
}

func main() {
	var err error
	err = cgi.Serve(http.HandlerFunc(handler))
	if err != nil {
		fmt.Println("Content-type: text/plain\n")
		fmt.Println("Failed to serve CGI request")
	}
}

func send_sql_request(req string) ([]byte, error) {
	cmd := exec.Command("mysql", "-uu68869", "-p2604335", "-D", "u68869", "-e", req)
	output, err := cmd.CombinedOutput()
	return output, err
}

func validate_data(formData FormData) (bool, string) {

	flag := true
	s := ""

	if formData.Fio == "" {
		flag = false
		s += "Заполните поле ФИО"
	}
	re := regexp.MustCompile(`^([a-zA-z]+\s){2}[a-zA-z]+$`)
	if !re.MatchString(formData.Fio) || len(formData.Fio) > 150 {
		flag = false
		s += "\n Неправильный формат ФИО. Введите латиницей."
	}

	if formData.Email == "" {
		flag = false
		s += "\n Введите почту"
	}
	re = regexp.MustCompile(`^[\w\.-_]+@[a-zA-Z]+\.[a-zA-z]+$`)
	if !re.MatchString(formData.Email) {
		flag = false
		s += "Введите адрес почты корректно, она должна соответствовать форме adress@mail.domen"
	}

	if formData.Phone == "" {
		flag = false
		s += "Поле 'Телефон' обязательно для заполнения"
	}
	re = regexp.MustCompile(`^\+\d{11}$`)
	if !re.MatchString(formData.Phone) {
		flag = false
		s += "Введите номер телефона корректно, он должен начинаться с + и после этого содержать 11 цифр"
	}

	if len(formData.Langs) == 0 {
		flag = false
		s += "Выюерите хотя бы один язык программирования"
	}

	re = regexp.MustCompile(`^\d{4}(-\d{2}){2}$`)
	if !re.MatchString(formData.Dob) {
		flag = false
		s += "Введите дату нууу када вы родились"
	}

	if formData.Bio == "" {
		flag = false
		s += "Ну когоч биоггафию нада"
	}

	if formData.Contract == false {
		flag = false
		s += "Здеся нада нуу галочку тык-тык"
	}

	if len(s) == 0 {
		s = "Ваши данные успешно сохранены!"
	}

	return flag, fmt.Sprintf("'%s'", s)
}

func handler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		w.Header().Set("Content-Type", "text/html")
		http.ServeFile(w, r, "index.html")
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

	formData := FormData{
		Fio:      r.FormValue("fullname"),
		Phone:    r.FormValue("phone"),
		Email:    r.FormValue("email"),
		Dob:      r.FormValue("birthdate"),
		Gender:   r.FormValue("gender"),
		Bio:      r.FormValue("bio"),
		Contract: r.FormValue("check") == "on",
		Langs:    r.Form["languages"],
	}
	flag := 0
	if formData.Contract == true {
		flag = 1
	}

	is_valid, val_answer := validate_data(formData)
	if !is_valid {
		alert_message := fmt.Sprintf("./?error='%s'", val_answer)
		http.Redirect(w, r, alert_message, http.StatusSeeOther)
		return
	}

	req := fmt.Sprintf("INSERT INTO users (fio, gender, phone, mail, date, bio, contact) VALUES ('%s', '%s', '%s', '%s', '%s', '%s', %d);", formData.Fio,
		formData.Gender, formData.Phone, formData.Email, formData.Dob, formData.Bio, flag)
	output, err := send_sql_request(req)
	if err != nil {
		fmt.Fprint(w, err, string(output))
		fmt.Fprintf(w, "'%s'", formData.Gender)
	}
	req = "SELECT MAX(id) FROM users;"
	output, err = send_sql_request(req)
	last_user_id, err := strconv.Atoi(strings.Split(string(output), "\n")[2])
	for _, lang_id := range formData.Langs {
		lang, _ := strconv.Atoi(lang_id)
		req = fmt.Sprintf("INSERT INTO languages_on_user (user_id, lang_id) VALUES (%d, %d);", last_user_id, lang)
		output, err = send_sql_request(req)
	}
	fmt.Fprint(w, val_answer)
}
