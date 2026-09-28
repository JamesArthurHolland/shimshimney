module github.com/shimshimney/example/backend-4

go 1.20

require github.com/shimshimney/example/common v0.0.0

require github.com/shimshimney/pkg v0.0.0 // indirect

replace github.com/shimshimney/example/common => ../common

replace github.com/shimshimney/pkg => ../../../pkg
