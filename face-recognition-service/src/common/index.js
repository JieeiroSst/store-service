const axios = require('axios');

const FeatchApi = async (url, token,headers) => {
    const headers = {
        "token": token,
    };
    const options = {
        method: "POST",
        url: url,
        headers: headers,
        data: form
    };
    let data = await axios(options);
    return data;
}

module.exports = {
    FeatchApi
}