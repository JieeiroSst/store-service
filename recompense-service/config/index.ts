interface Config {
    Port: number | undefined;
    RedisHost: string | undefined;
    Postgres: Postgres;
}

interface Postgres {
    Host: string | undefined;
    Port: number | undefined;
    User: string | undefined;
    Password: string | undefined;
    Database: string | undefined;
}

var postgres: Postgres = {
    Host: process.env.HostPostgres,
    Port: Number(process.env.PortPostgres),
    User: process.env.UserPostgres,
    Password: process.env.PasswordPostgres,
    Database: process.env.DatabasePostgres,
}

var config: Config = {
    Port: Number(process.env.PORT),
    RedisHost: process.env.RedisHost,
    Postgres: postgres,
}

export default config;