package main

import (
	"log"

	"github.com/gocql/gocql"
	sccore "github.com/phamnam2003/challenges/tech/scylladb/sc_core"
	"github.com/scylladb/gocqlx/v2"
	"github.com/scylladb/gocqlx/v2/table"
)

type Mutant struct {
	FirstName       string `db:"first_name"`
	LastName        string `db:"last_name"`
	Address         string `db:"address"`
	PictureLocation string `db:"picture_location"`
}

func main() {
	cluster := sccore.CreateCluster(gocql.Quorum, "catalog", "localhost:9042")

	session, err := gocqlx.WrapSession(cluster.CreateSession())
	if err != nil {
		log.Fatal("unable connect to scylladb", err)
	}
	defer session.Close()

	log.Println("connected to scylladb")

	// create table metadata
	mutantMetadata := table.Metadata{
		Name:    "mutant_data",
		Columns: []string{"first_name", "last_name", "address", "picture_location"},
		PartKey: []string{"first_name", "last_name"},
		SortKey: []string{},
	}
	mutantTable := table.New(mutantMetadata)

	m := Mutant{
		FirstName:       "Pham",
		LastName:        "Nam",
		Address:         "ThaiBinh, Vietnam",
		PictureLocation: "https://example.com/pic.jpg",
	}

	// insert row into table
	q := session.Query(mutantTable.Insert()).BindStruct(m)
	if err := q.ExecRelease(); err != nil {
		log.Fatal("unable to insert mutant data", err)
	}

	// get one row from table with filter call bind struct filter
	q = session.Query(mutantTable.Get()).BindStruct(m)
	if err := q.GetRelease(&m); err != nil {
		log.Printf("unable to get mutant data: %s", err)
	}
	log.Printf("get one mutant_data: %+v", m)

	q = session.Query(mutantTable.Delete()).BindStruct(m)
	if err := q.ExecRelease(); err != nil {
		log.Fatal("unable to delete mutant data", err)
	}

	q = session.Query(mutantTable.Get()).BindStruct(m)
	if err := q.GetRelease(&m); err != nil {
		log.Printf("error get mutant data: %v %v", err, gocql.ErrNotFound)
	}

	// insert record into table mutant_data serve update query
	q = session.Query(mutantTable.Insert()).BindStruct(m)
	if err := q.ExecRelease(); err != nil {
		log.Fatal("unable to insert mutant data", err)
	}

	mU := Mutant{
		FirstName: "Pham",
		LastName:  "Nam",
		Address:   "HungYen, Vietnam",
	}
	q = session.Query(mutantTable.Update("address")).BindStruct(mU)
	if err := q.ExecRelease(); err != nil {
		log.Fatalf("error update: %s", err)
	}
}
