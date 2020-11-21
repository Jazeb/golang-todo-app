package main;

import (
	"io"
	"fmt"
	"net/http"
	"strconv"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"github.com/jinzhu/gorm"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"encoding/json"
)

var db, _ = gorm.Open("mysql", "testuser:password@/GoTodoList?charset=utf8&parseTime=True&loc=Local")

type todoItemModel struct {
	ID int `gorm:"primary_key"`
    Description string
    Completed bool
}

func createItem(w http.ResponseWriter, r *http.Request){
	description := r.FormValue("description")
	log.WithFields(log.Fields{"description":description}).Info("Add new todo item, saving to databse.")
	todo := &todoItemModel{Description:description, Completed:false}
	db.Create(&todo)
	result := db.Last(&todo)
	w.Header().Set("content-type","application/json")
	json.NewEncoder(w).Encode(result.Value)
}
func getHealthz(w http.ResponseWriter, r *http.Request){
	fmt.Print("\nHello getting healths")
	w.Header().Set("Content-Type", "application/json")
	io.WriteString(w, `{"alive":true}`)
}

func updateItem(res http.ResponseWriter, req *http.Request) { // res, req
	// bring schema here and assign it to todo
	todo := &todoItemModel{} // we need to use {} here
	vars := mux.Vars(req) // get id from params like api/1
	id, _ := strconv.Atoi(vars["id"]) // convert string (id) to int 
	item := db.First(&todo, id) // get this item by id from db using todo schema model
	if item.Error != nil {
		log.Warn("Item not found") // check if item exist in the db
		res.Header().Set("Content-Type", "application/json")
		io.WriteString(res, `{"message":"item does not exist"}`)
	} else {
		comp := req.FormValue("completed")
		completed, _ := strconv.ParseBool(comp) // get completed value from body in req module
		todo.Completed = completed // mark completed item as true in todo model
		db.Save(&todo) // save the obj in the db
		res.Header().Set("Content-Type", "application/json") //set resp header 
		io.WriteString(res, `{"completed":true}`) // send resp
		// fmt.Printf("%T", id) int
		// json.NewEncoder(w).Encode(item) //send resp as json
	}
}

func getItemByID(res http.ResponseWriter, req *http.Request)  { // dont know what the * means here
	vars := mux.Vars(req)
	id, _ := strconv.Atoi(vars["id"])
	log.Warn(id)
	res.Header().Set("Content-Type", "application/json")
	todo := &todoItemModel{}
	item := db.First(&todo, id)

	if item.Error != nil {
		io.WriteString(res, `{"message":"item does not exist"}`)
	}else { // cannot use else in next line :)
		json.NewEncoder(res).Encode(item.Value)
	}
	// io.WriteString(res, "please provide id")
}

func deleteItem(res http.ResponseWriter, req *http.Request)  {
	vars := mux.Vars(req)
	id := vars["id"]
	res.Header().Set("Content-Type", "application/json")
	todo := &todoItemModel{}
	db.Delete(&todo, id)
	io.WriteString(res, `{"message":"Item deleted successfully"}`)
}

func completedItems(res http.ResponseWriter, req *http.Request)  {
	var todos []todoItemModel
	TodoItems := db.Where("completed = ?", true).Find(&todos).Value // create an array to store items first
	res.Header().Set("Content-Type", "application/json")
	json.NewEncoder(res).Encode(TodoItems)
}

func main() {
	defer db.Close()
	// db.Debug().DropTableIfExists(&todoItemModel{})
	// db.Debug().AutoMigrate(&todoItemModel{})
	fmt.Print("Starting golang server...")
	router := mux.NewRouter();
	router.HandleFunc("/health", getHealthz).Methods("GET");
	router.HandleFunc("/todo/create", createItem).Methods("POST");
	router.HandleFunc("/getbyid/{id}", getItemByID).Methods("GET");
	router.HandleFunc("/update/{id}", updateItem).Methods("PUT");
	router.HandleFunc("/delete/{id}", deleteItem).Methods("DELETE");
	router.HandleFunc("/completedItems", completedItems).Methods("GET");
	http.ListenAndServe(":8000", router)
}