package connect

import (
	log "github.com/sirupsen/logrus"
	"sync"
)

var (
	dbConnectionPool               = sync.Map{}
	dbConnectionPoolEntryInitMutex = sync.Mutex{}
)

// GetPooledDBConnection gets connection out of the pool.
// In case no connection available new connection is created and returned.
func GetPooledDBConnection(params *CHConnectionParams) *CHConnection {
	key := makePoolKey(params)

	if connection, existed := dbConnectionPool.Load(key); existed {
		log.Infof("Found pooled connection: %s", params.GetDSNWithHiddenCredentials())
		return connection.(*CHConnection)
	}

	// Pooled connection not found, need to add it to the pool

	dbConnectionPoolEntryInitMutex.Lock()
	defer dbConnectionPoolEntryInitMutex.Unlock()

	// Double check for race condition
	if connection, existed := dbConnectionPool.Load(key); existed {
		log.Infof("Found pooled connection: %s", params.GetDSNWithHiddenCredentials())
		return connection.(*CHConnection)
	}

	log.Infof("Add connection to the pool: %s", params.GetDSNWithHiddenCredentials())
	dbConnectionPool.Store(key, NewConnection(params))

	// Fetch from the pool
	if connection, existed := dbConnectionPool.Load(key); existed {
		log.Infof("Found pooled connection: %s", params.GetDSNWithHiddenCredentials())
		return connection.(*CHConnection)
	}

	return nil
}

// makePoolKey makes key out of connection params to be used by the pool
func makePoolKey(params *CHConnectionParams) string {
	return params.GetDSN()
}
