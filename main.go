package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Task struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Done      bool      `json:"done"`
	CreatedAt time.Time `json:"created_at"`
}

type TaskStore struct {
	mu     sync.RWMutex
	nextID int
	tasks  []Task
}

func NewTaskStore() *TaskStore {
	now := time.Now()
	return &TaskStore{
		nextID: 3,
		tasks: []Task{
			{ID: 1, Title: "Conhecer a aplicação", Done: true, CreatedAt: now.Add(-2 * time.Hour)},
			{ID: 2, Title: "Executar o projeto com Docker", CreatedAt: now.Add(-time.Hour)},
		},
	}
}

func (s *TaskStore) list() []Task {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]Task, len(s.tasks))
	copy(result, s.tasks)
	return result
}

func (s *TaskStore) add(title string) Task {
	s.mu.Lock()
	defer s.mu.Unlock()
	task := Task{ID: s.nextID, Title: title, CreatedAt: time.Now()}
	s.nextID++
	s.tasks = append(s.tasks, task)
	return task
}

func (s *TaskStore) toggle(id int) (Task, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.tasks {
		if s.tasks[i].ID == id {
			s.tasks[i].Done = !s.tasks[i].Done
			return s.tasks[i], true
		}
	}
	return Task{}, false
}

func (s *TaskStore) remove(id int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.tasks {
		if s.tasks[i].ID == id {
			s.tasks = append(s.tasks[:i], s.tasks[i+1:]...)
			return true
		}
	}
	return false
}

type Server struct {
	store    *TaskStore
	template *template.Template
}

func (s *Server) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", s.home)
	mux.HandleFunc("GET /health", s.health)
	mux.HandleFunc("GET /api/tasks", s.listTasks)
	mux.HandleFunc("POST /api/tasks", s.createTask)
	mux.HandleFunc("PATCH /api/tasks/{id}", s.toggleTask)
	mux.HandleFunc("DELETE /api/tasks/{id}", s.deleteTask)
	return logging(mux)
}

func (s *Server) home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	if err := s.template.Execute(w, map[string]any{"Tasks": s.store.list()}); err != nil {
		log.Printf("rendering home page: %v", err)
	}
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "service": "lista-de-tarefas"})
}

func (s *Server) listTasks(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.store.list())
}

func (s *Server) createTask(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title string `json:"title"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil || strings.TrimSpace(input.Title) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "informe um título para a tarefa"})
		return
	}
	task := s.store.add(strings.TrimSpace(input.Title))
	writeJSON(w, http.StatusCreated, task)
}

func (s *Server) toggleTask(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "id inválido"})
		return
	}
	task, found := s.store.toggle(id)
	if !found {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "tarefa não encontrada"})
		return
	}
	writeJSON(w, http.StatusOK, task)
}

func (s *Server) deleteTask(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || !s.store.remove(id) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "tarefa não encontrada"})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s (%s)", r.Method, r.URL.Path, time.Since(started).Round(time.Microsecond))
	})
}

const page = `<!doctype html>
<html lang="pt-BR">
<head>
  <meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Minha lista de tarefas</title>
  <style>
    :root { color-scheme: light; font-family: Inter, ui-sans-serif, system-ui, sans-serif; color: #172033; background: #f5f7fb; }
    * { box-sizing: border-box; } body { margin: 0; } main { max-width: 720px; margin: 0 auto; padding: 56px 20px; }
    .eyebrow { color: #5468e7; font-size: .78rem; font-weight: 800; letter-spacing: .12em; text-transform: uppercase; }
    h1 { margin: 8px 0; font-size: clamp(2rem, 6vw, 3.2rem); letter-spacing: -.05em; } .intro { color: #64708a; margin-bottom: 32px; }
    .card { background: white; border: 1px solid #e4e8f0; border-radius: 18px; box-shadow: 0 12px 30px #1720330d; overflow: hidden; }
    form { display: flex; gap: 10px; padding: 18px; border-bottom: 1px solid #edf0f5; } input { flex: 1; padding: 13px 15px; border: 1px solid #d9dfeb; border-radius: 10px; font-size: 1rem; outline-color: #5468e7; }
    button { border: 0; border-radius: 10px; background: #5468e7; color: white; padding: 0 18px; font-weight: 700; cursor: pointer; } button:hover { background: #4355ca; }
    ul { list-style: none; padding: 0 18px; margin: 0; } li { display: flex; align-items: center; gap: 12px; padding: 17px 0; border-bottom: 1px solid #f0f2f6; } li:last-child { border: 0; }
    .check { width: 22px; height: 22px; padding: 0; border: 2px solid #bcc5d7; border-radius: 50%; background: white; flex: 0 0 auto; } .check.done { border-color: #28a579; background: #28a579; } .title { flex: 1; } .done + .title { color: #9aa3b5; text-decoration: line-through; }
    .remove { background: transparent; color: #9aa3b5; padding: 5px; font-size: 1.1rem; } .remove:hover { color: #d4475b; background: transparent; } .empty { color: #8490a4; text-align: center; padding: 32px 0; }
    footer { color: #8993a7; font-size: .85rem; margin-top: 18px; }
  </style>
</head>
<body><main><div class="eyebrow">Projeto Docker · Semana 4</div><h1>Minha lista de tarefas</h1><p class="intro">Uma rotina simples para organizar o que precisa ser feito.</p>
<section class="card"><form id="task-form"><input id="title" autocomplete="off" placeholder="O que você precisa fazer?" required><button>Adicionar</button></form><ul id="tasks">{{range .Tasks}}<li data-id="{{.ID}}"><button class="check {{if .Done}}done{{end}}" onclick="toggle({{.ID}})" aria-label="Concluir tarefa"></button><span class="title">{{.Title}}</span><button class="remove" onclick="removeTask({{.ID}})" aria-label="Remover tarefa">×</button></li>{{else}}<li class="empty">Nenhuma tarefa por aqui.</li>{{end}}</ul></section><footer>Executando com Go e pronto para um container Docker.</footer></main>
<script>
const form = document.querySelector('#task-form'), input = document.querySelector('#title');
form.addEventListener('submit', async (event) => { event.preventDefault(); const title = input.value.trim(); if (!title) return; await fetch('/api/tasks', {method:'POST', headers:{'Content-Type':'application/json'}, body:JSON.stringify({title})}); input.value = ''; location.reload(); });
async function toggle(id) { await fetch('/api/tasks/' + id, {method:'PATCH'}); location.reload(); }
async function removeTask(id) { await fetch('/api/tasks/' + id, {method:'DELETE'}); location.reload(); }
</script></body></html>`

func main() {
	port := "8080"
	server := &Server{store: NewTaskStore(), template: template.Must(template.New("home").Parse(page))}
	log.Printf("servidor iniciado em http://localhost:%s", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), server.routes()); err != nil {
		log.Fatal(err)
	}
}
