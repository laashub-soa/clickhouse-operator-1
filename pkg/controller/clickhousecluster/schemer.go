package clickhousecluster

import (
	"fmt"
	"github.com/MakeNowJust/heredoc"
	clickhousev1 "github.com/mackwong/clickhouse-operator/pkg/apis/clickhouse/v1"
	"github.com/mackwong/clickhouse-operator/pkg/connect"
	log "github.com/sirupsen/logrus"
	"strings"
)

const (
	// Comma-separated ''-enclosed list of database names to be ignored
	// ignoredDBs = "'system'"

	// Max number of tries for SQL queries
	defaultMaxTries = 10
)

// Schemer
type Schemer struct {
	Username string
	Password string
	Port     int
}

// NewSchemer
func NewSchemer(cc *clickhousev1.ClickHouseCluster) *Schemer {
	result := decodeUsersXML(cc.Spec.Users)
	var username, password string
	for u, p := range result {
		username = u
		password = p
	}
	return &Schemer{
		Username: username,
		Password: password,
		Port:     chDefaultHTTPPortNumber,
	}
}

// getCHConnection
func (s *Schemer) getCHConnection(hostname string) *connect.CHConnection {
	log.Infof("%s:%s", s.Username, s.Password)
	return connect.GetPooledDBConnection(connect.NewCHConnectionParams(hostname, s.Username, s.Password, s.Port))
}

// getObjectListFromClickHouse
func (s *Schemer) getObjectListFromClickHouse(endpoints []string, sql string) ([]string, []string, error) {
	if len(endpoints) == 0 {
		// Nowhere to fetch data from
		return nil, nil, nil
	}

	// Results
	var names []string
	var statements []string
	var err error

	// Fetch data from any of specified services
	var query *connect.Query = nil
	for _, endpoint := range endpoints {
		log.Infof("Run query on: %s of %v", endpoint, endpoints)

		query, err = s.getCHConnection(endpoint).Query(sql)
		if err == nil {
			// One of specified services returned result, no need to iterate more
			break
		} else {
			log.Infof("Run query on: %s of %v FAILED skip to next. err: %v", endpoint, endpoints, err)
		}
	}
	if err != nil {
		log.Infof("Run query FAILED on all %v", endpoints)
		return nil, nil, err
	}

	// Some data available, let's fetch it
	defer query.Close()

	for query.Rows.Next() {
		var name, statement string
		if err := query.Rows.Scan(&name, &statement); err == nil {
			names = append(names, name)
			statements = append(statements, statement)
		} else {
			log.Infof("UNABLE to scan row err: %v", err)
		}
	}

	return names, statements, nil
}

// getCreateDistributedObjects returns a list of objects that needs to be created on a shard in a cluster
// That includes all Distributed tables, corresponding local tables, and databases, if necessary
func (s *Schemer) getCreateDistributedObjects(hosts []string) ([]string, []string, error) {
	cluster_tables := fmt.Sprintf("remote('%s', system, tables, '%s', '%s')", strings.Join(hosts, ","), s.Username, s.Password)

	sqlDBs := heredoc.Doc(strings.ReplaceAll(`
		SELECT DISTINCT 
			database AS name, 
			concat('CREATE DATABASE IF NOT EXISTS "', name, '"') AS create_query
		FROM 
		(
			SELECT DISTINCT arrayJoin([database, extract(engine_full, 'Distributed\\([^,]+, *\'?([^,\']+)\'?, *[^,]+')]) database
			FROM cluster('all-sharded', system.tables) tables
			WHERE engine = 'Distributed'
			SETTINGS skip_unavailable_shards = 1
		)`,
		"cluster('all-sharded', system.tables)",
		cluster_tables,
	))
	sqlTables := heredoc.Doc(strings.ReplaceAll(`
		SELECT DISTINCT 
			concat(database,'.', name) as name, 
			replaceRegexpOne(create_table_query, 'CREATE (TABLE|VIEW|MATERIALIZED VIEW)', 'CREATE \\1 IF NOT EXISTS')
		FROM 
		(
			SELECT 
			    database, name,
				create_table_query,
				2 AS order
			FROM cluster('all-sharded', system.tables) tables
			WHERE engine = 'Distributed'
			SETTINGS skip_unavailable_shards = 1
			UNION ALL
			SELECT 
				extract(engine_full, 'Distributed\\([^,]+, *\'?([^,\']+)\'?, *[^,]+') AS database, 
				extract(engine_full, 'Distributed\\([^,]+, [^,]+, *\'?([^,\\\')]+)') AS name,
				t.create_table_query,
				1 AS order
			FROM cluster('all-sharded', system.tables) tables
			LEFT JOIN (SELECT distinct database, name, create_table_query 
			             FROM cluster('all-sharded', system.tables) SETTINGS skip_unavailable_shards = 1)  t USING (database, name)
			WHERE engine = 'Distributed' AND t.create_table_query != ''
			SETTINGS skip_unavailable_shards = 1
		) tables
		ORDER BY order
		`,
		"cluster('all-sharded', system.tables)",
		cluster_tables,
	))

	log.Infof("fetch dbs list")
	log.Infof("dbs sql\n%v", sqlDBs)
	names1, sqlStatements1, err := s.getObjectListFromClickHouse(hosts, sqlDBs)
	if err != nil {
		return nil, nil, err
	}
	log.Infof("names1:")
	for _, v := range names1 {
		log.Infof("names1: %s", v)
	}
	log.Infof("sql1:")
	for _, v := range sqlStatements1 {
		log.Infof("sql1: %s", v)
	}

	log.Infof("fetch table list")
	log.Infof("tbl sql\n%v", sqlTables)
	names2, sqlStatements2, _ := s.getObjectListFromClickHouse(hosts, sqlTables)
	log.Infof("names2:")
	for _, v := range names2 {
		log.Infof("names2: %s", v)
	}
	log.Infof("sql2:")
	for _, v := range sqlStatements2 {
		log.Infof("sql2: %s", v)
	}

	return append(names1, names2...), append(sqlStatements1, sqlStatements2...), nil
}

// getCreateReplicaObjects returns a list of objects that needs to be created on a host in a cluster
func (s *Schemer) getCreateReplicaObjects(hosts []string) ([]string, []string, error) {

	if len(hosts) <= 1 {
		log.Info("Single replica in a shard. Nothing to create a schema from.")
		return nil, nil, nil
	}
	// remove new replica from the list. See https://stackoverflow.com/questions/37334119/how-to-delete-an-element-from-a-slice-in-golang
	log.Infof("Extracting replicated table definitions from %v", hosts)

	system_tables := fmt.Sprintf("remote('%s', system, tables, '%s', '%s')", strings.Join(hosts, ","), s.Username, s.Password)

	sqlDBs := heredoc.Doc(strings.ReplaceAll(`
		SELECT DISTINCT 
			database AS name, 
			concat('CREATE DATABASE IF NOT EXISTS "', name, '"') AS create_db_query
		FROM system.tables
		WHERE database != 'system'
		SETTINGS skip_unavailable_shards = 1`,
		"system.tables", system_tables,
	))
	sqlTables := heredoc.Doc(strings.ReplaceAll(`
		SELECT DISTINCT 
			name, 
			replaceRegexpOne(create_table_query, 'CREATE (TABLE|VIEW|MATERIALIZED VIEW)', 'CREATE \\1 IF NOT EXISTS')
		FROM system.tables
		WHERE database != 'system' and create_table_query != '' and name not like '.inner.%'
		SETTINGS skip_unavailable_shards = 1`,
		"system.tables",
		system_tables,
	))

	names1, sqlStatements1, err := s.getObjectListFromClickHouse(hosts, sqlDBs)
	if err != nil {
		return nil, nil, err
	}
	names2, sqlStatements2, err := s.getObjectListFromClickHouse(hosts, sqlTables)
	if err != nil {
		return nil, nil, err
	}
	return append(names1, names2...), append(sqlStatements1, sqlStatements2...), nil
}

func (s *Schemer) StatefulSetCreateTables(clusterName string, hosts []string) error {

	names, createSQLs, err1 := s.getCreateReplicaObjects(hosts)
	if err1 != nil {
		log.Errorf("Get replica object err")
		return err1
	}
	log.Infof("Creating replica objects at %s: %v", clusterName, names)
	log.Infof("\n%v", createSQLs)
	err1 = s.hostApplySQLs(hosts, createSQLs, true)

	names, createSQLs, err2 := s.getCreateDistributedObjects(hosts)
	if err2 != nil {
		log.Errorf("Get distribute object err")
		return err1
	}
	log.Infof("Creating distributed objects at %s: %v", clusterName, names)
	log.Infof("\n%v", createSQLs)
	err2 = s.hostApplySQLs(hosts, createSQLs, true)

	if err2 != nil {
		return err2
	}
	if err1 != nil {
		return err1
	}

	return nil
}

// hostApplySQLs runs set of SQL queries over the replica
func (s *Schemer) hostApplySQLs(hosts []string, sqls []string, retry bool) error {
	return s.applySQLs(hosts, sqls, retry)
}

// applySQLs runs set of SQL queries on set on hosts
// Retry logic traverses the list of SQLs multiple times until all SQLs succeed
func (s *Schemer) applySQLs(hosts []string, sqls []string, retry bool) error {
	var err error = nil
	// For each host in the list run all SQL queries
	for _, host := range hosts {
		conn := s.getCHConnection(host)
		maxTries := 1
		if retry {
			maxTries = defaultMaxTries
		}
		log.Info("Start to create table for %s", host)
		err = Retry(maxTries, "Applying sqls", func() error {
			for _, sql := range sqls {
				if len(sql) == 0 {
					log.Infof("Skip malformed or already executed SQL query, move to the next one, host: %s", host)
					continue
				}
				err = conn.Exec(sql)
				if err != nil && strings.Contains(err.Error(), "Code: 253,") && strings.Contains(sql, "CREATE TABLE") {
					log.Info("Replica is already in ZooKeeper. Trying ATTACH TABLE instead")
					sqlAttach := strings.ReplaceAll(sql, "CREATE TABLE", "ATTACH TABLE")
					err = conn.Exec(sqlAttach)
				}
				if err != nil {
					return err
				}
			}
			return nil
		})
	}
	return err
}
