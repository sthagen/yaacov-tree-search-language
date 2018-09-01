// Copyright 2018 Yaacov Zamir <kobi.zamir@gmail.com>
// and other contributors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package tsl describe and parse the Tree Search Language (TSL),
// TSL is a wonderful search language, With similar grammar to SQL's
// where part.
//
// TSL grammar examples:
//  "name = 'joe' or name = 'jane'"
//  "city in ('paris', 'rome', 'milan') and state != 'spane'"
//  "pages between 100 and 200 and author.name ~= 'Hilbert'"
//
// The package provide the ParseTSL method to convert TSL string into TSL tree.
//
// TSL tree can be used to generate SQL and MongoDB query filters. SquirrelWalk
// and BSONWalk methods can be used to create such filters.
//
// Squirrel walk code:  https://github.com/yaacov/tsl/blob/master/pkg/tsl/tsl_squirrel_walk.go
//
// Usage:
//   filter, err := tsl.SquirrelWalk(tree)
//
//   sql, args, err := sq.Select("name, city, state").
//        From("users").
//        Where(filter).
//        ToSql()
//
// BSON walk code: https://github.com/yaacov/tsl/blob/master/pkg/tsl/tsl_bson_walk.go
//
// Usage:
//   // Prepare a bson filter
//   filter, err = tsl.BSONWalk(tree)
//
//   // Run query
//   cur, err := collection.Find(ctx, bson.NewDocument(filter))
//
// squirrel: https://github.com/Masterminds/squirrel
//
// mongo-go-driver: https://github.com/mongodb/mongo-go-driver
//
package tsl
