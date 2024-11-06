module main

go 1.19

replace config => ./config

require config v0.0.0-00010101000000-000000000000

require github.com/lib/pq v1.10.9 // indirect
