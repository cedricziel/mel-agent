package credentials

// This file ensures that all credential definitions in this package
// are registered when the package is imported.
// 
// The individual credential files (baserow_jwt.go, baserow_token.go, api_key.go)
// each have init() functions that register their credential definitions.
// By importing this package, all those init() functions will be executed.