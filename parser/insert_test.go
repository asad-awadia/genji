package parser

import (
	"testing"

	"github.com/asdine/genji/query"
	"github.com/stretchr/testify/require"
)

func TestParserInsert(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		expected query.Statement
		fails    bool
	}{
		{"Documents", `INSERT INTO test VALUES {a: 1, "b": "foo", c: 'bar', d: 1 = 1, e: {f: "baz"}}`,
			query.InsertStmt{
				TableName: "test",
				Values: query.LiteralExprList{
					query.KVPairs{
						query.KVPair{K: "a", V: query.Int8Value(1)},
						query.KVPair{K: "b", V: query.StringValue("foo")},
						query.KVPair{K: "c", V: query.StringValue("bar")},
						query.KVPair{K: "d", V: query.Eq(query.Int8Value(1), query.Int8Value(1))},
						query.KVPair{K: "e", V: query.KVPairs{
							query.KVPair{K: "f", V: query.StringValue("baz")},
						}},
					},
				}}, false},
		{"Documents / Multiple", `INSERT INTO test VALUES {"a": 'a', b: -2.3}, {a: 1, d: true}`,
			query.InsertStmt{
				TableName: "test",
				Values: query.LiteralExprList{
					query.KVPairs{
						query.KVPair{K: "a", V: query.StringValue("a")},
						query.KVPair{K: "b", V: query.Float64Value(-2.3)},
					},
					query.KVPairs{query.KVPair{K: "a", V: query.Int8Value(1)}, query.KVPair{K: "d", V: query.BoolValue(true)}},
				},
			}, false},
		{"Documents / Positional Param", "INSERT INTO test VALUES ?, ?",
			query.InsertStmt{
				TableName: "test",
				Values:    query.LiteralExprList{query.PositionalParam(1), query.PositionalParam(2)},
			},
			false},
		{"Documents / Named Param", "INSERT INTO test VALUES $foo, $bar",
			query.InsertStmt{
				TableName: "test",
				Values:    query.LiteralExprList{query.NamedParam("foo"), query.NamedParam("bar")},
			},
			false},
		{"Values / With columns", "INSERT INTO test (a, b) VALUES ('c', 'd', 'e')",
			query.InsertStmt{
				TableName:  "test",
				FieldNames: []string{"a", "b"},
				Values: query.LiteralExprList{
					query.LiteralExprList{query.StringValue("c"), query.StringValue("d"), query.StringValue("e")},
				},
			}, false},
		{"Values / Multiple", "INSERT INTO test (a, b) VALUES ('c', 'd'), ('e', 'f')",
			query.InsertStmt{
				TableName:  "test",
				FieldNames: []string{"a", "b"},
				Values: query.LiteralExprList{
					query.LiteralExprList{query.StringValue("c"), query.StringValue("d")},
					query.LiteralExprList{query.StringValue("e"), query.StringValue("f")},
				},
			}, false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			q, err := ParseQuery(test.s)
			if test.fails {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Len(t, q.Statements, 1)
			require.EqualValues(t, test.expected, q.Statements[0])
		})
	}
}
