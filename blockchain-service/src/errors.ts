export class HttpError extends Error {
    constructor(readonly status: number, message: string, readonly details?: unknown) {
        super(message)
    }
}

export const unauthenticated = () => new HttpError(401, 'authentication required')
export const forbidden = () => new HttpError(403, 'forbidden')
export const notFound = (what: string) => new HttpError(404, `${what} not found`)
