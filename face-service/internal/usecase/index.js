const repositories = require("../repository")

const createInfoConnect = (object) => {
    return await repositories.createInfoConnect(object)
}

const usecase = {
    createInfoConnect
}

module.exports = usecase;