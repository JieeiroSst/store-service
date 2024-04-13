import { createClient } from 'redis';

import config from '../../config';

const redis = createClient({
    url: config.RedisHost, 
});

export default redis;