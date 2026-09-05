# Lista de tarefas

Aplicação web simples feita em Go para a atividade Somativa 1 de DevOps. Ela permite criar, concluir e remover tarefas, com uma API JSON por trás da interface.

## Executar localmente

```bash
go run .
```

Acesse http://localhost:8080. Para rodar os testes:

```bash
go test ./...
```

## Executar com Docker

```bash
docker build -t lista-tarefas:local .
docker run --rm -p 8080:8080 lista-tarefas:local
```

Depois, acesse http://localhost:8080. O endpoint `/health` retorna o estado do serviço.

> Os dados ficam em memória e são reiniciados quando o container é encerrado. Isso é intencional para manter o exemplo pequeno e fácil de avaliar.
