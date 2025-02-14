module highgoPG

go 1.22

toolchain go1.23.6

require (
	github.com/lib/pq v1.10.9
	gorm.io/gorm v1.21.10
)

replace github.com/lib/pq v1.10.9 => github.com/ean0728/highgopq v1.0.0

require (
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
)
