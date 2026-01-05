package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

// DB wraps the SQLite database connection
type DB struct {
	conn *sql.DB
}

// New creates a new database connection and initializes tables
func New(dbPath string) (*DB, error) {
	conn, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Enable foreign keys
	if _, err := conn.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	// Enable WAL mode for better concurrency
	if _, err := conn.Exec("PRAGMA journal_mode = WAL"); err != nil {
		return nil, fmt.Errorf("failed to enable WAL mode: %w", err)
	}

	// Set busy timeout to 5 seconds to handle lock contention
	if _, err := conn.Exec("PRAGMA busy_timeout = 5000"); err != nil {
		return nil, fmt.Errorf("failed to set busy timeout: %w", err)
	}

	db := &DB{conn: conn}

	// Initialize tables
	if err := db.initTables(); err != nil {
		return nil, fmt.Errorf("failed to initialize tables: %w", err)
	}

	return db, nil
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.conn.Close()
}

// initTables creates the database schema
func (db *DB) initTables() error {
	schema := `
	CREATE TABLE IF NOT EXISTS trades (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		symbol TEXT NOT NULL,
		side TEXT NOT NULL CHECK(side IN ('BUY', 'SELL')),
		quantity REAL NOT NULL,
		price REAL NOT NULL,
		total REAL NOT NULL,
		strategy TEXT NOT NULL,
		indicator_values TEXT,
		signal_reason TEXT,
		paper_trade BOOLEAN NOT NULL DEFAULT 1,
		timestamp DATETIME NOT NULL,
		binance_order_id TEXT,
		profit_loss REAL,
		profit_loss_percent REAL,
		related_buy_id INTEGER,
		FOREIGN KEY (related_buy_id) REFERENCES trades(id)
	);

	CREATE TABLE IF NOT EXISTS positions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		symbol TEXT NOT NULL,
		quantity REAL NOT NULL,
		entry_price REAL NOT NULL,
		entry_time DATETIME NOT NULL,
		exit_price REAL,
		exit_time DATETIME,
		strategy TEXT NOT NULL,
		is_open BOOLEAN NOT NULL DEFAULT 1,
		profit_loss REAL,
		profit_loss_percent REAL,
		buy_trade_id INTEGER NOT NULL,
		sell_trade_id INTEGER,
		FOREIGN KEY (buy_trade_id) REFERENCES trades(id),
		FOREIGN KEY (sell_trade_id) REFERENCES trades(id)
	);

	CREATE INDEX IF NOT EXISTS idx_trades_timestamp ON trades(timestamp);
	CREATE INDEX IF NOT EXISTS idx_trades_symbol ON trades(symbol);
	CREATE INDEX IF NOT EXISTS idx_positions_symbol ON positions(symbol);
	CREATE INDEX IF NOT EXISTS idx_positions_is_open ON positions(is_open);
	`

	_, err := db.conn.Exec(schema)
	return err
}

// InsertTrade inserts a new trade into the database
func (db *DB) InsertTrade(trade *Trade) (int64, error) {
	query := `
		INSERT INTO trades (
			symbol, side, quantity, price, total, strategy,
			indicator_values, signal_reason, paper_trade, timestamp,
			binance_order_id, profit_loss, profit_loss_percent, related_buy_id
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := db.conn.Exec(
		query,
		trade.Symbol,
		trade.Side,
		trade.Quantity,
		trade.Price,
		trade.Total,
		trade.Strategy,
		trade.IndicatorValues,
		trade.SignalReason,
		trade.PaperTrade,
		trade.Timestamp,
		trade.BinanceOrderID,
		nullFloat64(trade.ProfitLoss),
		nullFloat64(trade.ProfitLossPercent),
		nullInt64(trade.RelatedBuyID),
	)

	if err != nil {
		return 0, fmt.Errorf("failed to insert trade: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert ID: %w", err)
	}

	return id, nil
}

// InsertTradesInTransaction inserts multiple trades in a single transaction
// This is much faster and avoids database lock issues when inserting bulk data
func (db *DB) InsertTradesInTransaction(trades []*Trade) error {
	tx, err := db.conn.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback() // Rollback if we don't commit

	query := `
		INSERT INTO trades (
			symbol, side, quantity, price, total, strategy,
			indicator_values, signal_reason, paper_trade, timestamp,
			binance_order_id, profit_loss, profit_loss_percent, related_buy_id
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	stmt, err := tx.Prepare(query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, trade := range trades {
		_, err := stmt.Exec(
			trade.Symbol,
			trade.Side,
			trade.Quantity,
			trade.Price,
			trade.Total,
			trade.Strategy,
			trade.IndicatorValues,
			trade.SignalReason,
			trade.PaperTrade,
			trade.Timestamp,
			trade.BinanceOrderID,
			nullFloat64(trade.ProfitLoss),
			nullFloat64(trade.ProfitLossPercent),
			nullInt64(trade.RelatedBuyID),
		)
		if err != nil {
			return fmt.Errorf("failed to insert trade: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// InsertPosition inserts a new position into the database
func (db *DB) InsertPosition(pos *Position) (int64, error) {
	query := `
		INSERT INTO positions (
			symbol, quantity, entry_price, entry_time, exit_price,
			exit_time, strategy, is_open, profit_loss, profit_loss_percent,
			buy_trade_id, sell_trade_id
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := db.conn.Exec(
		query,
		pos.Symbol,
		pos.Quantity,
		pos.EntryPrice,
		pos.EntryTime,
		nullFloat64(pos.ExitPrice),
		pos.ExitTime,
		pos.Strategy,
		pos.IsOpen,
		nullFloat64(pos.ProfitLoss),
		nullFloat64(pos.ProfitLossPercent),
		pos.BuyTradeID,
		nullInt64(pos.SellTradeID),
	)

	if err != nil {
		return 0, fmt.Errorf("failed to insert position: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert ID: %w", err)
	}

	return id, nil
}

// UpdatePosition updates an existing position (used when closing a position)
func (db *DB) UpdatePosition(id int64, exitPrice float64, exitTime time.Time, profitLoss, profitLossPercent float64, sellTradeID int64) error {
	query := `
		UPDATE positions
		SET exit_price = ?, exit_time = ?, is_open = 0,
			profit_loss = ?, profit_loss_percent = ?, sell_trade_id = ?
		WHERE id = ?
	`

	_, err := db.conn.Exec(query, exitPrice, exitTime, profitLoss, profitLossPercent, sellTradeID, id)
	if err != nil {
		return fmt.Errorf("failed to update position: %w", err)
	}

	return nil
}

// GetOpenPosition retrieves the currently open position for a symbol
func (db *DB) GetOpenPosition(symbol string) (*Position, error) {
	query := `
		SELECT id, symbol, quantity, entry_price, entry_time, strategy, buy_trade_id
		FROM positions
		WHERE symbol = ? AND is_open = 1
		LIMIT 1
	`

	var pos Position
	err := db.conn.QueryRow(query, symbol).Scan(
		&pos.ID,
		&pos.Symbol,
		&pos.Quantity,
		&pos.EntryPrice,
		&pos.EntryTime,
		&pos.Strategy,
		&pos.BuyTradeID,
	)

	if err == sql.ErrNoRows {
		return nil, nil // No open position
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get open position: %w", err)
	}

	pos.IsOpen = true
	return &pos, nil
}

// GetRecentTrades retrieves the most recent trades
func (db *DB) GetRecentTrades(limit int) ([]Trade, error) {
	query := `
		SELECT id, symbol, side, quantity, price, total, strategy,
			   indicator_values, signal_reason, paper_trade, timestamp,
			   binance_order_id, profit_loss, profit_loss_percent, related_buy_id
		FROM trades
		ORDER BY timestamp DESC
		LIMIT ?
	`

	rows, err := db.conn.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query recent trades: %w", err)
	}
	defer rows.Close()

	var trades []Trade
	for rows.Next() {
		var t Trade
		var profitLoss, profitLossPercent sql.NullFloat64
		var relatedBuyID sql.NullInt64
		var binanceOrderID sql.NullString

		err := rows.Scan(
			&t.ID,
			&t.Symbol,
			&t.Side,
			&t.Quantity,
			&t.Price,
			&t.Total,
			&t.Strategy,
			&t.IndicatorValues,
			&t.SignalReason,
			&t.PaperTrade,
			&t.Timestamp,
			&binanceOrderID,
			&profitLoss,
			&profitLossPercent,
			&relatedBuyID,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan trade: %w", err)
		}

		if profitLoss.Valid {
			t.ProfitLoss = profitLoss.Float64
		}
		if profitLossPercent.Valid {
			t.ProfitLossPercent = profitLossPercent.Float64
		}
		if relatedBuyID.Valid {
			t.RelatedBuyID = relatedBuyID.Int64
		}
		if binanceOrderID.Valid {
			t.BinanceOrderID = binanceOrderID.String
		}

		trades = append(trades, t)
	}

	return trades, nil
}

// GetTradesByDateRange retrieves trades within a date range
func (db *DB) GetTradesByDateRange(start, end time.Time) ([]Trade, error) {
	query := `
		SELECT id, symbol, side, quantity, price, total, strategy,
			   indicator_values, signal_reason, paper_trade, timestamp,
			   binance_order_id, profit_loss, profit_loss_percent, related_buy_id
		FROM trades
		WHERE timestamp BETWEEN ? AND ?
		ORDER BY timestamp DESC
	`

	rows, err := db.conn.Query(query, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to query trades by date: %w", err)
	}
	defer rows.Close()

	var trades []Trade
	for rows.Next() {
		var t Trade
		var profitLoss, profitLossPercent sql.NullFloat64
		var relatedBuyID sql.NullInt64
		var binanceOrderID sql.NullString

		err := rows.Scan(
			&t.ID,
			&t.Symbol,
			&t.Side,
			&t.Quantity,
			&t.Price,
			&t.Total,
			&t.Strategy,
			&t.IndicatorValues,
			&t.SignalReason,
			&t.PaperTrade,
			&t.Timestamp,
			&binanceOrderID,
			&profitLoss,
			&profitLossPercent,
			&relatedBuyID,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan trade: %w", err)
		}

		if profitLoss.Valid {
			t.ProfitLoss = profitLoss.Float64
		}
		if profitLossPercent.Valid {
			t.ProfitLossPercent = profitLossPercent.Float64
		}
		if relatedBuyID.Valid {
			t.RelatedBuyID = relatedBuyID.Int64
		}
		if binanceOrderID.Valid {
			t.BinanceOrderID = binanceOrderID.String
		}

		trades = append(trades, t)
	}

	return trades, nil
}

// GetTradeSummary calculates aggregate statistics
func (db *DB) GetTradeSummary() (*TradeSummary, error) {
	query := `
		SELECT
			COUNT(*) as total_trades,
			COALESCE(SUM(CASE WHEN side = 'BUY' THEN 1 ELSE 0 END), 0) as total_buys,
			COALESCE(SUM(CASE WHEN side = 'SELL' THEN 1 ELSE 0 END), 0) as total_sells,
			COALESCE(SUM(CASE WHEN side = 'SELL' THEN profit_loss ELSE 0 END), 0) as total_profit_loss,
			COALESCE(AVG(CASE WHEN side = 'SELL' THEN profit_loss ELSE NULL END), 0) as avg_profit_loss,
			COALESCE(MAX(CASE WHEN side = 'SELL' THEN profit_loss ELSE NULL END), 0) as largest_win,
			COALESCE(MIN(CASE WHEN side = 'SELL' THEN profit_loss ELSE NULL END), 0) as largest_loss,
			MIN(timestamp) as start_date,
			MAX(timestamp) as end_date
		FROM trades
	`

	var summary TradeSummary
	var startDateStr, endDateStr sql.NullString

	err := db.conn.QueryRow(query).Scan(
		&summary.TotalTrades,
		&summary.TotalBuys,
		&summary.TotalSells,
		&summary.TotalProfitLoss,
		&summary.AverageProfitLoss,
		&summary.LargestWin,
		&summary.LargestLoss,
		&startDateStr,
		&endDateStr,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to calculate summary: %w", err)
	}

	// Parse string timestamps to time.Time
	if startDateStr.Valid {
		if t, err := time.Parse(time.RFC3339, startDateStr.String); err == nil {
			summary.StartDate = t
		}
	}
	if endDateStr.Valid {
		if t, err := time.Parse(time.RFC3339, endDateStr.String); err == nil {
			summary.EndDate = t
		}
	}

	// Calculate win rate
	if summary.TotalSells > 0 {
		winQuery := `SELECT COUNT(*) FROM trades WHERE side = 'SELL' AND profit_loss > 0`
		var wins int
		if err := db.conn.QueryRow(winQuery).Scan(&wins); err == nil {
			summary.WinRate = (float64(wins) / float64(summary.TotalSells)) * 100
		}
	}

	return &summary, nil
}

// Helper functions for NULL handling
func nullFloat64(f float64) sql.NullFloat64 {
	if f == 0 {
		return sql.NullFloat64{Valid: false}
	}
	return sql.NullFloat64{Float64: f, Valid: true}
}

func nullInt64(i int64) sql.NullInt64 {
	if i == 0 {
		return sql.NullInt64{Valid: false}
	}
	return sql.NullInt64{Int64: i, Valid: true}
}

// SerializeIndicatorValues converts a map to JSON string for storage
func SerializeIndicatorValues(values map[string]float64) string {
	data, err := json.Marshal(values)
	if err != nil {
		return "{}"
	}
	return string(data)
}

// DeserializeIndicatorValues converts JSON string back to map
func DeserializeIndicatorValues(jsonStr string) map[string]float64 {
	var values map[string]float64
	if err := json.Unmarshal([]byte(jsonStr), &values); err != nil {
		return make(map[string]float64)
	}
	return values
}

// ClearPaperTrades deletes all paper trades and their associated positions
func (db *DB) ClearPaperTrades() error {
	tx, err := db.conn.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Delete positions associated with paper trades
	_, err = tx.Exec(`
		DELETE FROM positions
		WHERE buy_trade_id IN (SELECT id FROM trades WHERE paper_trade = 1)
		OR sell_trade_id IN (SELECT id FROM trades WHERE paper_trade = 1)
	`)
	if err != nil {
		return fmt.Errorf("failed to delete paper positions: %w", err)
	}

	// Delete paper trades
	_, err = tx.Exec("DELETE FROM trades WHERE paper_trade = 1")
	if err != nil {
		return fmt.Errorf("failed to delete paper trades: %w", err)
	}

	return tx.Commit()
}
