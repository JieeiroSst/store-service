import axios, { AxiosError } from 'axios'

export const starton = axios.create({
    baseURL: 'https://api.starton.com/v3',
    headers: {
        'x-api-key': process.env.STARTON_API_KEY
    }
})

export const selectedNetwork = 'avalanche-fuji'