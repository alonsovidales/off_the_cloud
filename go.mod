module otc

go 1.19

replace otc/api => ./api

replace otc/controllers => ./controllers

replace otc/models => ./models

require github.com/go-sql-driver/mysql v1.6.0 // indirect
