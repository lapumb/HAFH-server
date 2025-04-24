package database

import (
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Database represents a SQLite database connection.
type Database struct {
	db *sql.DB
}

// PeripheralType represents the type of a peripheral device.
type PeripheralType int

const (
	PeripheralTypeUnknown PeripheralType = iota
	PeripheralTypeSensor
	PeripheralTypeActuator
	PeripheralTypeController
)

// String returns the string representation of the PeripheralType.
func (pt PeripheralType) String() string {
	switch pt {
	case PeripheralTypeSensor:
		return "Sensor"
	case PeripheralTypeActuator:
		return "Actuator"
	case PeripheralTypeController:
		return "Controller"
	default:
		return "Unknown"
	}
}

// PeripheralTypeFromString converts a string to a PeripheralType.
func PeripheralTypeFromString(s string) PeripheralType {
	switch s {
	case "Sensor":
		return PeripheralTypeSensor
	case "Actuator":
		return PeripheralTypeActuator
	case "Controller":
		return PeripheralTypeController
	default:
		return PeripheralTypeUnknown
	}
}

// Peripheral represents a device in the system.
type Peripheral struct {
	SerialNumber string         `json:"serial_number"`
	Type         PeripheralType `json:"type"`
	Name         string         `json:"name"`
	CreatedAt    time.Time      `json:"created_at"`
}

// ToJson serializes the Peripheral to JSON.
func (p *Peripheral) ToJson() ([]byte, error) {
	return json.Marshal(p)
}

// PeripheralFromJson is a factory-like function that deserializes JSON data into a Peripheral.
func PeripheralFromJson(data []byte) (*Peripheral, error) {
	var p Peripheral
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, err
	}

	if p.SerialNumber == "" {
		return nil, errors.New("JSON key 'serial_number' is required")
	} else if p.Type == PeripheralTypeUnknown {
		return nil, errors.New("JSON key 'type' is required and must be valid")
	}

	return &p, nil
}

// String returns the string representation of the Peripheral.
func (p *Peripheral) String() string {
	json, err := p.ToJson()
	if err != nil {
		return "{}"
	}

	return string(json)
}

// Reading represents a reading from a peripheral.
type Reading struct {
	ID           int            `json:"id"`
	SerialNumber string         `json:"serial_number"`
	Timestamp    time.Time      `json:"timestamp"`
	Data         map[string]any `json:"data"`
}

// ToJson serializes the Reading to JSON.
func (r *Reading) ToJson() ([]byte, error) {
	return json.Marshal(r)
}

// ReadingFromJson is a factory-like function that deserializes JSON data into a Reading.
func ReadingFromJson(data []byte) (*Reading, error) {
	var r Reading
	if err := json.Unmarshal(data, &r); err != nil {
		return nil, err
	}

	if r.SerialNumber == "" {
		return nil, errors.New("JSON key 'serial_number' is required")
	} else if len(r.Data) == 0 {
		return nil, errors.New("JSON key 'data' is required")
	}

	return &r, nil
}

// String returns the string representation of the Reading.
func (r *Reading) String() string {
	json, err := r.ToJson()
	if err != nil {
		return "{}"
	}

	return string(json)
}

// New initializes a new Database instance with the given SQLite database path.
func New(path string) (*Database, error) {
	// Open the SQLite database
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	// Set the connection pool settings. This allows multiple actors to use the same
	// database connection in a thread-safe, concurrent manner.
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	// Enable foreign key support
	_, err = db.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		db.Close()
		return nil, errors.New("failed to enable foreign key support: " + err.Error())
	}

	// Initialize the database schema
	d := &Database{db: db}
	if err := d.initSchema(); err != nil {
		return nil, err
	}

	return d, nil
}

func (d *Database) initSchema() error {
	peripheralsTable := `
	CREATE TABLE IF NOT EXISTS peripherals (
		serial_number TEXT PRIMARY KEY,
		type INTEGER NOT NULL,
		name TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`

	readingsTable := `
	CREATE TABLE IF NOT EXISTS readings (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		serial_number TEXT NOT NULL,
		timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		data JSON NOT NULL,
		FOREIGN KEY(serial_number) REFERENCES peripherals(serial_number)
	);`

	if _, err := d.db.Exec(peripheralsTable); err != nil {
		return err
	}

	if _, err := d.db.Exec(readingsTable); err != nil {
		return err
	}

	return nil
}

// AddPeripheral adds a new peripheral to the database.
func (d *Database) AddPeripheral(p *Peripheral) error {
	_, err := d.db.Exec(
		`INSERT OR IGNORE INTO peripherals (serial_number, type) VALUES (?, ?)`,
		p.SerialNumber, p.Type,
	)

	return err
}

// UpdatePeripheral updates the name of an existing peripheral.
func (d *Database) UpdatePeripheral(p *Peripheral) error {
	_, err := d.db.Exec(
		`UPDATE peripherals SET name = ? WHERE serial_number = ?`,
		p.Name, p.SerialNumber,
	)

	return err
}

// InsertReading inserts a new reading for a given peripheral.
func (d *Database) InsertReading(r *Reading) error {
	jsonData, err := json.Marshal(r.Data)
	if err != nil {
		return err
	}

	_, err = d.db.Exec(
		`INSERT INTO readings (serial_number, data) VALUES (?, ?)`,
		r.SerialNumber, string(jsonData),
	)

	return err
}

// GetLastReadings retrieves the last `limit` readings for a given peripheral.
func (d *Database) GetLastReadings(serial string, limit uint32) ([]Reading, error) {
	rows, err := d.db.Query(
		`SELECT id, serial_number, timestamp, data 
		 FROM readings 
		 WHERE serial_number = ? 
		 ORDER BY timestamp DESC 
		 LIMIT ?`,
		serial, limit,
	)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var results []Reading
	for rows.Next() {
		var r Reading
		var rawData string
		if err := rows.Scan(&r.ID, &r.SerialNumber, &r.Timestamp, &rawData); err != nil {
			return nil, err
		}

		if err := json.Unmarshal([]byte(rawData), &r.Data); err != nil {
			return nil, err
		}

		results = append(results, r)
	}

	return results, nil
}

// GetAllPeripherals retrieves all peripherals from the database.
func (d *Database) GetAllPeripherals() ([]Peripheral, error) {
	rows, err := d.db.Query(`SELECT serial_number, type, name, created_at FROM peripherals`)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var peripherals []Peripheral
	for rows.Next() {
		var p Peripheral
		var name sql.NullString
		if err := rows.Scan(&p.SerialNumber, &p.Type, &name, &p.CreatedAt); err != nil {
			return nil, err
		}

		// The 'name' is optional, and may be null.
		if name.Valid {
			p.Name = name.String
		} else {
			p.Name = ""
		}

		peripherals = append(peripherals, p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return peripherals, nil
}

// GetPeripheralBySerial retrieves a peripheral by its serial number.
func (d *Database) GetPeripheralBySerial(serial string) (*Peripheral, error) {
	row := d.db.QueryRow(
		`SELECT serial_number, type, created_at FROM peripherals WHERE serial_number = ?`,
		serial,
	)

	var p Peripheral
	if err := row.Scan(&p.SerialNumber, &p.Type, &p.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &p, nil
}

// GetPeripheralByName retrieves a peripheral by its name.
func (d *Database) GetPeripheralByName(name string) (*Peripheral, error) {
	row := d.db.QueryRow(
		`SELECT serial_number, type, created_at FROM peripherals WHERE name = ?`,
		name,
	)

	var p Peripheral
	if err := row.Scan(&p.SerialNumber, &p.Type, &p.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &p, nil
}

// Close closes the database connection.
func (d *Database) Close() error {
	return d.db.Close()
}
