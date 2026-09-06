package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newTestServer() http.Handler {
	return (&Server{
		store:    NewTaskStore(),
		template: templateForTest(),
	}).routes()
}

func TestHealth(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	res := httptest.NewRecorder()

	newTestServer().ServeHTTP(res, req)

	if res.Code != http.StatusOK ||
		!strings.Contains(res.Body.String(), `"status":"ok"`) {
		t.Fatalf(
			"resposta de health inesperada: %d %s",
			res.Code,
			res.Body.String(),
		)
	}
}

func TestCreateTaskRejectsEmptyTitle(t *testing.T) {
	req := httptest.NewRequest(
		http.MethodPost,
		"/api/tasks",
		strings.NewReader(`{"title":"  "}`),
	)

	req.Header.Set("Content-Type", "application/json")

	res := httptest.NewRecorder()
	newTestServer().ServeHTTP(res, req)

	if res.Code != http.StatusBadRequest {
		t.Fatalf("esperava 400, recebeu %d", res.Code)
	}
}

func TestCriarTarefa(t *testing.T) {
	req := httptest.NewRequest(
		http.MethodPost,
		"/api/tasks",
		strings.NewReader(`{"title":"Estudar CI e CD"}`),
	)

	req.Header.Set("Content-Type", "application/json")

	res := httptest.NewRecorder()
	newTestServer().ServeHTTP(res, req)

	if res.Code != http.StatusCreated {
		t.Fatalf("esperava status 201, recebeu %d", res.Code)
	}	
}

func TestListarTarefas(t *testing.T) {
	requisicao := httptest.NewRequest(
		http.MethodGet,
		"/api/tasks",
		nil,
	)

	resposta := httptest.NewRecorder()
	newTestServer().ServeHTTP(resposta, requisicao)

	if resposta.Code != http.StatusOK {
		t.Fatalf("esperava status 200, recebeu %d", resposta.Code)
	}

	if !strings.Contains(resposta.Body.String(), "Conhecer a aplicação") {
		t.Fatalf("a tarefa inicial não foi encontrada")
	}
}