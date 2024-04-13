import { Worker } from 'snowflake-uuid';

const worker = new Worker(0, 1);

const nextId = worker.nextId().toString; 

export default nextId;