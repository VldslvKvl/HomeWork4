package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"
)

type UsersXML struct {
	Users []UserXML `xml:"row"`
}

const (
	token = 1188
)

var allowedSortFields = map[string]struct{}{
	"Id":   {},
	"Name": {},
	"Age":  {},
}

type UserXML struct {
	ID        int    `xml:"id"`
	FirstName string `xml:"first_name"`
	LastName  string `xml:"last_name"`
	Age       int    `xml:"age"`
	About     string `xml:"about"`
	Gender    string `xml:"gender"`
}

func SearchServer(w http.ResponseWriter, r *http.Request) {
	if tok, _ := strconv.Atoi(r.Header.Get("AccessToken")); tok != token {
		http.Error(w, "Bad AccessToken", http.StatusUnauthorized)
		return
	}
	var selectedUser []User
	ServXML := &UsersXML{}
	file, err := os.ReadFile("dataset.xml")
	if err != nil {
		panic(err)
	}
	err = xml.Unmarshal(file, ServXML)
	if err != nil {
		fmt.Println("xml error:", err)
		panic(err)
	}
	_, err = strconv.Atoi(r.FormValue("limit"))
	if err != nil {
		http.Error(w, "SearchServer fatal error", http.StatusInternalServerError)
		return
	}
	offs, err := strconv.Atoi(r.FormValue("offset"))
	if err != nil {
		http.Error(w, "SearchServer fatal error", http.StatusInternalServerError)
		return
	}
	ordB, err := strconv.Atoi(r.FormValue("order_by"))
	if err != nil || OrderByAsc > ordB || OrderByDesc < ordB {
		http.Error(w, "OrderByIsOutOFBoundary", http.StatusBadRequest)
		return
	}
	ordF := r.FormValue("order_field")
	if _, ok := allowedSortFields[ordF]; !ok {
		er, _ := json.Marshal(SearchErrorResponse{"ErrorBadOrderField"})
		http.Error(w, string(er), http.StatusBadRequest)
		return
	}
	quer := r.FormValue("query")
	if quer == "" {
		er, _ := json.Marshal(SearchErrorResponse{"SearchServer fatal error"})
		http.Error(w, string(er), http.StatusBadRequest)
		return
	}
	for _, user := range ServXML.Users {
		if strings.Contains(user.FirstName+user.LastName, quer) || strings.Contains(user.About, quer) {
			selectedUser = append(selectedUser, User{
				Id:     user.ID,
				Name:   user.FirstName + user.LastName,
				Age:    user.Age,
				About:  user.About,
				Gender: user.Gender,
			})
		}
	}
	if ordB == OrderByAsc {
		sort.Slice(selectedUser, func(i, j int) bool {
			switch ordF {
			case "Age":
				return selectedUser[i].Age < selectedUser[j].Age
			case "Id":
				return selectedUser[i].Id < selectedUser[j].Id
			case "Name":
				return selectedUser[i].Name < selectedUser[j].Name
			}
			return false
		})
	} else if ordB == OrderByDesc {
		sort.Slice(selectedUser, func(i, j int) bool {
			switch ordF {
			case "Age":
				return selectedUser[i].Age > selectedUser[j].Age
			case "Id":
				return selectedUser[i].Id > selectedUser[j].Id
			case "Name":
				return selectedUser[i].Name > selectedUser[j].Name
			}
			return false
		})
	}
	if len(selectedUser) < 1 {
		err = fmt.Errorf("range is empty")
		return
	}
	if offs <= len(selectedUser) {
		selectedUser = selectedUser[offs:]
	} else {
		err = fmt.Errorf("limit and offset not in range boundary")
		return
	}
	OutFile, err := json.Marshal(selectedUser)
	if err != nil {
		panic(err)
	}
	_, err = w.Write(OutFile)
	if err != nil {
		return
	}
	if w.Header() == nil {
		w.WriteHeader(http.StatusOK)
	}
}

var validRespUsr = &User{
	Id:     27,
	Name:   "RebekahSutton",
	Age:    26,
	About:  "Aliqua exercitation ad nostrud et exercitation amet quis cupidatat esse nostrud proident.",
	Gender: "female",
}
var respCompare = func(r *SearchResponse) bool {
	if r.Users[0].Id == validRespUsr.Id &&
		r.Users[0].Age == validRespUsr.Age &&
		r.Users[0].Gender == validRespUsr.Gender &&
		r.Users[0].About == validRespUsr.About {
		return true
	}
	fmt.Println(r.Users[0])
	return false
}
var RequestList = []SearchRequest{
	{-1, 5, "officia", "Id", 0},
	{1, -2, "officia", "Name", 0},
	{0, 1, "Hilda", "Id", 0},
	{},
	{1, 1, "officia", "Gender", 0},
	{1, 1, "", "Id", 0},
	{26, 2, "officia", "Id", 0},
	{25, 0, "Panda", "Id", 0},
	{15, 0, "Sutton", "Id", 0},
}

func TestLimitOffset(T *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer ts.Close()
	var err error
	sClient := SearchClient{"1188", ts.URL}
	for idx, req := range RequestList {
		_, err = sClient.FindUsers(req)
		if err != nil {
			if idx == 0 {
				if err.Error() == "limit must be > 0" {
					fmt.Println("limit must be > 0")
					err = nil
				}
			} else {
				if err.Error() == "offset must be > 0" {
					fmt.Println("offset must be > 0")
					err = nil
				}
			}
		}
	}
	if err != nil {
		T.Fatal(err)
	}
}
func TestTimeOut(T *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		time.Sleep(time.Second)
		http.HandlerFunc(SearchServer).ServeHTTP(writer, request)
	}))
	defer ts.Close()
	var err error
	sClient := SearchClient{"1188", ts.URL}
	_, err = sClient.FindUsers(RequestList[2])
	if err != nil {
		fmt.Println(err)
		return
	}
	T.Fail()
}
func TestUknownError(T *testing.T) {
	httptest.NewServer(http.HandlerFunc(SearchServer))
	var err error
	sClient := SearchClient{"1188", "URL"}
	_, err = sClient.FindUsers(RequestList[3])
	if err != nil {
		fmt.Println(err)
		return
	}
	T.Fail()
}
func TestBodyClose(T *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer ts.Close()
	var err error
	sClient := SearchClient{"1188", ts.URL}
	res, err := sClient.FindUsers(RequestList[8])
	if err != nil {
		T.Fatal(err)
	}

	if res == nil {
		T.Fatal("expected response, got nil")
	}

	if len(res.Users) == 0 {
		T.Fatal("expected users, got empty list")
	}
	if !respCompare(res) {
		T.Fail()
	}
}
func TestStatusUnauthorized(T *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer ts.Close()
	var err error
	sClient := SearchClient{"1487", ts.URL}
	r, err := sClient.FindUsers(RequestList[3])
	if err == nil {
		T.Fatal(err)
	}
	if r != nil {
		T.Fatal("expected nil, got not nil")
	}
}
func TestStatusInternalErr(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()
	sClient := SearchClient{"1188", ts.URL}
	resp, err := sClient.FindUsers(RequestList[3])
	if err == nil {
		t.Fatal("expected SearchServer fatal error, got nil")
	}
	if err.Error() != "SearchServer fatal error" {
		t.Fatalf("expected SearchServer fatal error, got %v", err)
	}
	if resp != nil {
		t.Fatalf("expected nil response, got %#v", resp)
	}
}
func TestBadRequestErrBadOrdeer(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer ts.Close()
	sClient := SearchClient{"1188", ts.URL}
	resp, err := sClient.FindUsers(RequestList[4])
	if err == nil {
		t.Fatal("expected SearchServer fatal error, got nil")
	}
	if !strings.Contains(err.Error(), "OrderFeld") {
		t.Fatalf("OrderFeld expected, got %v", err)
	}
	if resp != nil {
		t.Fatalf("expected nil response, got %#v", resp)
	}
}
func TestParseErr(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer ts.Close()
	sClient := SearchClient{"1188", ts.URL}
	resp, err := sClient.FindUsers(RequestList[3])
	if err == nil {
		t.Fatal("expected cant unpack error json error, got nil")
	}
	if !strings.Contains(err.Error(), "cant unpack error json:") {
		t.Fatalf("cant unpack error json: expected, got %v", err)
	}
	if resp != nil {
		t.Fatalf("expected nil response, got %#v", resp)
	}
}
func TestValidRequest(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer ts.Close()
	sClient := SearchClient{"1188", ts.URL}
	resp, err := sClient.FindUsers(RequestList[8])
	if err != nil {
		fmt.Println(err)
		t.Fatal("expected nil, got error")
	}
	if resp == nil {
		t.Fatalf("expected  response, got nil")
	}
	if !respCompare(resp) {
		t.Fail()
	}
}
func TestSerchRequest(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer ts.Close()
	sClient := SearchClient{"1188", ts.URL}
	resp, err := sClient.FindUsers(RequestList[7])
	if err == nil {
		t.Fatal("expected cant unpack error json error, got nil")
	}
	if !strings.Contains(err.Error(), "cant unpack result json:") {
		t.Fatalf("cant unpack result json: expected, got %v", err)
	}
	if resp != nil {
		t.Fatalf("expected nil response, got %#v", resp)
	}
}

// код писать тут
