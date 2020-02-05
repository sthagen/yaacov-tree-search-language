#
# Copyright 2018 Yaacov Zamir <kobi.zamir@gmail.com>
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#   http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

# Major version directory
major:=v4

# Details of the version of 'antlr' that will be downloaded to process the
# grammar and generate the parser:
antlr_url:=https://www.antlr.org/download/antlr-4.7.1-complete.jar
antlr_sum:=f41dce7441d523baf9769cb7756a00f27a4b67e55aacab44525541f62d7f6688

tsl_parser_src := $(wildcard ./$(major)/cmd/tsl_parser/*.go)
tsl_sqlite_src := $(wildcard ./$(major)/cmd/tsl_sqlite/*.go)
tsl_gorm_src := $(wildcard ./$(major)/cmd/tsl_gorm/*.go)
tsl_mongo_src := $(wildcard ./$(major)/cmd/tsl_mongo/*.go)
tsl_graphql_src := $(wildcard ./$(major)/cmd/tsl_graphql/*.go)
tsl_mem_src := $(wildcard ./$(major)/cmd/tsl_mem/*.go)

all: fmt tsl_parser tsl_sqlite tsl_gorm tsl_mongo tsl_graphql tsl_mem

tsl_parser: $(tsl_parser_src)
	cd ./$(major) && go build ./cmd/tsl_parser

tsl_sqlite: $(tsl_sqlite_src)
	cd ./$(major) && go build ./cmd/tsl_sqlite

tsl_gorm: $(tsl_gorm_src)
	cd ./$(major) && go build ./cmd/tsl_gorm

tsl_mongo: $(tsl_mongo_src)
	cd ./$(major) && go build ./cmd/tsl_mongo

tsl_graphql: $(tsl_graphql_src)
	cd ./$(major) && go build ./cmd/tsl_graphql

tsl_mem: $(tsl_mem_src)
	cd ./$(major) && go build ./cmd/tsl_mem

.PHONY: lint
lint:
	cd ./$(major) && golangci-lint \
		run \
		--skip-dirs=/pkg/parser \
		--no-config \
		--issues-exit-code=1 \
		--deadline=15m \
		--disable-all \
		--enable=deadcode \
		--enable=gas \
		--enable=goconst \
		--enable=gocyclo \
		--enable=gofmt \
		--enable=golint \
		--enable=ineffassign \
		--enable=interfacer \
		--enable=lll \
		--enable=maligned \
		--enable=megacheck \
		--enable=misspell \
		--enable=structcheck \
		--enable=unconvert \
		--enable=varcheck \
		$(NULL)

.PHONY: fmt
fmt:
	cd ./$(major) && gofmt -s -l -w ./pkg/ ./cmd/

.PHONY: clean
clean:
	rm ./$(major)/tsl_parser
	rm ./$(major)/tsl_gorm
	rm ./$(major)/tsl_mongo
	rm ./$(major)/tsl_sqlite
	rm ./$(major)/tsl_graphql
	rm ./$(major)/tsl_mem
	rm antlr

.PHONY: test
test:
	cd ./$(major) && ginkgo -r ./cmd ./pkg

.PHONY: generate
generate: antlr
	java -jar antlr -Dlanguage=Go -o pkg/parser TSL.g4

antlr:
	wget --progress=dot:giga --output-document="$@" "$(antlr_url)"
	echo "$(antlr_sum) $@" | sha256sum --check
