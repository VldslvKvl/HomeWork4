package main

/*
import (
	"encoding/xml"
	"fmt"
	"net/http"
	"os"
)

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"os"
)

type ServerSerch struct {
	first_name string
	last_name  string
	about      string
	id         int
	age        int
	order_by   int
	limit      int
	offset     int
}

func SearchServer(w http.ResponseWriter, r *http.Request) {
	ServXML := &ServerSerch{}
	file, err := os.ReadFile("dataset.xml")
	if err != nil {
		panic(err)
	}
	xml.Unmarshal(file, ServXML)
	fmt.Fprintf(w, ServXML.about)
	var uploadFormTmplf = []byte(`
<html>
	<body>
	<form action="/upload" method="post" enctype="multipart/form-data">
		Image: <input type="file" name="my_file">
		<input type="submit" value="Upload">
	</form>
	</body>
</html>
`)
	w.Write(uploadFormTmplf)

}
func main() {

	http.HandleFunc("/", SearchServer)
	http.ListenAndServe(":8080", nil)

}*/
