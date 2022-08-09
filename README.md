# pxier_db_syncer
`pxier_db_syncer` is the syncer for [Pxier](https://github.com/JobberRT/pxier), it syncs read db and write db. write db is master,  read db is slave. It implements the syncing function rather than using mysql's replication because of easy deployment in docker

## Configuration
`write_url`: write db's url
`read_url`: read db's url
`sync_interval`: syncing interval

## How to use
Recommend to use [Pxier](https://github.com/JobberRT/pxier) README's docker-compose file to deploy. Otherwise, you can compile and change the configuration and rename the `config.example.yaml` to `config.yaml`, then you can start the executable.