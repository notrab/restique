# Restique

## Usage

```bash
./restique dev.db
```

```bash
./restique dev.db --port 5000
```

## Routes

### Get all tables

```bash
curl http://localhost:8080/
```

### Get a single table

```bash
curl http://localhost:8080/{tableName}
```

### Get a row by primary key

Doesn't work with composite primary keys.

```bash
curl http://localhost:8080/{tableName}/{primaryKeyValue}
```
