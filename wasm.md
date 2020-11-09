create index.html

```bash
cp $(go env GOROOT)/misc/wasm/wasm_exec.js .
```

```bash
GOOS=js GOARCH=wasm go build -o lunar-defence.wasm github.com/jatekalkotok/lunar-defence
```

run in docker otherwise it can't find the wasm_exec.js via AJAX
