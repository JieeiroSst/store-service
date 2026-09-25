interface Postgres {
    Host: string;
    Port: number;
    User: string;
    Password: string;
    Database: string;
}

interface Config {
    Port: number;
    RedisHost: string | undefined;
    CacheTTLSeconds: number;
    Postgres: Postgres;
}

const postgres: Postgres = {
    Host: process.env.HostPostgres ?? "localhost",
    Port: Number(process.env.PortPostgres ?? 5432),
    User: process.env.UserPostgres ?? "postgres",
    Password: process.env.PasswordPostgres ?? "",
    Database: process.env.DatabasePostgres ?? "recompense_service",
}

const config: Config = {
    Port: Number(process.env.PORT ?? 3000),
    RedisHost: process.env.RedisHost || undefined,
    CacheTTLSeconds: Number(process.env.CacheTTLSeconds ?? 60),
    Postgres: postgres,
}

export type { Config, Postgres };
export default config;
