module github.com/DoMinhHHung/beebox-dev/services/beebox-project

go 1.26.5

require (
	github.com/DoMinhHHung/beebox-dev/modules/beebox-auth v0.0.0-00010101000000-000000000000
	github.com/jackc/pgx/v5 v5.11.0
)

replace github.com/DoMinhHHung/beebox-dev/modules/beebox-auth => ../../modules/beebox-auth

require (
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	golang.org/x/sync v0.17.0 // indirect
	golang.org/x/text v0.29.0 // indirect
)