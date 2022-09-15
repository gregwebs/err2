module github.com/gregwebs/err2

replace (
	github.com/lainio/err2 => ./
	github.com/lainio/internal/debug => ./internal/debug
	github.com/lainio/internal/handler => ./internal/handler
	github.com/lainio/try => ./try
)

go 1.18

require github.com/lainio/err2 v0.8.8
