SELECT count(*) FROM {table}
SELECT count(*) FROM {table} WHERE eventNumber > 1
SELECT count(*) FROM {table} WHERE eventNumber > 20000
SELECT count(*) FROM {table} WHERE eventNumber >  500000
SELECT eventFile, count(*) FROM {table} GROUP BY eventFile
SELECT eventFile, count(*) FROM {table} WHERE eventNumber > 525000 GROUP BY eventFile
SELECT eventFile, eventTime, count(*) FROM {table} WHERE eventNumber > 525000 GROUP BY eventFile, eventTime ORDER BY eventFile DESC, eventTime ASC
SELECT eventFile, AVG(eventTime), AVG(multiplicity), MAX(runNumber), count(*) FROM {table} WHERE eventNumber > 20000 GROUP BY eventFile
