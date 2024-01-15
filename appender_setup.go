package duckdb

import (
	"context"
	"database/sql"
	"github.com/stretchr/testify/require"
	"testing"
)

package duckdb

import (
"context"
"database/sql"
"github.com/stretchr/testify/require"
"testing"
)

const (
	testAppenderTableNested = `
CREATE TABLE test(
	id BIGINT,
	charList VARCHAR[],
	intList INT[],
	nestedListInt INT[][],
	tripleNestedListInt INT[][][],
	Base STRUCT(I INT, V VARCHAR),
	Wrapper STRUCT(Base STRUCT(I INT, V VARCHAR)),
	TopWrapper STRUCT(Wrapper STRUCT(Base STRUCT(I INT, V VARCHAR))),
	structList STRUCT(I INT, V VARCHAR)[],
	listStruct STRUCT(L INT[]),
	mix STRUCT(A STRUCT(V VARCHAR[]), B STRUCT(L INT[])[]),
	mixList STRUCT(A STRUCT(V VARCHAR[]), B STRUCT(L INT[])[])[]
)`
)

func createAppenderNestedTable(db *sql.DB) *sql.Result {
	res, err := db.Exec(testAppenderTableNested)
	checkIfSucceded(err)
	return &res
}

func checkIfSucceded(err error) {
	if err != nil {
		panic(err)
	}
}

type dataRowInterface struct {
	charList            []interface{}
	intList             []interface{}
	nestedListInt       []interface{}
	tripleNestedListInt []interface{}
	base                interface{}
	wrapper             interface{}
	topWrapper          interface{}
	structList          []interface{}
	listStruct          interface{}
	mix                 interface{}
	mixList             []interface{}
}

type dataRow struct {
	ID                  int
	charList            ListString
	intList             ListInt
	nestedListInt       NestedListInt
	tripleNestedListInt TripleNestedListInt
	base                Base
	wrapper             Wrapper
	topWrapper          TopWrapper
	structList          []Base
	listStruct          ListInt
	mix                 Mix
	mixList             []Mix
}

func (dR *dataRow) Convert(i dataRowInterface) {
	dR.charList.FillFromInterface(i.charList)
	dR.intList.FillInnerFromInterface(i.intList)
	dR.nestedListInt.FillFromInterface(i.nestedListInt)
	dR.tripleNestedListInt.FillFromInterface(i.tripleNestedListInt)
	dR.base.FillFromInterface(i.base)
	dR.wrapper.FillFromInterface(i.wrapper)
	dR.topWrapper.FillFromInterface(i.topWrapper)
	dR.structList = dR.base.ListFillFromInterface(i.structList)
	dR.listStruct = dR.listStruct.FillFromInterface(i.listStruct)
	dR.mix.FillFromInterface(i.mix)
	dR.mixList = dR.mix.ListFillFromInterface(i.mixList)
}

func setupAppenderNested() *sql.DB {

	c, err := NewConnector("", nil)
	checkIfSucceded(err)

	db := sql.OpenDB(c)
	createAppenderNestedTable(db, t)
	defer db.Close()

	randRow := func(i int) dataRow {
		dR := dataRow{ID: i}
		dR.charList.Fill()
		dR.intList.Fill()
		dR.nestedListInt.Fill()
		dR.tripleNestedListInt.Fill()
		dR.base.Fill(i)
		dR.wrapper.Fill(i)
		dR.topWrapper.Fill(i)
		dR.structList = dR.base.ListFill(10)
		dR.listStruct.Fill()
		dR.mix.Fill()
		dR.mixList = dR.mix.ListFill(10)
		return dR
	}
	rows := []dataRow{}
	for i := 0; i < 100; i++ {
		rows = append(rows, randRow(i))
	}

	conn, err := c.Connect(context.Background())
	require.NoError(t, err)
	defer conn.Close()

	appender, err := NewAppenderFromConn(conn, "", "test")
	require.NoError(t, err)
	defer appender.Close()

	for _, row := range rows {
		err := appender.AppendRow(
			row.ID,
			row.charList.L,
			row.intList.L,
			row.nestedListInt.L,
			row.tripleNestedListInt.L,
			row.base,
			row.wrapper,
			row.topWrapper,
			row.structList,
			row.listStruct,
			row.mix,
			row.mixList,
		)
		require.NoError(t, err)
	}
	err = appender.Flush()
	require.NoError(t, err)

	return db
}

func compareAppenderNested(db *sql.DB) {
	res, err := db.QueryContext(
		context.Background(), `
			SELECT * FROM test ORDER BY id
    `)
	require.NoError(t, err)
	defer res.Close()

	i := 0
	for res.Next() {
		r := dataRow{}
		interfaces := dataRowInterface{}
		err := res.Scan(
			&r.ID,
			&interfaces.charList,
			&interfaces.intList,
			&interfaces.nestedListInt,
			&interfaces.tripleNestedListInt,
			&interfaces.base,
			&interfaces.wrapper,
			&interfaces.topWrapper,
			&interfaces.structList,
			&interfaces.listStruct,
			&interfaces.mix,
			&interfaces.mixList,
		)
		require.NoError(t, err)
		r.Convert(interfaces)
		require.Equal(t, rows[i], r)
		i++
	}
	// Ensure that the number of fetched rows equals the number of inserted rows.
	require.Equal(t, i, 100)
}
