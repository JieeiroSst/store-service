const uuid = require('uuid');

GenarateID = () => {
    return uuid.v4();
}

module.exports = { GenarateID };