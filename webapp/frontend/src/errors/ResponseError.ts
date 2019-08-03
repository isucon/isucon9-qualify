export class ResponseError extends Error {
    private readonly res: Response;

    constructor(message: string, response: Response) {
        super(message);
        this.res = response;
    }

    getResponse(): Response {
        return this.res;
    }
}