// +build dev

package server

import "net/http"

// Assets contains project assets.
var Assets http.FileSystem = http.Dir("../app/build/")
