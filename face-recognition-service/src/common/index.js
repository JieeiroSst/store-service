const axios = require('axios');

const FeatchApi = async (url, method,headers) => {
    const options = {
        method: method,
        url: url,
        headers: headers,
        data: form
    };
    let data = await axios(options);
    return data;
}


module.exports = {
    FeatchApi,
}