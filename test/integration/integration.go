package integration

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"

	"github.com/gin-gonic/gin"
	"github.com/ooni/orchestra/orchestrate/orchestrate"
	"github.com/ooni/orchestra/registry/registry"
	"github.com/ory/dockertest"
)

func performRequest(r http.Handler, method, path string, body io.Reader) (*httptest.ResponseRecorder, error) {
	return performRequestWithJWT(r, method, path, "", body)
}

func performRequestWithJWT(r http.Handler, method, path, authToken string, body io.Reader) (*httptest.ResponseRecorder, error) {
	req, err := http.NewRequest(method, path, body)
	if err != nil {
		return nil, err
	}
	if authToken != "" {
		req.Header.Add("Authorization", `Bearer `+authToken)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w, nil
}

func performRequestJSON(r http.Handler, method, path string, reqJSON interface{}) (*httptest.ResponseRecorder, error) {
	return performRequestJSONWithJWT(r, method, path, "", reqJSON)
}

func performRequestJSONWithJWT(r http.Handler, method, path, authToken string, reqJSON interface{}) (*httptest.ResponseRecorder, error) {
	body, err := json.Marshal(reqJSON)
	if err != nil {
		return nil, err
	}
	return performRequestWithJWT(r, method, path, authToken, bytes.NewReader(body))
}

// Shared by all tests
var orchTest *OrchestraTest

// NewOrchestraTest populates the OrchestraTest struct with sane defaults
func NewOrchestraTest() *OrchestraTest {
	return &OrchestraTest{
		pgUser:     "orchestra",
		pgPassword: "changeme",
		pgDB:       "testingorchestra",
	}
}

// OrchestraTest contains the integration testing environment
type OrchestraTest struct {
	dockerPool *dockertest.Pool
	db         *sql.DB
	pgResource *dockertest.Resource
	pgUser     string
	pgPassword string
	pgPort     string
	pgDB       string
	pgURL      string
}

// GetPGURL returns the postgres db URL
func (o *OrchestraTest) GetPGURL(dbname string) string {
	return fmt.Sprintf("postgres://%s:%s@localhost:%s/%s?sslmode=disable", o.pgUser, o.pgPassword, o.pgPort, dbname)
}

// Setup should be run once per test-suite
func (o *OrchestraTest) Setup() error {
	var err error
	o.dockerPool, err = dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
		return err
	}

	o.pgResource, err = o.dockerPool.Run("postgres", "9.6",
		[]string{
			"POSTGRES_USER=" + o.pgUser,
			"POSTGRES_PASSWORD=" + o.pgPassword,
			"POSTGRES_DB=" + o.pgDB,
		})
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
		return err
	}

	if err = o.dockerPool.Retry(func() error {
		var err error
		o.pgPort = o.pgResource.GetPort("5432/tcp")
		o.pgURL = o.GetPGURL(o.pgDB)
		o.db, err = sql.Open("postgres", o.pgURL)
		if err != nil {
			return err
		}
		return o.db.Ping()
	}); err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
		return err
	}
	return nil
}

// CleanDB will drop all the tables created by the orchstra user
func (o *OrchestraTest) CleanDB() error {
	if _, err := o.db.Exec("DROP SCHEMA public CASCADE;CREATE SCHEMA public;"); err != nil {
		return err
	}
	return nil
}

// Teardown should be run at the end of a test suite
func (o *OrchestraTest) Teardown() error {
	return o.dockerPool.Purge(o.pgResource)
}

// NewOrchestrateRouter creates a router object to use for testing
func NewOrchestrateRouter(dbURL string) (*gin.Engine, error) {
	router := orchestrate.SetupRouter(dbURL)
	if router == nil {
		return nil, errors.New("failed to start orchestrate server")
	}
	return router, nil
}

// NewRegistryRouter creates a router object to use for testing
func NewRegistryRouter(dbURL string) (*gin.Engine, error) {
	fmt.Println(dbURL)
	router := registry.SetupRouter(dbURL)
	if router == nil {
		return nil, errors.New("failed to start registry server")
	}
	return router, nil
}
