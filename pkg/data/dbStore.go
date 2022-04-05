package data

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v4"
	"github.com/spf13/viper"
)

type DB struct {
	conn                  *pgx.Conn
	documentsSchema       string
	wordToIntSchema       string
	wordIdsToDocIdsSchema string
}

type WordInt uint64
type DocumentId WordInt

type Doc struct {
	DocId      DocumentId
	WordInts   []WordInt
	InputDocId string
	UserId     string
	BusinessId string
	Stars      float32
	Useful     uint16
	Funny      uint16
	Cool       uint16
	Text       string
	Date       string
}

func DBConnect() (*DB, error) {

	conn, err := pgx.Connect(context.Background(), viper.GetString("output.connection"))
	if err != nil {
		fmt.Printf("Error connecting to postgres database using: %s\n",
			viper.GetString("output.location"))
		fmt.Printf("err: %v\n", err)
		return nil, err
	}

	documentsSchema := `(docid bigint,
			wordints bigint[],
			inputdocId varchar(25) unique,
			userid varchar(25),
			businessId varchar(25),
			stars real, 
			useful smallint,
			funny smallint,
			cool smallint,
			text text,
			date varchar(25))`

	wordToIntSchema := `(word text unique, int bigint)`
	wordIdsToDocIdsSchema := `(wordid bigint unique, docids bigint[])`

	db := DB{
		conn:                  conn,
		documentsSchema:       documentsSchema,
		wordToIntSchema:       wordToIntSchema,
		wordIdsToDocIdsSchema: wordIdsToDocIdsSchema,
	}

	return &db, nil
}

func (db *DB) createTable(tableName string) {

	checkIfExists := `select 'public.` + tableName + `'::regclass;`
	if _, err := db.conn.Exec(context.Background(), checkIfExists); err != nil {
		fmt.Printf("Table %s does not exist, so create it.\n", tableName)

		createString := `create table ` + tableName + ` ` + db.documentsSchema + `;`
		if _, err := db.conn.Exec(context.Background(), createString); err != nil {
			fmt.Printf("Failed to create the schema. err: %v\n", err)
			os.Exit(-1)
		}
	}
}

func (db *DB) CreateTable(tableName string, schema string) error {

	if db.conn == nil {
		fmt.Printf("Create db connection before creating schema\n")
		return fmt.Errorf("Create db connection before creating schema\n")
	}

	checkIfExists := `select 'public.` + tableName + `'::regclass;`
	if _, err := db.conn.Exec(context.Background(), checkIfExists); err != nil {
		fmt.Printf("Table %s does not exist, so create it.\n", tableName)

		createString := `create table ` + tableName + ` ` + schema + `;`
		fmt.Printf("createString: %s\n", createString)
		if _, err := db.conn.Exec(context.Background(), createString); err != nil {
			fmt.Printf("Failed to create the schema. err: %v\n", err)
			return err
		}
	}

	return nil
}

func (db *DB) CreateDocumentsTable() error {

	return db.CreateTable("documents", db.documentsSchema)
}

func (db *DB) StoreData(doc *Doc, tableName string, wordInts []WordInt) error {

	insertStatement := `insert into ` + tableName + ` ` + db.documentsSchema +
		` values ($1, $2, $3, $4, $5, 
		 $6, $7, $8, $9, $10, $11)
		 on conflict(inputdocId) do nothing;`

	if _, err := db.conn.Exec(context.Background(), insertStatement,
		doc.DocId, doc.WordInts, doc.InputDocId,
		doc.UserId, doc.BusinessId, doc.Stars, doc.Useful,
		doc.Funny, doc.Cool, doc.Text, doc.Date); err != nil {
		fmt.Printf("Store data failed. err: %v\n", err)
		return err
	}

	return nil
}

func (db *DB) StoreWordIntMappings(wordToIntTable string, wordToInt map[string]WordInt) error {

	createWordToIntString := `(word, int)`

	db.CreateTable(wordToIntTable, db.wordToIntSchema)

	wordToIntInsertStatement := `insert into ` + wordToIntTable + ` ` + createWordToIntString +
		`values ($1, $2)
		on conflict(word) do nothing;`
	for word, i := range wordToInt {
		if _, err := db.conn.Exec(context.Background(), wordToIntInsertStatement,
			word, i); err != nil {
			fmt.Printf("Store Int to Word mapping failed. err: %v\n", err)
			return err
		}
	}

	return nil
}

func (db *DB) StoreWordToDocMappings(wordIdsToDocIdsTable string,
	wordToDocs map[WordInt][]DocumentId) error {

	db.CreateTable(wordIdsToDocIdsTable, db.wordIdsToDocIdsSchema)

	// In this update statement, the excluded docids are the ones that were not
	// inserted in.
	upsertStatement := `
		insert into wordid_to_docids(wordid, docids) values($1, $2)
		on conflict(wordid) do
		update set docids=array(select distinct unnest(wordid_to_docids.docids || excluded.docids));
		`

	for k, v := range wordToDocs {

		if _, err := db.conn.Exec(context.Background(), upsertStatement,
			k, v); err != nil {
			fmt.Printf("Update failed. err: %v\n", err)
			return err
		}
	}

	return nil
}

func (db *DB) ReadData() *Doc {

	return nil
}

func (db *DB) DBDisconnect() error {

	if db.conn == nil {
		fmt.Printf("conn is nil\n")
		os.Exit(-1)
	}

	err := db.conn.Close(context.Background())
	if err != nil {
		fmt.Printf("Error closing DB connection: %v\n", err)
		return err
	}

	return nil
}
