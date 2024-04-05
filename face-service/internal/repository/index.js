const db = require("../../pkg/postgres");
const id = require("../../pkg/uuid");

const createInfoConnect = async(object) => {
    const info = {
        id: id.GenarateID(),
    }

    return await db("infomartion").insert(info).returning("*");
}

const repositories = {
    createInfoConnect,
}

module.exports = repositories;