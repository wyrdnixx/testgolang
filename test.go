package main

import (
	"database/sql"
	"encoding/xml"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	// installieren
	//ulewu@ubu:~/goLang/testgolang$ go get github.com/lib/pq
	//ulewu@ubu:~/goLang/testgolang$ go get github.com/tkanos/gonfig

	_ "github.com/lib/pq" // Postgres

	"github.com/tkanos/gonfig" //gonfig -> config aus json file lesen
	// SQL Library
)

type Configuration struct {
	SrvPort  string
	PGDBName string
	PGDBHost string
	PGDBUser string
	PGDBPass string
}

type Page struct {
	Title string
	Body  []byte
}

type Abteilungen struct {
	XMLName     xml.Name    `xml:"Abteilungen"`
	Abteilungen []Abteilung `xml:"Abteilung"`
}
type Abteilung struct {
	XMLName xml.Name `xml:"Abteilung"`
	Name    string   `xml:"Name"`
	Keys    []Key    `xml:"Key"`
}

type Key struct {
	XMLName xml.Name `xml:"Key"`
	KeyName string   `xml:",chardata"`
	KeyType string   `xml:"type,attr"`
}

// testfunktion um xml einzulesen

func processXmlFile(_file string) (*Abteilungen, error) {
	// Read XML File
	xmlFile, err := os.Open(_file)
	if err != nil {
		fmt.Println("ERRO: XML File konnte nicht gelesen werden: ", err.Error())
		return nil, err
	} else {
		var abt Abteilungen

		defer xmlFile.Close()
		byteValue, err := ioutil.ReadAll(xmlFile)

		if err != nil {
			fmt.Println("ERRO: XML to byte: ", err.Error())
		} else {

			err := xml.Unmarshal(byteValue, &abt)
			if err != nil {
				//				fmt.Println("ERROR: Error parsing XML File:  ", err.Error())
			}
			//			fmt.Println("INFO: found xml Abteilungen: " + strconv.Itoa(len(abt.Abteilungen)))
			for a := 0; a < len(abt.Abteilungen); a++ {

				//				fmt.Println("Abteilung Name: " + abt.Abteilungen[a].Name)
				//				fmt.Println("INFO: found xml Keys: " + strconv.Itoa(len(abt.Abteilungen[a].Keys)))
				for k := 0; k < len(abt.Abteilungen[a].Keys); k++ {
					//					fmt.Println("KeyName: " + abt.Abteilungen[a].Keys[k].KeyName)
					//					fmt.Println("KeyType: " + abt.Abteilungen[a].Keys[k].KeyType)

				}

			}

		}
		return &abt, nil
	}

}

func main() {
	fmt.Printf("Hello World\n")

	// Config aus file laden:
	configuration := Configuration{}
	err := gonfig.GetConf("./config.development.json", &configuration)
	if err != nil {
		fmt.Println("ERROR: Config konnte nicht geladen werden.")
	}

	fmt.Println("INFO: Server lsitening on Port: ", configuration.SrvPort)

	// XML File enlesen
	//processXmlFile("anf2.xml")

	// Webservice
	http.HandleFunc("/", handler)
	http.HandleFunc("/view/", viewHandler)
	http.HandleFunc("/edit/", editHandler)
	http.HandleFunc("/save/", saveHandler)

	http.HandleFunc("/send", sendHandler)

	// Test DB-Verbindung
	//	checkDB(configuration)

	log.Println("INFO: starting webservice ...")
	log.Fatal(http.ListenAndServe(configuration.SrvPort, nil))

}

// Test - auf die DB Verbinden
//https://astaxie.gitbooks.io/build-web-application-with-golang/en/05.4.html

// Create user in Postgres
// user create Script
//create user myuser with encrypted password 'mypass';
//grant all privileges on database database to myuser;

func checkDB(conf Configuration) {
	dbinfo := fmt.Sprintf("user=%s pass=%s dbname=%s sslmode=disable",
		conf.PGDBUser, conf.PGDBPass, conf.PGDBName)
	fmt.Printf("user=%s pass=%s dbname=%s sslmode=disable",
		conf.PGDBUser, conf.PGDBPass, conf.PGDBName)

	fmt.Println("INFO: Try DB Connect...")
	db, err := sql.Open("postgres", dbinfo)
	if err != nil {
		fmt.Println("ERROR: DB Connection: ", err.Error())
	}
	defer db.Close()

	rows, err := db.Query("select * from testtable;")
	if err != nil {
		fmt.Println("ERROR: DB Connection: ", err.Error())
	}
	for rows.Next() {
		var id int
		rows.Scan(&id)
		fmt.Println("%3v ")
	}

}

type myData struct {
	Owner string
	Name  string
}

type JSONDATA struct {
	key   string
	value string
}

func sendHandler(w http.ResponseWriter, r *http.Request) {
	//title := r.URL.Path[len("/save"):]
	//	body := r.FormValue("editbox")
	fmt.Printf("Executing SendHandler\n")

	////////////////////////////
	/// Dies war der erste test
	/*
		// key, value unten geht irgendwie nur, wenn vorher dieses FormValue gemacht wird.
		fmt.Println(r.FormValue("*"))

		for key, values := range r.Form { // range over map
			for _, value := range values { // range over []string
				fmt.Println(key, value)
			}
		}

		// encode body to json  t
		var jsd JSONDATA
		if r.Body == nil {
			http.Error(w, "Please send a request body", 400)
			return
		}
		err := json.NewDecoder(r.Body).Decode(&jsd)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		fmt.Println(jsd)
		//fmt.Println(jsd.value)

		log.Fatal(http.ListenAndServe(":8080", nil))
		///////////////
	*/

	http.Redirect(w, r, "/", http.StatusFound)
}

func saveHandler(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Path[len("/save"):]
	body := r.FormValue("editbox")

	fmt.Println("INFO: got from " + r.Method + " : " + body)

	p := &Page{Title: title, Body: []byte(body)}
	err := p.save()
	if err != nil {
		fmt.Println("ERROR: cant save file: " + err.Error())
	}
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

func (p *Page) save() error {
	filename := "./files/" + p.Title
	return ioutil.WriteFile(filename, p.Body, 0600)
}

func loadPage(title string) (*Page, error) {
	//filename := "./files/" + title + ".txt"
	filename := "./files/" + title
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println("Error: cant read file : ", filename)
		return nil, err
	} else {
		log.Println("INFO: Loaded file: ", filename)
		return &Page{Title: title, Body: body}, nil
	}

}

type FileList struct {
	Name      string
	FileNames []string
}

func handler(w http.ResponseWriter, r *http.Request) {

	fmt.Println("INFO: Main Page loading... ")

	_fileList := FileList{Name: "generated..."}

	files, err := ioutil.ReadDir("./files")

	if err != nil {
		log.Fatal(err)
	} else {
		for _, file := range files {
			fmt.Println("INFO: File found: " + file.Name())
			//		_body = append(_body, file.Name())
			_fileList.FileNames = append(_fileList.FileNames, file.Name())
		}
	}
	// t, _ := template.ParseFiles("main.html", "header.html")
	// t.Execute(w, _fileList)

	abt, _ := processXmlFile("anf2.xml")
	x, _ := template.ParseFiles("main.html", "header.html")
	x.Execute(w, abt)

}

func viewHandler(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Path[len("/view/"):]
	p, _ := loadPage(title)
	renderTemplate(w, "view", p)
}

func editHandler(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Path[len("/edit/"):]

	p, err := loadPage(title)
	if err != nil {
		p = &Page{Title: title}
	}
	renderTemplate(w, "edit", p)

}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {

	t, _ := template.ParseFiles(tmpl+".html", "header.html")
	t.Execute(w, p)
}
