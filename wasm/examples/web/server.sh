#!/bin/sh

echo "Starting server at http://localhost:8080"
goexec 'http.ListenAndServe(`:8080`, http.FileServer(http.Dir(`.`)))'