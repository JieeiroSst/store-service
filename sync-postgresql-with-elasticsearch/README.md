Usage

```
docker-compose up --build

# wait until it's setup
sudo chmod +x start.sh
./start.sh
```

Testing

```
# Check contents of the PostgreSQL database:
docker-compose exec postgres bash -c 'psql -U $POSTGRES_USER $POSTGRES_DATABASE -c "SELECT * FROM users"'

# Check contents of the Elasticsearch database:
curl http://localhost:9200/users/_search?pretty
```

```
# Check contents of the PostgreSQL database:
docker-compose exec postgres bash -c 'psql -U $POSTGRES_USER $POSTGRES_DATABASE -c "SELECT * FROM users"'

# Check contents of the Elasticsearch database:
curl http://localhost:9200/users/_search?pretty

curl http://localhost:9200/users/_search?q=id:6
```