require("dotenv").config();

const { CLIENT, URL } = process.env;

module.exports = {
  development: {
    client: CLIENT,
    connection: URL,
    migrations: {
      directory: __dirname + "../../script/migrations",
    },
    seeds: {
      directory: __dirname + "../../script/seeds",
    },
  },
};