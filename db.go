package main

import (
	"database/sql"
	"log"

	"github.com/lib/pq"
)

const queryPerTxn = 100

func beginBatch(db *sql.DB) (*sql.Tx, *sql.Stmt, error) {

	txn, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	stmt, err := txn.Prepare(pq.CopyIn("hosts", "hostname"))
	if err != nil {
		log.Fatal(err)
	}

	return txn, stmt, nil
}

func endBatch(txn *sql.Tx, stmt *sql.Stmt) error {

	_, err := stmt.Exec()
	if err != nil {
		return err
	}

	err = stmt.Close()
	if err != nil {
		return err
	}

	err = txn.Commit()
	if err != nil {
		return err
	}

	return nil
}

func dbInsertWorker(db *sql.DB, ch chan interface{}) {

	var err error
	count := 1

	// setup initial tx
	txn, stmt, err := beginBatch(db)
	if err != nil {
		log.Fatalf("Error beginning postgres txn: %v", err)
	}

	for {

		// initiate new txn every 10k
		if count%queryPerTxn == 0 {

			err = endBatch(txn, stmt)
			if err != nil {
				log.Fatalf("Error ending postgres txn: %v", err)
			}

			txn, stmt, err = beginBatch(db)
			if err != nil {
				log.Fatalf("Error beginning postgres txn: %v", err)
			}
		}

		t, more := <-ch

		if !more {
			break
		}

		switch t := t.(type) {

		case Host:
			_, err = stmt.Exec(t.Ip)
		case DNSEntry:
			_, err = stmt.Exec(t.Name)
		}

		if err != nil {
			log.Printf("Error inserting: %v", err)
		}

		count++

		wg.Done()

	}

	// finish last batch
	err = endBatch(txn, stmt)
	if err != nil {
		log.Fatalf("Error ending postgres txn: %v", err)
	}

}
