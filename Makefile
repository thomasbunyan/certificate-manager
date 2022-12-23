.PHONY: build
build:
	go build -o target/

.PHONY: package
package: build
	zip -9 -j ${ZIP_NAME} target/${SERVICE}
