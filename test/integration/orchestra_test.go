package integration

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"testing"

	"github.com/ooni/orchestra/common/middleware"
	registry_handler "github.com/ooni/orchestra/registry/registry/handler"
)

const adminUsername = "test_admin"
const testingPassword = "testing"

func mapFromJSON(data []byte) map[string]interface{} {
	var result interface{}
	json.Unmarshal(data, &result)
	return result.(map[string]interface{})
}

func registerClient(r http.Handler, cd registry_handler.ClientData) (string, error) {

	w, err := performRequestJSON(r, "POST", "/api/v1/register", cd)
	if err != nil {
		return "", err
	}

	result := mapFromJSON(w.Body.Bytes())
	return result["client_id"].(string), nil
}

func updateClient(r http.Handler, clientID, authToken string, cd registry_handler.ClientData) (string, error) {
	w, err := performRequestJSONWithJWT(r, "PUT", "/api/v1/update/"+clientID, authToken, cd)
	if err != nil {
		return "", err
	}

	result := mapFromJSON(w.Body.Bytes())
	return result["status"].(string), nil
}

func login(r http.Handler, username, password string) (string, error) {
	reqJSON := middleware.Login{
		Username: username,
		Password: password,
	}
	w, err := performRequestJSON(r, "POST", "/api/v1/login", reqJSON)
	if err != nil {
		return "", err
	}

	result := mapFromJSON(w.Body.Bytes())
	return result["token"].(string), nil
}

func TestRegistryUpdate(t *testing.T) {
	err := orchTest.CleanDB()
	if err != nil {
		t.Fatal(err)
	}

	r, err := NewRegistryRouter(orchTest.pgURL)
	if err != nil {
		t.Fatal(err)
	}

	cd := registry_handler.ClientData{
		ProbeCC:            "IT",
		ProbeASN:           "AS1234",
		Platform:           "android",
		SoftwareName:       "ooni-testing",
		SoftwareVersion:    "0.0.1",
		SupportedTests:     []string{"web_connectivity"},
		NetworkType:        "wifi",
		AvailableBandwidth: "100",
		Language:           "en",
		Token:              "XXXX-TESTING",
		Password:           testingPassword,
	}
	clientID, err := registerClient(r, cd)
	if err != nil {
		t.Error(err)
	}

	fmt.Printf("Registered: %s\n", clientID)
	token, err := login(r, clientID, testingPassword)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("Login token: %s", token)

	cd.ProbeCC = "GR"
	status, err := updateClient(r, clientID, token, cd)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("Update status: %s", status)
}

func TestEmptyToken(t *testing.T) {
	err := orchTest.CleanDB()
	if err != nil {
		t.Fatal(err)
	}

	r, err := NewRegistryRouter(orchTest.pgURL)
	if err != nil {
		t.Fatal(err)
	}

	cd := registry_handler.ClientData{
		ProbeCC:            "IT",
		ProbeASN:           "AS1234",
		Platform:           "android",
		SoftwareName:       "ooni-testing",
		SoftwareVersion:    "0.0.1",
		SupportedTests:     []string{"web_connectivity"},
		NetworkType:        "wifi",
		AvailableBandwidth: "100",
		Language:           "en",
		Token:              "",
		Password:           testingPassword,
	}
	clientID, err := registerClient(r, cd)
	if err != nil {
		t.Error(err)
	}

	fmt.Printf("Registered: %s\n", clientID)
	token, err := login(r, clientID, testingPassword)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("Login token: %s", token)

	cd.Token = "XXX-Some-Real-Token"
	status, err := updateClient(r, clientID, token, cd)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("Update status: %s", status)
}

func TestMain(m *testing.M) {
	orchTest = NewOrchestraTest()
	err := orchTest.Setup()
	if err != nil {
		log.Fatal(err)
	}

	exitCode := m.Run()

	err = orchTest.Teardown()
	if err != nil {
		log.Fatal(err)
	}

	os.Exit(exitCode)
}
