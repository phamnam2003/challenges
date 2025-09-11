package main

import (
	"github.com/gocql/gocql"
	logger "github.com/phamnam2003/challenges/tech/scylladb/log"
	sccore "github.com/phamnam2003/challenges/tech/scylladb/sc_core"
	"go.uber.org/zap"
)

func main() {
	l := logger.CreateLogger("info")
	cluster := sccore.CreateCluster(gocql.Quorum, "catalog", "localhost:9042")
	session, err := gocql.NewSession(*cluster)
	if err != nil {
		l.Fatal("unable connect to scylladb", zap.Error(err))
	}
	l.Info("connected to scylladb")
	defer session.Close()

	sccore.SelectQuery(session, l)
	insertQuery(session, l)
	// insert in scylladb is upsert, so we can run it multiple times, it not make error when duplicate primary key, then update fields in last query
	insertQuery(session, l)
	sccore.SelectQuery(session, l)
	deleteQuery(session, l)
	sccore.SelectQuery(session, l)
}

func insertQuery(session *gocql.Session, logger *zap.Logger) {
	logger.Info("Inserting Mike")
	if err := session.Query("INSERT INTO mutant_data (first_name,last_name,address,picture_location) VALUES ('Mike','Tyson','1515 Main St', 'http://www.facebook.com/mtyson')").Exec(); err != nil {
		logger.Error("insert catalog.mutant_data", zap.Error(err))
	}
}

func deleteQuery(session *gocql.Session, logger *zap.Logger) {
	logger.Info("Deleting Mike")
	if err := session.Query("DELETE FROM mutant_data WHERE first_name = 'Mike' and last_name = 'Tyson'").Exec(); err != nil {
		logger.Error("delete catalog.mutant_data", zap.Error(err))
	}
}
