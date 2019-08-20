package libs

import (
    "database/sql"
    "fmt"

    _ "github.com/mattn/go-sqlite3"
)

type DatabaseDAO struct {
    Db *sql.DB
}

func InitDb() DatabaseDAO {
    // Opens the database & if needed creates the tables
    database, _ := sql.Open("sqlite3", "./transactor.db")
    statement, _ := database.Prepare(
        "CREATE TABLE IF NOT EXISTS certificates (UUID string primary key, SIGNATURE string, SIGNER string)")
    _, _ = statement.Exec()

    statement, _ = database.Prepare(
        "CREATE TABLE IF NOT EXISTS secrets (UUID string primary key, recipientPrivateKey string)")
    _, _ = statement.Exec()

    dao := DatabaseDAO{
        Db: database,
    }
    return dao
}

func (dao *DatabaseDAO) AddCertificateEntry(uuid string, signature string, signer string) {
    // Adds a new certificate to the DB

    statement, _ := dao.Db.Prepare("INSERT INTO certificates VALUES (?, ?, ?)")
    _, _ = statement.Exec(uuid, signature, signer)
}

func (dao *DatabaseDAO) AddSecretEntry(uuid string, recipientPrivateKey string) {
    // Adds a new secret to the DB

    statement, _ := dao.Db.Prepare("INSERT INTO secrets VALUES (?, ?)")
    _, _ = statement.Exec(uuid, recipientPrivateKey)
}

func (dao *DatabaseDAO) RemoveCertificate(uuid string) {
    // Removes a certificate with given UUID from the DB

    statement, _ := dao.Db.Prepare("DELETE FROM certificates WHERE uuid = ?")
    _, err := statement.Exec(uuid)
    if err != nil {
        fmt.Println("Error deleting from DB", err)
    }
}

func (dao *DatabaseDAO) RemoveSecret(uuid string) {
    // Removes a secret with given UUID from the DB

    statement, _ := dao.Db.Prepare("DELETE FROM secrets WHERE uuid = ?")
    _, err := statement.Exec(uuid)
    if err != nil {
        fmt.Println("Error deleting from DB", err)
    }
}

func (dao *DatabaseDAO) GetSignatureAndSigner(uuid string) ([]byte, []byte, error) {
    // Returns the corresponding signature and signer to a certificate UUID
    statement, _ := dao.Db.Prepare("SELECT signature, signer FROM certificates WHERE uuid = ?")
    rows, _ := statement.Query(uuid)
    defer func() {
        _ = rows.Close()
    }()

    if rows.Next() {
        var signature string
        var signer string
        _ = rows.Scan(&signature, &signer)
        return []byte(signature), []byte(signer), nil
    }
    return nil, nil, fmt.Errorf("no such uuid in the database")
}

func (dao *DatabaseDAO) GetSecretDecryptingKey(uuid string) string {
    // Returns the recipient private key corresponding to a secrets transaction
    statement, _ := dao.Db.Prepare("SELECT recipientPrivateKey FROM secrets WHERE uuid = ?")
    rows, _ := statement.Query(uuid)
    defer func() {
        _ = rows.Close()
    }()

    if rows.Next() {
        var privKey string
        _ = rows.Scan(&privKey)
        return privKey
    }
    return ""
}

func (dao *DatabaseDAO) UpdateCertificateOptions() []string {
    // Gets and returns the []string of certificate UUIDS in the DB

    // Get count of uuids
    row := dao.Db.QueryRow("SELECT COUNT(DISTINCT uuid) FROM certificates")
    var count int
    _ = row.Scan(&count)
    result := make([]string, count+1) // Count+1 to have a place for the adding option
    // Fill the slice
    rows, _ := dao.Db.Query("SELECT DISTINCT uuid FROM certificates")
    var uuid string
    i := 0
    for rows.Next() {
        _ = rows.Scan(&uuid)
        result[i] = uuid
        i++
    }
    // Add option last
    result[i] = "Add certificate..."
    return result
}

func (dao *DatabaseDAO) UpdateSecretOptions() []string {
    // Gets and returns the []string of secrets UUIDS in the DB

    // Get count of secrets
    row := dao.Db.QueryRow("SELECT COUNT(DISTINCT uuid) FROM secrets")
    var count int
    _ = row.Scan(&count)
    result := make([]string, count+1) // Count+1 to have a place for the adding option
    // Fill the slice
    rows, _ := dao.Db.Query("SELECT DISTINCT uuid FROM secrets")
    var uuid string
    i := 0
    for rows.Next() {
        _ = rows.Scan(&uuid)
        result[i] = uuid
        i++
    }
    // Add option last
    result[i] = "Add secret..."
    return result

}
