package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"
)

type UserXML struct {
	Id        int    `xml:"id"`
	FirstName string `xml:"first_name"`
	LastName  string `xml:"last_name"`
	Age       int    `xml:"age"`
	About     string `xml:"about"`
	Gender    string `xml:"gender"`
	IsActive  bool   `xml:"isActive" json:"-"`
}

type XML struct {
	Users []UserXML `xml:"row"`
}

const (
	rightToken  string = "Right"
	errorToken  string = "Error"
	headerToken string = "AccessToken"
)

func checkToken(r *http.Request) bool {
	token := r.Header.Get(headerToken)
	return token == rightToken
}

func findUser(id int) ([]User, error) {
	var result []User = []User{}
	for _, item := range getUsers() {
		if id == -1 || item.Id == id {
			result = append(result, item)
		}
	}
	return result, nil
}

func SearchServer(w http.ResponseWriter, r *http.Request) {
	if !checkToken(r) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	var id int = -1
	var err error
	var limit = 25
	query := r.URL.Query().Get("query")
	limitString := r.URL.Query().Get("limit")
	if limitString != "" {
		limit, err = strconv.Atoi(limitString)
		if err != nil {
			fmt.Fprintf(w, err.Error())
			return
		}
	}
	switch query {
	case "internal error":
		w.WriteHeader(http.StatusInternalServerError)
		return
	case "unknown error":
		panic(fmt.Errorf("ERROR"))

	case "bad request":
		w.WriteHeader(http.StatusBadRequest)
		result, _ := json.Marshal(SearchErrorResponse{
			Error: "bad request",
		})
		w.Write(result)
		return
	case "bad request json":
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("bad json"))
		return
	case "bad json":
		w.Write([]byte("bad json"))
		return
	case "bad order":
		w.WriteHeader(http.StatusBadRequest)
		result, _ := json.Marshal(SearchErrorResponse{
			Error: "ErrorBadOrderField",
		})
		w.Write(result)
		return
	}

	if query != "" {
		id, err = strconv.Atoi(strings.Split(query, "=")[1])
		if err != nil {
			fmt.Fprintf(w, err.Error())
			return
		}
	}
	users, err := findUser(id)
	if err != nil {
		fmt.Fprintf(w, err.Error())
		return
	}
	if len(users) > limit {
		users = users[:limit]
	}
	result, err := json.Marshal(users)
	if err != nil {
		fmt.Fprintf(w, err.Error())
		return
	}
	w.Write(result)
}

func userXMltoUser(usersXML []UserXML) []User {
	var result []User = make([]User, 0, 35)
	for _, userXMLItem := range usersXML {
		result = append(result, User{
			Id:     userXMLItem.Id,
			Age:    userXMLItem.Age,
			About:  userXMLItem.About,
			Name:   userXMLItem.FirstName + " " + userXMLItem.LastName,
			Gender: userXMLItem.Gender,
		})
	}
	return result
}

func getUsers() []User {
	data, err := ioutil.ReadFile("dataset.xml")
	if err != nil {
		panic("dataset doesn`t exists")
	}
	xmlData := new(XML)
	xml.Unmarshal(data, &xmlData)
	return userXMltoUser(xmlData.Users)
}

type testCase struct {
	token         string
	result        SearchResponse
	searchRequest SearchRequest
	err           error
}

var testCases []testCase = []testCase{
	{
		token: errorToken,
		err:   fmt.Errorf("Bad AccessToken"),
	},
	{
		token: rightToken,
		searchRequest: SearchRequest{
			Limit: 27,
		},
		result: SearchResponse{
			Users:    getUsers()[:25],
			NextPage: true,
		},
	},
	{
		token: rightToken,
		result: SearchResponse{
			Users: []User{{
				Id:     4,
				Age:    30,
				Gender: "male",
				Name:   "Owen Lynn",
				About:  "Elit anim elit eu et deserunt veniam laborum commodo irure nisi ut labore reprehenderit fugiat. Ipsum adipisicing labore ullamco occaecat ut. Ea deserunt ad dolor eiusmod aute non enim adipisicing sit ullamco est ullamco. Elit in proident pariatur elit ullamco quis. Exercitation amet nisi fugiat voluptate esse sit et consequat sit pariatur labore et.\n",
			}},
		},
		searchRequest: SearchRequest{
			Limit: 1,
			Query: "id=4",
		},
	},
	{
		searchRequest: SearchRequest{
			Limit: -1,
		},
		err: fmt.Errorf("limit must be > 0"),
	},
	{
		searchRequest: SearchRequest{
			Offset: -1,
		},
		err: fmt.Errorf("offset must be > 0"),
	},
	{
		token: rightToken,
		err:   fmt.Errorf("SearchServer fatal error"),
		searchRequest: SearchRequest{
			Limit: 1,
			Query: "internal error",
		},
	},
	{
		token: rightToken,
		searchRequest: SearchRequest{
			Query: "bad request",
		},
		err: fmt.Errorf("unknown bad request error: bad request"),
	},
	{
		token: rightToken,
		searchRequest: SearchRequest{
			Query: "bad request json",
		},
		err: fmt.Errorf("cant unpack error json: invalid character 'b' looking for beginning of value"),
	},
	{
		token: rightToken,
		searchRequest: SearchRequest{
			Query: "bad json",
		},
		err: fmt.Errorf("cant unpack result json: invalid character 'b' looking for beginning of value"),
	},
	{
		token: rightToken,
		searchRequest: SearchRequest{
			OrderField: "error field",
			Query:      "bad order",
		},
		err: fmt.Errorf("OrderFeld error field invalid"),
	},
}

func TestGet(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer ts.Close()
	for caseNum, testCase := range testCases {
		sc := &SearchClient{
			URL:         ts.URL,
			AccessToken: testCase.token,
		}
		result, err := sc.FindUsers(testCase.searchRequest)
		if testCase.err != nil {
			if testCase.err.Error() != err.Error() {
				fmt.Print(err.Error())
				t.Errorf("Error not equals for case num %d", caseNum)
			}
			continue
		}
		if result.NextPage != testCase.result.NextPage {
			t.Errorf("Not equal next page for case num %d", caseNum)
		}
		if result.Users == nil && testCase.result.Users == nil {
			continue
		}
		if result.Users == nil && testCase.result.Users != nil ||
			result.Users != nil && testCase.result.Users == nil {
			t.Errorf("Not equal users for case num %d", caseNum)
		}
		if len(testCase.result.Users) != len(result.Users) {
			t.Errorf("Users slices are not equal for case num %d", caseNum)
			return
		}
		for index, user := range testCase.result.Users {
			resultUser := result.Users[index]
			if user.Id != resultUser.Id {
				t.Errorf("User ids are not equal for user %d case num %d", index, caseNum)
				return
			}
			if user.About != resultUser.About {
				t.Errorf("User About %d not equal for case num %d", index, caseNum)
				return
			}
			if user.Age != resultUser.Age {
				t.Errorf("User Age %d not equal for case num %d", index, caseNum)
				return
			}
			if user.Gender != resultUser.Gender {
				t.Errorf("User Gender %d not equal for case num %d", index, caseNum)
				return
			}
			if user.Name != resultUser.Name {
				t.Errorf("User Name %d not equal for case num %d", index, caseNum)
				return
			}
		}

	}
}

func TestWithoutUrl(t *testing.T) {
	sc := &SearchClient{}
	_, err := sc.FindUsers(SearchRequest{})
	fmt.Println(err)
	if !strings.HasPrefix(err.Error(), "unknown error") {
		t.Errorf("Bad error")
	}
}

func TestTimoutError(t *testing.T) {
	ser := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(time.Second + 1)
			w.Write([]byte("asd"))
		}),
	)
	defer ser.Close()
	sc := &SearchClient{
		URL: ser.URL,
	}
	_, err := sc.FindUsers(SearchRequest{})
	if err.Error() != "timeout for limit=1&offset=0&order_by=0&order_field=&query=" {
		t.Errorf("bad timeout error")
	}
}
