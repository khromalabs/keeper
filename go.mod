module github.org/khromalabs/keeper

go 1.20

replace khromalabs/keeper/storage => ./storage

replace khromalabs/keeper/storage/sqlite => ./storage/sqlite

require (
	gopkg.in/yaml.v2 v2.4.0
	khromalabs/keeper/storage v0.0.0-00010101000000-000000000000
	khromalabs/keeper/storage/sqlite v0.0.0-00010101000000-000000000000
)
