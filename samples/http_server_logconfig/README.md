# HTTP Server com Configuração de Logging

## Problema Resolvido

1. **WebSocket Hijacking**: Adicionado suporte para `http.Hijacker` e `http.Flusher` no `statusResponseWriter`
2. **Filtro de Logs**: Adicionada capacidade de ignorar logs para paths específicos (arquivos estáticos)

## Uso Básico (Padrão - Loga Tudo)

```go
package main

import (
    "github.com/faelp22/go-commons-libs/core/config"
    "github.com/faelp22/go-commons-libs/pkg/adapter/httpserver"
    "github.com/gorilla/mux"
)

func main() {
    conf := config.NewDefaultConf()
    router := mux.NewRouter()
    
    // Comportamento padrão: loga todas as requisições
    srv := httpserver.New(router, conf, nil)
    srv.ListenAndServe()
}
```

## Uso com Configuração de Logging

### Exemplo 1: Ignorar Arquivos Estáticos

```go
logConfig := &httpserver.LoggingMiddlewareConfig{
    Enabled: true,
    IgnorePaths: []string{
        "/assets/",     // Ignora /assets/app.js, /assets/style.css, etc.
        "/static/",     // Ignora /static/image.png, etc.
        "/favicon.ico", // Ignora favicon
    },
}

srv := httpserver.NewWithLogConfig(router, conf, nil, logConfig)
```

### Exemplo 2: Desabilitar Logging Completamente

```go
logConfig := &httpserver.LoggingMiddlewareConfig{
    Enabled: false, // Nenhum log será gerado
}

srv := httpserver.NewWithLogConfig(router, conf, nil, logConfig)
```

### Exemplo 3: Ignorar Múltiplos Padrões

```go
logConfig := &httpserver.LoggingMiddlewareConfig{
    Enabled: true,
    IgnorePaths: []string{
        "/assets/",
        "/static/",
        "/public/",
        "/images/",
        "/css/",
        "/js/",
        "/fonts/",
        "/favicon.ico",
        "/robots.txt",
        "/sitemap.xml",
    },
}

srv := httpserver.NewWithLogConfig(router, conf, nil, logConfig)
```

## Como Funciona

O `IgnorePaths` usa `strings.HasPrefix()` para verificar se o path da requisição começa com algum dos prefixos configurados:

- `/assets/` → Ignora: `/assets/app.js`, `/assets/css/style.css`, `/assets/deep/nested/file.js`
- `/favicon.ico` → Ignora apenas: `/favicon.ico`

## Resultado

### Antes (loga tudo):
```json
{"level":"info","Path":"/api/users","StatusCode":"200"}
{"level":"info","Path":"/assets/app.js","StatusCode":"200"}
{"level":"info","Path":"/assets/style.css","StatusCode":"200"}
{"level":"info","Path":"/static/logo.png","StatusCode":"200"}
```

### Depois (com IgnorePaths configurado):
```json
{"level":"info","Path":"/api/users","StatusCode":"200"}
```

## Compatibilidade

- `httpserver.New()` → Mantém comportamento padrão (loga tudo)
- `httpserver.NewWithLogConfig()` → Permite configuração customizada

## WebSocket Support

O `statusResponseWriter` agora implementa:
- `http.Hijacker` → Permite upgrade de WebSocket
- `http.Flusher` → Permite streaming de dados

```go
// Agora funciona sem erros!
upgrader := websocket.Upgrader{}
ws, err := upgrader.Upgrade(w, r, nil) // ✓ Funciona!
```
