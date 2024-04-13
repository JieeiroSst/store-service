import { Client } from 'pg';


import config from '../../config';

const sql = new Client({
    host: config.Postgres.Host,
    port: config.Postgres.Port,
    user: config.Postgres.User,
    password: config.Postgres.Password,
    database: config.Postgres.Database,
})



export default sql